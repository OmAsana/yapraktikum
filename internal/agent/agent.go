package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/OmAsana/yapraktikum/internal/handlers"
	"github.com/OmAsana/yapraktikum/internal/logging"
	"github.com/OmAsana/yapraktikum/internal/metrics"
)

var defaultBaseURL *url.URL

type config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	BaseURL        *url.URL
	HashKey        string
}
type Agent struct {
	registry   *metrics.Registry
	cfg        config
	httpClient *http.Client
}

func NewDefaultAgent() *Agent {
	defaultBaseURL, _ = url.Parse(fmt.Sprintf("http://%s", DefaultAddress))
	return &Agent{registry: metrics.NewRegistry(),
		httpClient: &http.Client{},
		cfg: config{
			PollInterval:   DefaultPollInterval,
			ReportInterval: DefaultReportInterval,
			BaseURL:        defaultBaseURL,
		}}
}

func NewAgentWithBaseURL(baseURL *url.URL) *Agent {
	agent := NewDefaultAgent()
	agent.cfg.BaseURL = baseURL
	return agent
}

func NewAgentWithOptions(opts ...Option) (*Agent, error) {
	agent := NewDefaultAgent()

	for _, opt := range opts {
		err := opt(agent)
		if err != nil {
			return nil, err
		}
	}

	return agent, nil
}

func (a *Agent) logState() {
	logging.Log.S().Infof(
		"Agent started. PollInterval: %.2fs, ReportInterval: %.2fs, Report to address: %q",
		a.cfg.PollInterval.Seconds(),
		a.cfg.ReportInterval.Seconds(),
		a.cfg.BaseURL.String(),
	)
}

func (a *Agent) Server(ctx context.Context) {
	a.logState()

	pollTicker := time.NewTicker(a.cfg.PollInterval)
	reportTicker := time.NewTicker(a.cfg.ReportInterval)
	for {
		select {
		case <-pollTicker.C:
			a.registry.Collect()
		case <-reportTicker.C:
			//a.report()
			//a.reportAPIv2()
			a.reportAPIv3()
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) sendRequest(req *http.Request) error {

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.New(err.Error())
		}
		return fmt.Errorf("something went wrong: %s", string(bodyBytes))
	}

	defer resp.Body.Close()
	return nil

}

func (a *Agent) plainTextRequest(path string) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := a.cfg.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain")
	return req, nil
}

func (a *Agent) jsonRequest(path string, body io.Reader) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := a.cfg.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (a *Agent) updateGaugeURL(gauge metrics.Gauge) string {
	return fmt.Sprintf("/update/gauge/%s/%f", gauge.Name, gauge.Value)
}

func (a *Agent) updateCounterURL(counter metrics.Counter) string {
	return fmt.Sprintf("/update/counter/%s/%d", counter.Name, counter.Value)
}

func (a *Agent) report() {
	for _, gauge := range a.registry.Gauges {
		req, err := a.plainTextRequest(a.updateGaugeURL(gauge))
		if err != nil {
			logging.Log.S().Error(err)
		}
		_ = a.sendRequest(req)
	}

	for _, counter := range a.registry.Counters {
		req, err := a.plainTextRequest(a.updateCounterURL(counter))
		if err != nil {
			logging.Log.S().Error(err)
		}
		_ = a.sendRequest(req)

	}

	req, err := a.plainTextRequest(a.updateCounterURL(a.registry.PollCounter))
	if err != nil {
		logging.Log.S().Error(err)
	}
	_ = a.sendRequest(req)
}

func (a Agent) reportAPIv3() {

	var metrics []*handlers.Metrics
	for _, gauge := range a.registry.Gauges {
		metric := &handlers.Metrics{
			ID:    gauge.Name,
			MType: "gauge",
			Value: &gauge.Value,
		}
		metrics = append(metrics, metric)
	}
	for _, counter := range a.registry.Counters {
		metric := &handlers.Metrics{
			ID:    counter.Name,
			MType: "counter",
			Delta: &counter.Value,
		}
		metrics = append(metrics, metric)

	}

	metrics = append(metrics, &handlers.Metrics{
		ID:    a.registry.PollCounter.Name,
		MType: "counter",
		Delta: &a.registry.PollCounter.Value},
	)

	if a.cfg.HashKey != "" {
		for _, m := range metrics {
			err := m.HashMetric(a.cfg.HashKey)
			if err != nil {
				logging.Log.S().Error("Error hashing metric: ", err)
				return
			}

		}
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(metrics)
	if err != nil {
		logging.Log.S().Error("Error encoding metrics: ", err)
		return
	}

	req, err := a.jsonRequest("/updates/", &buf)
	if err != nil {
		logging.Log.S().Error("Error preparing request: ", err)
		return
	}
	err = a.sendRequest(req)
	if err != nil {
		logging.Log.S().Error("Could not complete request: ", err)
	}
}

func (a Agent) reportAPIv2() {

	var sendRequest = func(metricStream <-chan handlers.Metrics) {

		for {
			m, open := <-metricStream
			if !open {
				return
			}
			if a.cfg.HashKey != "" {
				err := m.HashMetric(a.cfg.HashKey)
				if err != nil {
					logging.Log.S().Error(err)
					continue
				}
			}

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(m)
			if err != nil {
				logging.Log.S().Error(err)
				continue
			}
			req, err := a.jsonRequest("/update/", &buf)
			if err != nil {
				logging.Log.S().Error(err)
				continue
			}
			_ = a.sendRequest(req)

		}
	}

	metricStream := make(chan handlers.Metrics)
	go sendRequest(metricStream)

	for _, gauge := range a.registry.Gauges {
		metric := handlers.Metrics{
			ID:    gauge.Name,
			MType: "gauge",
			Value: &gauge.Value,
		}
		metricStream <- metric

	}
	for _, counter := range a.registry.Counters {
		metric := handlers.Metrics{
			ID:    counter.Name,
			MType: "counter",
			Delta: &counter.Value,
		}
		metricStream <- metric

	}

	metricStream <- handlers.Metrics{
		ID:    a.registry.PollCounter.Name,
		MType: "counter",
		Delta: &a.registry.PollCounter.Value,
	}

	close(metricStream)

}
