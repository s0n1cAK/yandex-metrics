package agent

import (
	"testing"

	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"github.com/stretchr/testify/require"
)

func TestAgent_CollectRuntime(t *testing.T) {
	storage := memstorage.New()
	agent := &Config{
		Storage: storage,
	}

	err := agent.CollectRuntime()
	require.NoError(t, err)

	// require.NotEmpty(t, storage.GetAll())
	// require.Equal(t, models.Gauge, storage.GetAll())
	// require.NotEmpty(t, storage.GetAll().ID)
}

func TestAgent_RandomValue(t *testing.T) {
	storage := memstorage.New()
	agent := &Config{
		Storage: storage,
	}

	agent.RandomValue()

	require.NotEmpty(t, storage.GetAll())
	// require.Equal(t, "RandomValue", storage.GetAll())
	// require.Equal(t, models.Gauge, storage.GetAll().MType)
	// require.NotEmpty(t, storage.GetAll()[0].ID)
}

func TestAgent_Counter(t *testing.T) {
	storage := memstorage.New()
	agent := &Config{
		Storage: storage,
	}

	agent.IncrementCounter("PollCount", 1)

	// require.NotEmpty(t, storage.GetAll()[0])
	// require.Equal(t, "PollCount", storage.GetAll()[0].ID)
	// require.Equal(t, models.Counter, storage.GetAll()[0].MType)
	// require.Equal(t, lib.IntPtr(1), storage.GetAll()[0].Delta)
	// require.NotEmpty(t, storage.GetAll()[0].ID)
}
