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
	Logger         *zap.Logger
	Hash           string `env:"KEY"`
}

var (
	DefaultEndpoint       = customtype.Endpoint("http://localhost:8080")
	DefaultReportInterval = customtype.Time(10 * time.Second)
	DefaultPollInterval   = customtype.Time(2 * time.Second)
	DefaultHashKey        = ""
)
