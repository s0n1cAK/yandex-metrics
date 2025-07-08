package agent

import (
	"testing"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"github.com/stretchr/testify/require"
)

func TestAgent_CollectRuntime(t *testing.T) {
	storage := memstorage.New()
	agent := &Agent{
		Storage: storage,
	}

	err := agent.CollectRuntime()
	require.NoError(t, err)
	require.NotEmpty(t, storage.GetAll())
}

func TestAgent_RandomValue(t *testing.T) {
	storage := memstorage.New()
	agent := &Agent{
		Storage: storage,
	}

	err := agent.CollectRandomValue()
	require.NoError(t, err)
	require.NotEmpty(t, storage.GetAll())

	metrics := storage.GetAll()
	for _, metric := range metrics {
		require.Equal(t, "RandomValue", metric.ID)
		require.Equal(t, models.Gauge, metric.MType)
	}

}

func TestAgent_Counter(t *testing.T) {
	storage := memstorage.New()
	agent := &Agent{
		Storage: storage,
	}

	err := agent.CollectIncrementCounter("PollCount", 1)
	require.NoError(t, err)

	metric, ok := storage.Get("PollCount")
	require.Equal(t, ok, true)
	require.NotEmpty(t, storage.GetAll())
	require.Equal(t, models.Counter, metric.MType)
	require.Equal(t, int64(1), *metric.Delta)

	err = agent.CollectIncrementCounter("PollCount", 1)
	require.NoError(t, err)

	metric, ok = storage.Get("PollCount")
	require.Equal(t, ok, true)
	require.Equal(t, int64(2), *metric.Delta)
}
