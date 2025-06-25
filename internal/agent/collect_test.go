package agent

import (
	"testing"

	agentStorage "github.com/s0n1cAK/yandex-metrics/internal/agent/storage"
	"github.com/s0n1cAK/yandex-metrics/internal/lib"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/stretchr/testify/require"
)

func TestAgent_CollectRuntime(t *testing.T) {
	storage := agentStorage.New()
	agent := &Config{
		Storage: storage,
	}

	err := agent.CollectRuntime()
	require.NoError(t, err)

	require.NotEmpty(t, storage.GetAll()[0])
	require.Equal(t, models.Gauge, storage.GetAll()[0].MType)
	require.NotEmpty(t, storage.GetAll()[0].ID)
}

func TestAgent_RandomValue(t *testing.T) {
	storage := agentStorage.New()
	agent := &Config{
		Storage: storage,
	}

	agent.RandomValue()

	require.NotEmpty(t, storage.GetAll()[0])
	require.Equal(t, "RandomValue", storage.GetAll()[0].ID)
	require.Equal(t, models.Gauge, storage.GetAll()[0].MType)
	require.NotEmpty(t, storage.GetAll()[0].ID)
}

func TestAgent_Counter(t *testing.T) {
	storage := agentStorage.New()
	agent := &Config{
		Storage: storage,
	}

	agent.Counter(1)

	require.NotEmpty(t, storage.GetAll()[0])
	require.Equal(t, "PollCount", storage.GetAll()[0].ID)
	require.Equal(t, models.Counter, storage.GetAll()[0].MType)
	require.Equal(t, lib.IntPtr(1), storage.GetAll()[0].Delta)
	require.NotEmpty(t, storage.GetAll()[0].ID)
}
