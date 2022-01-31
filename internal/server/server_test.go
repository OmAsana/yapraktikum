package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/metrics"
	"github.com/OmAsana/yapraktikum/internal/repository"
	"github.com/OmAsana/yapraktikum/internal/repository/inmemorystore"
)

func SetupRepo(t *testing.T, opts ...inmemorystore.Options) repository.MetricsRepository {
	t.Helper()
	repo, err := inmemorystore.NewInMemoryRepo(opts...)
	assert.NoError(t, err)
	return repo
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func executeTestRequest(t *testing.T, ts *httptest.Server, reqFunc func() (*http.Request, error)) (*http.Response, string) {
	t.Helper()
	req, err := reqFunc()
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)

}

func TestServer(t *testing.T) {
	srv, err := NewMetricsServer(SetupRepo(t))
	assert.NoError(t, err)
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
				uri:        "/update/unknown",
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
				srv, err := NewMetricsServer(SetupRepo(t))
				assert.NoError(t, err)
				ts := httptest.NewServer(srv)
				defer ts.Close()
				resp, body := testRequest(t, ts, http.MethodPost, test.uri, nil)
				defer resp.Body.Close()
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
				srv, err := NewMetricsServer(SetupRepo(t))
				assert.NoError(t, err)
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, counter := range test.inputCounters {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/counter/%s/%d", counter.Name, counter.Value), nil)
					defer resp.Body.Close()
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
				srv, err := NewMetricsServer(SetupRepo(t))
				assert.NoError(t, err)
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, g := range test.inputGauge {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/gauge/%s/%f", g.Name, g.Value), nil)
					defer resp.Body.Close()
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
				srv, err := NewMetricsServer(SetupRepo(t))
				assert.NoError(t, err)
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, val := range test.insertValues {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/gauge/%s/%f", test.gaugeName, val), nil)
					defer resp.Body.Close()
					require.Equal(t, http.StatusOK, resp.StatusCode, body)
				}
				resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/value/gauge/%s", test.gaugeName), nil)
				defer resp.Body.Close()
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
				srv, err := NewMetricsServer(SetupRepo(t))
				assert.NoError(t, err)
				ts := httptest.NewServer(srv)
				defer ts.Close()
				for _, val := range test.insertValues {
					resp, body := testRequest(t, ts, http.MethodPost, fmt.Sprintf("/update/counter/%s/%d", test.counterName, val), nil)
					defer resp.Body.Close()
					require.Equal(t, http.StatusOK, resp.StatusCode, body)
				}
				resp, body := testRequest(t, ts, http.MethodGet, fmt.Sprintf("/value/counter/%s", test.counterName), nil)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode, body)
				require.Equal(t, strconv.FormatInt(test.wantValue, 10), body)

			})

		}

	})

}

func TestMetricsServer_Value(t *testing.T) {
	type params struct {
		name          string
		rawMetricJSON string
		header        string
		wantCode      int
	}

	tests := []params{
		{
			name: "valid counter",
			rawMetricJSON: `{
"id": "counter1",
"type": "counter",
"delta": 22
}`,
			header:   "application/json",
			wantCode: http.StatusNotFound,
		},
		{
			name: "valid gauge",
			rawMetricJSON: `{
"id": "gauge1",
"type": "gauge",
"value": 1.02
}`,
			header:   "application/json",
			wantCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			srv, err := NewMetricsServer(SetupRepo(t))
			assert.NoError(t, err)
			ts := httptest.NewServer(srv)
			defer ts.Close()
			resp, body := executeTestRequest(t, ts, func() (*http.Request, error) {
				req, err := http.NewRequest(http.MethodPost, ts.URL+"/value/", strings.NewReader(tt.rawMetricJSON))
				if err != nil {
					return req, err
				}

				req.Header.Set("Accept", tt.header)
				return req, err
			})
			defer resp.Body.Close()

			require.Equal(t, tt.wantCode, resp.StatusCode, body)
		})

	}

}

func TestMetricsServer_Update(t *testing.T) {
	type params struct {
		name          string
		rawMetricJSON string
		header        string
		wantCode      int
	}

	tests := []params{
		{
			name: "valid counter",
			rawMetricJSON: `{
"id": "counter1",
"type": "counter",
"delta": 22
}`,
			header:   "application/json",
			wantCode: http.StatusOK,
		},
		{
			name: "invalid counter. wrong delta type",
			rawMetricJSON: `{
"id": "counter1",
"type": "counter",
"delta": 1.02
}`,
			header:   "application/json",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid metric. delta or value must be set",
			rawMetricJSON: `{
"id": "counter1",
"type": "counter"
}`,
			header:   "application/json",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "valid gauge",
			rawMetricJSON: `{
"id": "gauge1",
"type": "gauge",
"value": 1.02
}`,
			header:   "application/json",
			wantCode: http.StatusOK,
		},
		{
			name: "invalid gauge. wrong value type",
			rawMetricJSON: `{
"id": "gauge1",
"type": "gauge",
"value": "some val"
}`,
			header:   "application/json",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid json values",
			rawMetricJSON: `{
"some": true,
"random": true,
"json": true
}`,
			header:   "application/json",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid accept header",
			rawMetricJSON: `{
"some": true,
"random": true,
"json": true
}`,
			header:   "application/txt",
			wantCode: http.StatusNotImplemented,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			srv, err := NewMetricsServer(SetupRepo(t))
			assert.NoError(t, err)
			ts := httptest.NewServer(srv)
			defer ts.Close()
			resp, body := executeTestRequest(t, ts, func() (*http.Request, error) {
				req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", strings.NewReader(tt.rawMetricJSON))
				if err != nil {
					return req, err
				}

				req.Header.Set("Content-Type", tt.header)
				return req, err
			})
			defer resp.Body.Close()

			require.Equal(t, tt.wantCode, resp.StatusCode, body)
		})
	}
}
