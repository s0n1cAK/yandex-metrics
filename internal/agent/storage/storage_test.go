package storage

import (
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/stretchr/testify/require"
)

func TestAgentStorage_New(t *testing.T) {
	storage := New()
	require.Empty(t, storage.GetAll())
	require.NotNil(t, storage)

}

func TestAgentStorage_SetGet(t *testing.T) {
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
			storage.Set(test.want.value)
			require.Contains(t, storage.GetAll(), test.want.value)
		})
	}
}

func TestAgentStorage_Clear(t *testing.T) {
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
			storage.Set(test.want.value)
			storage.Clear()
			require.NotContains(t, storage.GetAll(), test.want.value)
		})
	}
}
