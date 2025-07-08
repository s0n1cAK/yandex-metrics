package memstorage

import (
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_New(t *testing.T) {
	storage := New()
	require.Empty(t, storage.GetAll())
	require.NotNil(t, storage)

}

func TestMemStorage_SetGet(t *testing.T) {
	storage := New()

	type want struct {
		key     string
		value   models.Metrics
		wantErr bool
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "Valid Request",
			want: want{
				key: "TestMetric",
				value: models.Metrics{
					ID:    "TestMetric",
					MType: models.Gauge,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage.Set(test.want.key, test.want.value)

			value := storage.GetAll()["TestMetric"]
			require.Equal(t, "TestMetric", value.ID)
			require.Equal(t, models.Gauge, value.MType)
			require.NotNil(t, value.Value)
			require.InEpsilon(t, 0.41, *value.Value, 0.00001)
		})
	}
}

func TestMemStorage_Set(t *testing.T) {
	storage := New()

	type want struct {
		key     string
		value   models.Metrics
		wantErr bool
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "Valid Request",
			want: want{
				key: "TestMetric",
				value: models.Metrics{
					ID:    "TestMetric",
					MType: models.Gauge,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: false,
			},
		},
		{
			name: "Key is Empty",
			want: want{
				key: "",
				value: models.Metrics{
					ID:    "TestMetric",
					MType: models.Gauge,
					Value: lib.FloatPtr(0.41),
				},
				wantErr: true,
			},
		},
		{
			name: "Unknown type of metric",
			want: want{
				key: "TestMetric",
				value: models.Metrics{
					ID:    "TestMetric",
					MType: "somethin_else",
					Value: lib.FloatPtr(0.41),
				},
				wantErr: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Set(test.want.key, test.want.value)

			if test.want.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
