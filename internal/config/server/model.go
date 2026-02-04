package server

import (
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint      customtype.Endpoint `env:"ADDRESS"`
	StoreInterval customtype.Time     `env:"STORE_INTERVAL"`
	File          string              `env:"FILE_STORAGE_PATH"`
	Restore       bool                `env:"RESTORE"`
	DSN           customtype.DSN      `env:"DATABASE_DSN"`
	HashKey       string              `env:"KEY"`
	AuditFile     string              `env:"AUDIT_FILE"`
	AuditURL      string              `env:"AUDIT_URL"`
	CryptoKey     string              `env:"CRYPTO_KEY"`
	TrustedSubnet string              `env:"TRUSTED_SUBNET"`
	GRPCAddress   string              `env:"GRPC_ADDRESS"`
	Logger        *zap.Logger
}

var (
	DefaultEndpoint      = customtype.Endpoint("http://localhost:8080")
	DefaultStoreInterval = customtype.Time(300 * time.Second)
	DefaultFile          = "Metrics.data"
	DefaultRestore       = true
	DefaultTrustedSubnet = ""
)
