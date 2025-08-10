package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
)

func Migration(ctx context.Context, DSN customtype.DSN) error {
	m, err := migrate.New(
		"file://migrations",
		DSN.String(),
	)

	if err != nil {
		return fmt.Errorf("неудалось создать миграцию: %v", err)
	}

	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("миграция не выполнена: %w", err)
	}

	return nil
}
