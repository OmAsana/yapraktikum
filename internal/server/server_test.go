package server

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

func SetupRepo(t *testing.T) MetricsRepository {
	t.Helper()
	repo := NewRepositoryMock()
	return repo
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestServer(t *testing.T) {
	srv := NewMetricsServer(SetupRepo(t))
	ts := httptest.NewServer(srv)
	defer ts.Close()

}

func TestMetricsServer_UpdateCounters(t *testing.T) {

	t.Run("Allow correct params only", func(t *testing.T) {
		type params struct {
			uri        string
			wantStatus int
		}

		tests := []params{
			{
				uri:        "/update/counter/a/1",
				wantStatus: http.StatusOK,
			},
			{
				uri:        "/update/counter/a/-1",
				wantStatus: http.StatusBadRequest,
			},
			{
				uri:        "/update",
				wantStatus: http.StatusNotImplemented,
			},
			{
				uri:        "/update/counter/a",
				wantStatus: http.StatusNotFound,
			},
			{
				uri:        "/update/counter/a/1.2/c",
				wantStatus: http.StatusNotFound,
			},
		}

		for _, test := range tests {
			t.Run(test.uri, func(t *testing.T) {
				srv := NewMetricsServer(SetupRepo(t))
				ts := httptest.NewServer(srv)
				defer ts.Close()
				resp, body := testRequest(t, ts, http.MethodPost, test.uri)
				require.Equal(
					t,
					test.wantStatus,
					resp.StatusCode,
					fmt.Sprintf("response body: %s, request uri: %s", body, test.uri),
				)
			})
		}
	})

	t.Run("Insert counter", func(t *testing.T) {

		type params struct {
			name          string
			inputCounters []metrics.Counter
			wantCouter    metrics.Counter
			wantStatus    int
			wantErr       bool
		}

		tests := []params{
			{
				name: "Add multiple counter values",
				inputCounters: []metrics.Counter{
					{Name: "some_val", Value: 1},
					{Name: "some_val", Value: 2},
				},
				wantCouter: metrics.Counter{Name: "some_val", Value: 3},
				wantStatus: http.StatusOK,
			},
			{
				name: "Counter not found",
				inputCounters: []metrics.Counter{
					{Name: "some_val", Value: 1},
				},
				wantCouter: metrics.Counter{Name: "some_other_val", Value: 3},
				wantStatus: http.StatusOK,
				wantErr:    true,
			},
			{
				name: "Add negative value",
				inputCounters: []metrics.Counter{
					{Name: "some_val", Value: -20},
				},
				wantStatus: http.StatusBadRequest,
				wantErr:    true,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				srv := NewMetricsServer(SetupRepo(t))
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, counter := range test.inputCounters {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/counter/%s/%d", counter.Name, counter.Value))
					require.Equal(t, test.wantStatus, resp.StatusCode, body)

				}
				if test.wantStatus == http.StatusOK {
					got, err := srv.db.RetrieveCounter(test.wantCouter.Name)
					if test.wantErr {
						require.Error(t, err, err)
					} else {
						require.NoError(t, err)
						assert.Equal(t, test.wantCouter, got)

					}

				}

			})

		}

	})
}

func TestMetricsServer_UpdateGauge(t *testing.T) {
	t.Run("Add gauges", func(t *testing.T) {
		type params struct {
			name       string
			inputGauge []metrics.Gauge
			wantGauge  metrics.Gauge
			wantStatus int
			wantErr    bool
		}

		tests := []params{
			{
				name: "Add single gauge",
				inputGauge: []metrics.Gauge{
					{Name: "g1", Value: 2.0001},
				},
				wantGauge:  metrics.Gauge{Name: "g1", Value: 2.0001},
				wantStatus: http.StatusOK,
			},
			{
				name: "Add multiple gauge",
				inputGauge: []metrics.Gauge{
					{Name: "g1", Value: 2.0001},
					{Name: "g2", Value: 7.0},
					{Name: "g1", Value: 5.0},
				},
				wantGauge:  metrics.Gauge{Name: "g1", Value: 5.0},
				wantStatus: http.StatusOK,
			},
			{
				name: "Add NaN gauge",
				inputGauge: []metrics.Gauge{
					{Name: "g1", Value: math.NaN()},
				},
				wantStatus: http.StatusBadRequest,
			},
			{
				name: "Add inf gauge",
				inputGauge: []metrics.Gauge{
					{Name: "g1", Value: math.Inf(+1)},
				},
				wantStatus: http.StatusBadRequest,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				srv := NewMetricsServer(SetupRepo(t))
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, g := range test.inputGauge {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/gauge/%s/%f", g.Name, g.Value))
					require.Equal(t, test.wantStatus, resp.StatusCode, body)
				}
				if test.wantStatus == http.StatusOK {
					got, err := srv.db.RetrieveGauge(test.wantGauge.Name)
					if test.wantErr {
						require.Error(t, err, err)
					} else {
						require.NoError(t, err)
						assert.Equal(t, test.wantGauge, got)

					}
				}
			})
		}
	})
}

func TestGetMetric(t *testing.T) {
	t.Run("Insert and get gauges", func(t *testing.T) {
		type params struct {
			name         string
			gaugeName    string
			insertValues []float64
			wantValue    float64
		}

		tests := []params{
			{
				name:         "Insert single gauge",
				gaugeName:    "SomeName",
				insertValues: []float64{22},
				wantValue:    22,
			},
			{
				name:         "Insert multiple gauges",
				gaugeName:    "SomeName",
				insertValues: []float64{22, 44, 88, 1212},
				wantValue:    1212,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				srv := NewMetricsServer(SetupRepo(t))
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, val := range test.insertValues {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/gauge/%s/%f", test.gaugeName, val))
					require.Equal(t, http.StatusOK, resp.StatusCode, body)
				}
				resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/value/gauge/%s", test.gaugeName))
				require.Equal(t, http.StatusOK, resp.StatusCode, body)
				require.Equal(t, strconv.FormatFloat(test.wantValue, 'g', -1, 64), body)

			})

		}
	})

	t.Run("Insert and get counter", func(t *testing.T) {
		type params struct {
			name         string
			counterName  string
			insertValues []int64
			wantValue    int64
		}

		tests := []params{
			{
				name:         "Insert single counter",
				counterName:  "SomeName",
				insertValues: []int64{22},
				wantValue:    22,
			},
			{
				name:         "Insert multiple counter",
				counterName:  "SomeName",
				insertValues: []int64{22, 22},
				wantValue:    44,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				srv := NewMetricsServer(SetupRepo(t))
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, val := range test.insertValues {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/counter/%s/%d", test.counterName, val))
					require.Equal(t, http.StatusOK, resp.StatusCode, body)
				}
				resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/value/counter/%s", test.counterName))
				require.Equal(t, http.StatusOK, resp.StatusCode, body)
				require.Equal(t, strconv.FormatInt(test.wantValue, 10), body)

			})

		}

	})

}
