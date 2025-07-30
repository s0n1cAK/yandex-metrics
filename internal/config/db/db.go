package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/s0n1cAK/yandex-metrics/internal/config"
)

func InitMigration(ctx context.Context, DSN config.DSN) error {
	db, err := sql.Open("pgx", DSN.String())
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := migrate.New(
		"file://internal/migrations",
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", DSN.User, DSN.Password, DSN.Host, DSN.Name, DSN.SSLMode),
	)
	if err != nil {
		return fmt.Errorf("Неудалось создать миграцию: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("миграция не выполнена: %v", err)
	}

	return nil
}

func PingDB(ctx context.Context, DSN string) error {
	db, err := sql.Open("pgx", DSN)
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
