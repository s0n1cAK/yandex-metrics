package agent

// import (
// 	"net/http"
// 	"testing"
// 	"time"

// 	"github.com/s0n1cAK/yandex-metrics/internal/config"
// 	"github.com/s0n1cAK/yandex-metrics/internal/logger"
// 	memstorage "github.com/s0n1cAK/yandex-metrics/internal/storage/memStorage"
// 	"github.com/stretchr/testify/require"
// )

// func TestAgent_New(t *testing.T) {
// 	logger, err := logger.NewLogger()
// 	agent := New(config.AgentConfig{
// 		Client:   &http.Client{},
// 		Endpoint: "localhost:8080",
// 		Logger:   logger,
// 	}, memstorage.New())
// 	require.NotNil(t, agent)
// 	require.NoError(t, err)
// }

// func TestAgent_Run(t *testing.T) {
// 	var agent Agent
// 	tests := []struct {
// 		name                         string
// 		pollInterval, reportInterval time.Duration
// 		wantErr                      bool
// 	}{
// 		{
// 			name:           "pollInterval == 0",
// 			pollInterval:   time.Second * 0,
// 			reportInterval: time.Second * 10,
// 			wantErr:        true,
// 		},
// 		{
// 			name:           "reportInterval > 501",
// 			pollInterval:   time.Second * 0,
// 			reportInterval: time.Second * 501,
// 			wantErr:        true,
// 		},
// 		{
// 			name:           "pollInterval > reportInterval",
// 			pollInterval:   time.Second * 20,
// 			reportInterval: time.Second * 10,
// 			wantErr:        true,
// 		},
// 	}
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			err := agent.Run(test.pollInterval, test.reportInterval)
// 			if test.wantErr == true {
// 				require.Error(t, err)
// 			}
// 		})
// 	}
// }
