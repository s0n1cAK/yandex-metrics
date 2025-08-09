package db

import (
	"context"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	"github.com/s0n1cAK/yandex-metrics/internal/storage/dbStorage/retries"
)

type DBPinger struct {
	DSN customtype.DSN
}

func NewPinger(DSN customtype.DSN) *DBPinger {
	return &DBPinger{DSN: DSN}
}

func (p *DBPinger) Ping(ctx context.Context) error {
	return PingDB(ctx, p.DSN.String())
}

func PingDB(ctx context.Context, DSN string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := retries.OpenDBWithRetry(ctx, DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
