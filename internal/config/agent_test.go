package config

import (
	"flag"
	"os"
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestNewAgentConfigDefault(t *testing.T) {
	_ = os.Unsetenv("ADDRESS")
	_ = os.Unsetenv("REPORT_INTERVAL")
	_ = os.Unsetenv("POLL_INTERVAL")

	logger, err := logger.NewLogger()
	require.NoError(t, err)

	cfg, err := NewAgentConfig(logger)
	require.NoError(t, err)

	require.Equal(t, cfg.Endpoint, Endpoint(defaultAgentEndpoint))
	require.Equal(t, cfg.ReportTime, customTime(defaultReportTime))
	require.Equal(t, cfg.PollTime, customTime(defaultPollTime))

}

func TestNewAgentConfigENV(t *testing.T) {
	logger, err := logger.NewLogger()
	require.NoError(t, err)

	tests := []struct {
		name      string
		cfg       map[string]string
		wantValue map[string]string
		wantErr   bool
	}{
		{
			name: "Set IP like Address",
			cfg: map[string]string{
				"ADDRESS":         "http://127.0.0.1:8080",
				"REPORT_INTERVAL": "30",
				"POLL_INTERVAL":   "15",
			},
			wantErr: false,
		},
		{
			name: "Set IP like Address without scheme",
			cfg: map[string]string{
				"ADDRESS": "127.0.0.1:8080",
			},
			wantValue: map[string]string{
				"ADDRESS": "http://127.0.0.1:8080",
			},
			wantErr: false,
		},
		{
			name: "Set ReportTime",
			cfg: map[string]string{
				"REPORT_INTERVAL": "300s",
			},
			wantErr: false,
		},
		{
			name: "Set PollTime",
			cfg: map[string]string{
				"POLL_INTERVAL": "300s",
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = os.Unsetenv("ADDRESS")
			_ = os.Unsetenv("REPORT_INTERVAL")
			_ = os.Unsetenv("POLL_INTERVAL")

			for envName, envValue := range test.cfg {
				err = os.Setenv(envName, envValue)
				require.NoError(t, err)
			}

			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			cfg, err := NewAgentConfigWithFlags(fs, []string{}, logger)
			if test.wantErr == true {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if _, ok := test.cfg["ADDRESS"]; ok {
				if cfg.Endpoint != Endpoint(defaultAgentEndpoint) {
					if val, ok := test.wantValue["ADDRESS"]; ok {
						require.Equal(t, cfg.Endpoint, Endpoint(val))
					} else {
						require.Equal(t, cfg.Endpoint, Endpoint(test.cfg["ADDRESS"]))
					}
				}
			}

			if _, ok := test.cfg["REPORT_INTERVAL"]; ok {
				if cfg.ReportTime != customTime(defaultReportTime) {
					timeTest, err := formatCustomTime(test.cfg["REPORT_INTERVAL"])
					require.NoError(t, err)
					require.Equal(t, cfg.ReportTime, timeTest)
				}
			}
			if _, ok := test.cfg["POLL_INTERVAL"]; ok {
				if cfg.PollTime != customTime(defaultPollTime) {
					timeTest, err := formatCustomTime(test.cfg["POLL_INTERVAL"])
					require.NoError(t, err)
					require.Equal(t, cfg.PollTime, timeTest)
				}
			}
		})
	}

}
