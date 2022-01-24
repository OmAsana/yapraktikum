package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/handlers"
	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/pkg"
	"github.com/OmAsana/yapraktikum/internal/server"
)

func ExampleAgent_Server() {
	ctx := context.Background()
	agent := NewDefaultAgent()
	agent.Server(ctx)
}

func SetupRepo(t *testing.T) server.MetricsRepository {
	t.Helper()
	repo := server.NewRepositoryMock()
	return repo
}

func TestNewAgent(t *testing.T) {
	handler, err := server.NewMetricsServer(SetupRepo(t))
	assert.NoError(t, err)
	metricServer := httptest.NewServer(handler)
	baseURL, _ := url.Parse(metricServer.URL)
	agent := NewAgentWithBaseURL(baseURL)
	agent.httpClient = metricServer.Client()
	defer metricServer.Close()

	t.Run("api v1", func(t *testing.T) {
		t.Run("Add gauge", func(t *testing.T) {
			testGauge := metrics.Gauge{
				Name:  "Test",
				Value: 1.2,
			}
			req, err := agent.plainTextRequest(agent.updateGaugeURL(testGauge))
			require.NoError(t, err)

			err = agent.sendRequest(req)
			require.NoError(t, err)
		})

		t.Run("Add counter", func(t *testing.T) {
			testCounter := metrics.Counter{
				Name:  "testCounter",
				Value: 1,
			}
			req, err := agent.plainTextRequest(agent.updateCounterURL(testCounter))
			require.NoError(t, err)

			err = agent.sendRequest(req)
			require.NoError(t, err)
		})

		t.Run("Add wrong metric", func(t *testing.T) {

			req, err := agent.plainTextRequest("some/random/val")
			require.NoError(t, err)
			err = agent.sendRequest(req)
			require.Error(t, err)
		})

	})
	t.Run("JSON api", func(t *testing.T) {
		var prepJSONRequest = func(t *testing.T, metric handlers.Metrics) (*http.Request, error) {
			t.Helper()
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(metric)
			if err != nil {
				return nil, err
			}
			return agent.jsonRequest("/update/", &buf)

		}
		t.Run("Add gauge", func(t *testing.T) {

			metric := handlers.Metrics{
				ID:    "Test",
				MType: "gauge",
				Value: pkg.PointerFloat(1.2),
			}

			req, err := prepJSONRequest(t, metric)
			require.NoError(t, err)

			err = agent.sendRequest(req)
			require.NoError(t, err)
		})

		t.Run("Add counter", func(t *testing.T) {

			metric := handlers.Metrics{
				ID:    "Test",
				MType: "counter",
				Delta: pkg.PointerInt(10),
			}

			req, err := prepJSONRequest(t, metric)
			require.NoError(t, err)

			err = agent.sendRequest(req)
			require.NoError(t, err)
		})

	})
}

func TestNewAgentWithOptions(t *testing.T) {
	t.Run("check options are applied", func(t *testing.T) {
		newAddress := "127.0.0.1:1234"
		newPollInterval := "2s"
		newReportInterval := "5s"
		pkg.SetEnv(t, "ADDRESS", newAddress)
		pkg.SetEnv(t, "POLL_INTERVAL", newPollInterval)
		pkg.SetEnv(t, "REPORT_INTERVAL", newReportInterval)

		cfg := DefaultConfig
		err := cfg.initEnvArgs()
		assert.NoError(t, err)

		agent, err := NewAgentWithOptions(WithAddress(cfg.Address),
			WithPollInterval(cfg.PollInterval), WithReportInterval(cfg.ReportInterval))

		require.NoError(t, err)

		assert.Equal(t, agent.cfg.BaseURL.String(), "http://"+newAddress)
		assert.Equal(t, agent.cfg.PollInterval, func() interface{} {
			t, _ := time.ParseDuration(newPollInterval)
			return t
		}())
		assert.Equal(t, agent.cfg.ReportInterval, func() interface{} {
			t, _ := time.ParseDuration(newReportInterval)
			return t
		}())
	})
}
