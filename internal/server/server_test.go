package server

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

func setupRepo(t *testing.T) MetricsRepository {
	t.Helper()
	repo := NewRepositoryMock()
	return repo
}

func TestMetricsServer_UpdateCounters(t *testing.T) {

	t.Run("Allow only POST http method", func(t *testing.T) {
		// all except http.MethodGet
		wrongMethod := []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}

		srv := NewMetricsServer(setupRepo(t))

		for _, m := range wrongMethod {
			request := httptest.NewRequest(m, "/", nil)

			w := httptest.NewRecorder()
			h := srv.UpdateCounters()
			h.ServeHTTP(w, request)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		}
	})

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
				wantStatus: http.StatusBadRequest,
			},
			{
				uri:        "/update/counter/a",
				wantStatus: http.StatusBadRequest,
			},
			{
				uri:        "/update/counter/a/1.2/c",
				wantStatus: http.StatusBadRequest,
			},
		}

		for _, test := range tests {
			t.Run(test.uri, func(t *testing.T) {
				srv := NewMetricsServer(setupRepo(t))
				request := httptest.NewRequest(http.MethodPost, test.uri, nil)
				w := httptest.NewRecorder()
				h := srv.UpdateCounters()
				h.ServeHTTP(w, request)
				resp := w.Result()
				defer resp.Body.Close()

				bodyBytes, err := io.ReadAll(resp.Body)

				require.NoError(t, err)
				require.Equal(t, test.wantStatus, resp.StatusCode, string(bodyBytes))
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
				srv := NewMetricsServer(setupRepo(t))
				for _, counter := range test.inputCounters {
					request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/counter/%s/%d", counter.Name, counter.Value), nil)
					w := httptest.NewRecorder()
					h := srv.UpdateCounters()
					h.ServeHTTP(w, request)
					resp := w.Result()

					bodyBytes, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					defer resp.Body.Close()
					require.Equal(t, test.wantStatus, resp.StatusCode, string(bodyBytes))

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

func Test_floatIsNumber(t *testing.T) {
	type args struct {
		float    string
		isNumber bool
	}
	tests := []args{
		{
			"NaN",
			false,
		},
		{
			"inf",
			false,
		},
		{
			"-inf",
			false,
		},
		{
			"+inf",
			false,
		},
		{
			"-0",
			true,
		},
		{
			"0",
			true,
		},
		{
			"1.2",
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test float %q", test.float), func(t *testing.T) {
			val, err := strconv.ParseFloat(test.float, 64)
			require.NoError(t, err)
			require.Equal(t, test.isNumber, floatIsNumber(val))

		})

	}
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

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				srv := NewMetricsServer(setupRepo(t))
				for _, g := range tt.inputGauge {
					request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/gauge/%s/%f", g.Name, g.Value), nil)
					w := httptest.NewRecorder()
					h := srv.UpdateGauge()
					h.ServeHTTP(w, request)
					resp := w.Result()
					bodyBytes, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					defer resp.Body.Close()
					require.Equal(t, tt.wantStatus, resp.StatusCode, string(bodyBytes))
				}
				if tt.wantStatus == http.StatusOK {
					got, err := srv.db.RetrieveGauge(tt.wantGauge.Name)
					if tt.wantErr {
						require.Error(t, err, err)
					} else {
						require.NoError(t, err)
						assert.Equal(t, tt.wantGauge, got)

					}
				}
			})
		}
	})
}
