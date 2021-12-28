package agent

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/metrics"
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

	t.Run("Add gauge", func(t *testing.T) {
		testGauge := metrics.Gauge{
			Name:  "Test",
			Value: 1.2,
		}
		err := agent.sendRequest(agent.updateGaugeURL(testGauge))
		require.NoError(t, err)
	})

	t.Run("Add counter", func(t *testing.T) {
		testCounter := metrics.Counter{
			Name:  "testCounter",
			Value: 1,
		}
		err := agent.sendRequest(agent.updateCounterURL(testCounter))
		require.NoError(t, err)
	})

	t.Run("Add wrong metric", func(t *testing.T) {
		err := agent.sendRequest("some/random/val")
		require.Error(t, err)
	})
}
