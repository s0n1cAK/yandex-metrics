package agent

import (
	"net/http"
	"testing"
	"time"

	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
	"github.com/stretchr/testify/require"
)

func TestAgent_New(t *testing.T) {
	agent := New(&http.Client{}, "localhost:8080", memstorage.New(), time.Now())
	require.NotNil(t, agent)
}
