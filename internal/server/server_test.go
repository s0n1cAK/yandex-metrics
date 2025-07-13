package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/config"
	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	memStorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"github.com/stretchr/testify/require"
)

func TestServerValidation_New(t *testing.T) {
	storage := memStorage.New()

	type want struct {
		sAddr   string
		sPort   int
		Storage *memStorage.MemStorage
		wantErr bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Valid Test (IP)",
			want: want{
				sAddr:   "127.0.0.1",
				sPort:   8080,
				Storage: storage,
				wantErr: false,
			},
		},
		{
			name: "Valid Test (DNS)",
			want: want{
				sAddr:   "localhost",
				sPort:   8008,
				Storage: storage,
				wantErr: false,
			},
		},
		{
			name: "Invalid Port",
			want: want{
				sAddr:   "localhost",
				sPort:   80080,
				Storage: storage,
				wantErr: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(&config.ServerConfig{
				Address: test.want.sAddr,
				Port:    test.want.sPort,
			}, test.want.Storage)

			if test.want.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerRoutes_SetMetric(t *testing.T) {
	storage := memStorage.New()

	type want struct {
		sAddr   string
		sPort   int
		storage *memStorage.MemStorage
		metric  models.Metrics
		wantErr bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Valid Request",
			want: want{
				sAddr:   "localhost",
				sPort:   8080,
				storage: storage,
				metric: models.Metrics{
					ID:    "Test",
					MType: models.Gauge,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: false,
			},
		},
		{
			name: "Invalid Request (Counter as Value)",
			want: want{
				sAddr:   "localhost",
				sPort:   8080,
				storage: storage,
				metric: models.Metrics{
					ID:    "Test",
					MType: models.Counter,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: true,
			},
		},
		{
			name: "Invalid Request (Gauge as Delta)",
			want: want{
				sAddr:   "localhost",
				sPort:   8080,
				storage: storage,
				metric: models.Metrics{
					ID:    "Test",
					MType: models.Gauge,
					Delta: lib.IntPtr(33),
				},
				wantErr: true,
			},
		},
		{
			name: "Invalid Request (Value and Delta is Empty)",
			want: want{
				sAddr:   "localhost",
				sPort:   8080,
				storage: storage,
				metric: models.Metrics{
					ID:    "Test",
					MType: models.Counter,
				},
				wantErr: true,
			},
		},
		{
			name: "Invalid Request (Unknown type of metric)",
			want: want{
				sAddr:   "localhost",
				sPort:   8080,
				storage: storage,
				metric: models.Metrics{
					ID:    "Test",
					MType: "Unknown",
				},
				wantErr: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv, err := New(&config.ServerConfig{
				Address: test.want.sAddr,
				Port:    test.want.sPort,
			},
				test.want.storage)
			require.NoError(t, err)

			var req *http.Request
			var url string

			switch test.want.metric.MType {
			case models.Gauge:
				valueStr := "null"
				if test.want.metric.Value != nil {
					valueStr = strconv.FormatFloat(*test.want.metric.Value, 'f', -1, 64)
				}
				url = path.Join("/update", test.want.metric.MType, test.want.metric.ID, valueStr)

			case models.Counter:
				deltaStr := "null"
				if test.want.metric.Delta != nil {
					deltaStr = strconv.FormatInt(int64(*test.want.metric.Delta), 10)
				}

				url = path.Join("/update", test.want.metric.MType, test.want.metric.ID, deltaStr)

			default:
				url = path.Join("/update", test.want.metric.MType, test.want.metric.ID, "unknown")
			}

			req = httptest.NewRequest(http.MethodPost, url, nil)
			w := httptest.NewRecorder()
			if req == nil {
				t.Fatal("request not formed due to invalid metric setup")
			}
			srv.Router.ServeHTTP(w, req)

			if test.want.wantErr {
				require.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				require.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestServerRoutes_GetMetric(t *testing.T) {

	type want struct {
		metric  models.Metrics
		wantErr bool
	}
	tests := []struct {
		name  string
		want  want
		store models.Metrics
	}{
		{
			name: "Valid Request",
			want: want{
				metric: models.Metrics{
					ID:    "Test",
					MType: models.Gauge,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: false,
			},
			store: models.Metrics{
				ID:    "Test",
				MType: models.Gauge,
				Value: lib.FloatPtr(0.41),
			},
		},
		{
			name: "Invalid Request (Ask Counter get Gauge)",
			want: want{
				metric: models.Metrics{
					ID:    "Test1",
					MType: models.Counter,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: true,
			},
			store: models.Metrics{
				ID:    "Test1",
				MType: models.Gauge,
				Delta: lib.IntPtr(41),
			},
		},
		{
			name: "Invalid Request (Ask Gauge get Counter)",
			want: want{
				metric: models.Metrics{
					ID:    "Test",
					MType: models.Gauge,
					Delta: lib.IntPtr(33),
				},
				wantErr: true,
			},
			store: models.Metrics{
				ID:    "Test",
				MType: models.Counter,
				Value: lib.FloatPtr(41),
			},
		},
		{
			name: "Invalid Request (Empty MType)",
			want: want{
				metric: models.Metrics{
					ID:    "Test",
					MType: "",
				},
				wantErr: true,
			},
			store: models.Metrics{
				ID:    "Test",
				MType: models.Counter,
				Delta: lib.IntPtr(41),
			},
		},
		{
			name: "Invalid Request (Unknown type of metric)",
			want: want{
				metric: models.Metrics{
					ID:    "Test",
					MType: "Unknown",
				},
				wantErr: true,
			},
			store: models.Metrics{
				ID:    "Test",
				MType: models.Counter,
				Delta: lib.IntPtr(41),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := memStorage.New()
			srv, err := New(&config.ServerConfig{
				Address: "localhost",
				Port:    8080,
			},
				storage)
			require.NoError(t, err)

			storage.Set(test.store.ID, test.store)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/%s", test.want.metric.MType, test.want.metric.ID), nil)
			w := httptest.NewRecorder()

			srv.Router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			response := string(body[:])
			if test.want.wantErr {
				require.NotEqual(t, http.StatusOK, w.Code)
				return
			}

			require.Equal(t, http.StatusOK, w.Code)

			switch test.want.metric.MType {
			case models.Gauge:
				value, _ := strconv.ParseFloat(response, 64)
				require.NotNil(t, body)
				require.InEpsilon(t, *test.want.metric.Value, value, 0.00001)
			case models.Counter:
				value, _ := strconv.ParseInt(response, 10, 64)
				require.NotNil(t, body)
				require.Equal(t, *test.want.metric.Delta, value)
			default:
				t.Fatalf("Unknown Type: %s", test.want.metric.MType)
			}
		})
	}
}

func TestServerRoutes_GetMetrics(t *testing.T) {
	s := memStorage.New()
	srv, err := New(&config.ServerConfig{
		Address: "localhost",
		Port:    8080,
	},
		s)
	require.NoError(t, err)

	testMetric := models.Metrics{
		ID:    "Test",
		MType: models.Gauge,
		Value: lib.FloatPtr(0.41),
	}

	s.Set("Test", testMetric)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	require.Equal(t, http.StatusOK, w.Code)
	srv.Router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NotNil(t, body)
}
