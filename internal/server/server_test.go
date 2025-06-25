package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	memStorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"github.com/stretchr/testify/require"
)

func TestServerValidation_New(t *testing.T) {
	storage := memStorage.New()

	type want struct {
		SAddr   string
		SPort   int
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
				SAddr:   "127.0.0.1",
				SPort:   8080,
				Storage: storage,
				wantErr: false,
			},
		},
		{
			name: "Valid Test (DNS)",
			want: want{
				SAddr:   "localhost",
				SPort:   8008,
				Storage: storage,
				wantErr: false,
			},
		},
		{
			name: "Invalid Port",
			want: want{
				SAddr:   "localhost",
				SPort:   80080,
				Storage: storage,
				wantErr: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.want.SAddr, test.want.SPort, test.want.Storage)

			if test.want.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerRoutes_Set(t *testing.T) {
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
					MType: models.Counter,
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
			srv, err := New(test.want.sAddr, test.want.sPort, test.want.storage)
			require.NoError(t, err)

			var req *http.Request
			var url string

			switch test.want.metric.MType {
			case models.Gauge:
				valueStr := "null"
				if test.want.metric.Value != nil {
					valueStr = strconv.FormatFloat(*test.want.metric.Value, 'f', -1, 64)
				}
				url = fmt.Sprintf("/update/%s/%s/%s", test.want.metric.MType, test.want.metric.ID, valueStr)

			case models.Counter:
				deltaStr := "null"
				if test.want.metric.Delta != nil {
					deltaStr = strconv.FormatInt(int64(*test.want.metric.Delta), 10)
				}
				url = fmt.Sprintf("/update/%s/%s/%s", test.want.metric.MType, test.want.metric.ID, deltaStr)

			default:
				url = fmt.Sprintf("/update/%s/%s/%s", test.want.metric.MType, test.want.metric.ID, "unknown")
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
