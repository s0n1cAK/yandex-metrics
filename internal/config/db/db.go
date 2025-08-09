package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"github.com/s0n1cAK/yandex-metrics/internal/storage/dbStorage/retries"
)

func InitMigration(ctx context.Context, DSN customtype.DSN) error {
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

func PingDB(ctx context.Context, DSN string) error {
	db, err := retries.OpenDBWithRetry(DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

type DBPinger struct {
	DSN customtype.DSN
}

func NewPinger(DSN customtype.DSN) *DBPinger {
	return &DBPinger{DSN: DSN}
}

func (p *DBPinger) Ping(ctx context.Context) error {
	return PingDB(ctx, p.DSN.String())
}
