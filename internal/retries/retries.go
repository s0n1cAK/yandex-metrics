package retries

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var (
	delays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
)

func DoWithRetry(ctx context.Context, action func() error, isRetriable func(error) bool, delays []time.Duration) error {
	var lastErr error

	for i := 0; i < len(delays)+1; i++ {
		if err := action(); err == nil {
			return nil
		} else if !isRetriable(err) {
			return fmt.Errorf("неповторяемая ошибка: %w", err)
		} else {
			lastErr = err
		}
		if i < len(delays) {
			select {
			case <-time.After(delays[i]):
			case <-ctx.Done():
				return fmt.Errorf("отменено: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("операция прервана после %d попыток: %w", len(delays)+1, lastErr)
}

func ExecuteWithRetry(ctx context.Context, action func() error) error {
	classifier := NewPostgresErrorClassifier()

	isRetriable := func(err error) bool {
		return classifier.Classify(err) != NonRetriable
	}

	return DoWithRetry(ctx, action, isRetriable, delays)
}

func OpenDBWithRetry(ctx context.Context, dsn string) (*sql.DB, error) {
	var db *sql.DB
	action := func() error {
		var err error
		db, err = sql.Open("postgres", dsn)
		return err
	}

	isRetriable := func(error) bool { return true }

	if err := DoWithRetry(ctx, action, isRetriable, delays); err != nil {
		return nil, err
	}
	return db, nil
}
