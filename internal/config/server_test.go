package config

import (
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/s0n1cAK/yandex-metrics/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestNewServerConfigDefault(t *testing.T) {
	_ = os.Unsetenv("ADDRESS")
	_ = os.Unsetenv("STORE_INTERVAL")
	_ = os.Unsetenv("FILE_STORAGE_PATH")
	_ = os.Unsetenv("RESTORE")

	logger, err := logger.NewLogger()
	require.NoError(t, err)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := NewServerConfigWithFlags(fs, []string{}, logger)
	require.NoError(t, err)

	require.Equal(t, Endpoint(defaultServerEndpoint), cfg.Endpoint)
	require.Equal(t, customTime(defaultStoreInterval), cfg.StoreInterval)
	require.Equal(t, defaultFile, cfg.File)
	require.Equal(t, defaultRestore, cfg.Restore)
}

func TestNewServerConfigENV(t *testing.T) {
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
				"ADDRESS":           "http://127.0.0.1:8080",
				"STORE_INTERVAL":    "30",
				"FILE_STORAGE_PATH": "./SomeWhere",
				"RESTORE":           "False",
			},
			wantErr: false,
		},
		{
			name: "Restore with error syntax",
			cfg: map[string]string{
				"RESTORE": "Fattlse",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = os.Unsetenv("ADDRESS")
			_ = os.Unsetenv("STORE_INTERVAL")
			_ = os.Unsetenv("FILE_STORAGE_PATH")
			_ = os.Unsetenv("RESTORE")

			for envName, envValue := range test.cfg {
				err = os.Setenv(envName, envValue)
				require.NoError(t, err)
			}

			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			cfg, err := NewServerConfigWithFlags(fs, []string{}, logger)
			if test.wantErr == true {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if _, ok := test.cfg["ADDRESS"]; ok {
				if cfg.Endpoint != Endpoint(defaultServerEndpoint) {
					if val, ok := test.wantValue["ADDRESS"]; ok {
						require.Equal(t, cfg.Endpoint, Endpoint(val))
					} else {
						require.Equal(t, cfg.Endpoint, Endpoint(test.cfg["ADDRESS"]))
					}
				}
			}

			if _, ok := test.cfg["STORE_INTERVAL"]; ok {
				if cfg.StoreInterval != customTime(defaultStoreInterval) {
					timeTest, err := formatCustomTime(test.cfg["STORE_INTERVAL"])
					require.NoError(t, err)
					require.Equal(t, cfg.StoreInterval, timeTest)
				}
			}

			if _, ok := test.cfg["FILE_STORAGE_PATH"]; ok {
				if cfg.File != defaultFile {
					if val, ok := test.wantValue["FILE_STORAGE_PATH"]; ok {
						require.Equal(t, cfg.File, val)
					} else {
						require.Equal(t, cfg.File, test.cfg["FILE_STORAGE_PATH"])
					}
				}
			}

			if _, ok := test.cfg["RESTORE"]; ok {
				if cfg.Restore != defaultRestore {

					if val, ok := test.wantValue["RESTORE"]; ok {
						value, _ := strconv.ParseBool(val)
						require.Equal(t, cfg.Restore, value)
					} else {
						value, _ := strconv.ParseBool(test.cfg["RESTORE"])
						require.Equal(t, cfg.Restore, value)
					}
				}
			}
		})
	}
}
