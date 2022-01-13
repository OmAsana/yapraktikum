package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
	handler := server.NewMetricsServer(SetupRepo(t))
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
