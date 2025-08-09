package agent

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// https://pkg.go.dev/github.com/hashicorp/go-retryablehttp
func configureRetries(cfg Config) {
	cfg.Client.RetryMax = 3
	cfg.Client.RetryWaitMin = 1 * time.Second
	cfg.Client.RetryWaitMax = 5 * time.Second
	cfg.Client.Backoff = retryablehttp.RateLimitLinearJitterBackoff

	cfg.Client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, nil
		}
		if resp == nil {
			return false, nil
		}
		switch resp.StatusCode {
		case 408, 429:
			return true, nil
		}
		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			return true, nil
		}
		return false, nil
	}

}
