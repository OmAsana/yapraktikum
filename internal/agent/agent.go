package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

var defaultBaseURL *url.URL

func init() {
	defaultBaseURL, _ = url.Parse("http://127.0.0.1:8080")
}

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	BaseURL        *url.URL
}
type Agent struct {
	registry   *metrics.Registry
	cfg        Config
	httpClient *http.Client
}

func NewDefaultAgent() *Agent {
	return &Agent{registry: metrics.NewRegistry(),
		httpClient: &http.Client{},
		cfg: Config{
			PollInterval:   time.Second * 2,
			ReportInterval: time.Second * 10,
			BaseURL:        defaultBaseURL,
		}}
}

func NewAgentWithBaseURL(baseURL *url.URL) *Agent {
	agent := NewDefaultAgent()
	agent.cfg.BaseURL = baseURL
	return agent
}

func (a *Agent) Server(ctx context.Context) {
	pollTicker := time.NewTicker(a.cfg.PollInterval)
	reportTicker := time.NewTicker(a.cfg.ReportInterval)
	for {
		select {
		case <-pollTicker.C:
			a.registry.Collect()
		case <-reportTicker.C:
			a.report()
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) sendRequest(path string) error {
	rel := &url.URL{Path: path}
	u := a.cfg.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.New(err.Error())
		}
		return errors.New(fmt.Sprintf("something went wrong: %s", string(bodyBytes)))
	}

	defer resp.Body.Close()
	return nil

}
func (a *Agent) updateGaugeURL(gauge metrics.Gauge) string {
	return fmt.Sprintf("/update/gauge/%s/%f", gauge.Name, gauge.Value)
}

func (a *Agent) updateCounterURL(counter metrics.Counter) string {
	return fmt.Sprintf("/update/counter/%s/%d", counter.Name, counter.Value)
}

func (a *Agent) report() {
	for _, gauge := range a.registry.Gauges {
		updateURL := a.updateGaugeURL(gauge)
		_ = a.sendRequest(updateURL)
	}

	for _, counter := range a.registry.Counters {
		updateURL := a.updateCounterURL(counter)
		_ = a.sendRequest(updateURL)

	}

	_ = a.sendRequest(a.updateCounterURL(a.registry.PollCounter))
}
