package retries

import (
	"database/sql"
	"fmt"
	"time"
)

func ExecuteWithRetry(action func() error) error {
	const maxRetries = 3
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	classifier := NewPostgresErrorClassifier()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := action()
		if err == nil {
			return nil
		}

		class := classifier.Classify(err)
		if class == NonRetriable {
			return fmt.Errorf("неповторяемая ошибка: %w", err)
		}

		lastErr = err
		if attempt < maxRetries-1 {
			time.Sleep(delays[attempt])
		}
	}

	return fmt.Errorf("операция прервана после %d попыток: %w", maxRetries, lastErr)
}

func OpenDBWithRetry(DSN string) (*sql.DB, error) {
	const maxRetries = 3
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		db, err := sql.Open("postgres", DSN)
		if err == nil {
			return db, err
		}

		lastErr = err
		if attempt < maxRetries-1 {
			time.Sleep(delays[attempt])
		}
	}

	return nil, fmt.Errorf("операция прервана после %d попыток: %w", maxRetries, lastErr)
}
