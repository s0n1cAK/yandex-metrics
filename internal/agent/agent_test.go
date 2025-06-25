package agent

import (
	"net/http"
	"testing"
	"time"

	agentStorage "github.com/s0n1cAK/yandex-metrics/internal/agent/storage"
	"github.com/stretchr/testify/require"
)

func TestAgent_New(t *testing.T) {
	agent := New(&http.Client{}, "localhost:8080", agentStorage.New(), time.Now())
	require.NotNil(t, agent)
}
