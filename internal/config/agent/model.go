package agent

import (
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"go.uber.org/zap"
)

type Config struct {
	Client         *retryablehttp.Client
	Endpoint       customtype.Endpoint `env:"ADDRESS"`
	ReportInterval customtype.Time     `env:"REPORT_INTERVAL"`
	PollInterval   customtype.Time     `env:"POLL_INTERVAL"`
	Hash           string              `env:"KEY"`
	RateLimit      int                 `env:"RATE_LIMIT"`
	CryptoKey      string              `env:"CRYPTO_KEY"`
	GRPCAddress    string              `env:"GRPC_ADDRESS"`
	Logger         *zap.Logger
}

var (
	DefaultEndpoint       = customtype.Endpoint("http://localhost:8080")
	DefaultReportInterval = customtype.Time(10 * time.Second)
	DefaultPollInterval   = customtype.Time(2 * time.Second)
	DefaultHashKey        = ""
	DefaultRateLimit      = 10
)
