package agent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	PushEndpoint   string
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
			PushEndpoint:   "http://127.0.0.1:8080",
		}}
}

func (a *Agent) Server(ctx context.Context) {
	pollTicker := time.NewTicker(a.cfg.PollInterval)
	reportTicker := time.NewTicker(a.cfg.ReportInterval)
	for {
		select {
		case <-pollTicker.C:
			fmt.Println("poll")
			a.registry.Collect()
		case <-reportTicker.C:
			a.report()
			fmt.Println("report")
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) report() {
	for _, gauge := range a.registry.Gauges {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", a.cfg.PushEndpoint, gauge.Name, gauge.Value)
		_, err := a.httpClient.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, counter := range a.registry.Counters {
		url := fmt.Sprintf("%s/update/counter/%s/%d", a.cfg.PushEndpoint, counter.Name, counter.Value)
		_, err := a.httpClient.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}

	}

	url := fmt.Sprintf("%s/update/counter/%s/%d", a.cfg.PushEndpoint, a.registry.PollCounter.Name, a.registry.PollCounter.Value)
	_, err := a.httpClient.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println(err)
	}

}
