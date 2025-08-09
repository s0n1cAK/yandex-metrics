package dbstorage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/s0n1cAK/yandex-metrics/internal/customtype"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
	"github.com/s0n1cAK/yandex-metrics/internal/storage/dbStorage/retries"
)

type PostgresStorage struct {
	db        *sql.DB
	tableName string
	ctx       context.Context
}

func NewPostgresStorage(ctx context.Context, DSN customtype.DSN) (*PostgresStorage, error) {
	db, err := retries.OpenDBWithRetry(DSN.String())

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db:        db,
		tableName: DSN.Name,
		ctx:       ctx,
	}, nil

}

func (p *PostgresStorage) Set(key string, value models.Metrics) error {
	err := retries.ExecuteWithRetry(func() error {
		q := fmt.Sprintf(`
		INSERT INTO %s (name, type, delta, value, hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name)
		DO UPDATE SET
			type = EXCLUDED.type,
			delta = %s.delta + EXCLUDED.delta,
			value = EXCLUDED.value,
			hash = EXCLUDED.hash;
		`, p.tableName, p.tableName)

		_, err := p.db.ExecContext(p.ctx, q, key, value.MType, value.Delta, value.Value, value.Hash)

		return err
	})

	return err
}

func (p *PostgresStorage) Get(key string) (models.Metrics, bool) {
	q := fmt.Sprintf(`SELECT name, type, delta, value, hash FROM %s WHERE name = $1`, p.tableName)
	row := p.db.QueryRowContext(p.ctx, q, key)

	var m models.Metrics
	err := row.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	if err == sql.ErrNoRows {
		return models.Metrics{}, false
	}
	if err != nil {
		return models.Metrics{}, false
	}
	return m, true
}

func (p *PostgresStorage) GetAll() (map[string]models.Metrics, error) {
	q := fmt.Sprintf(`SELECT name, type, delta, value, hash FROM %s`, p.tableName)
	rows, err := p.db.QueryContext(p.ctx, q)
	if err != nil {
		return map[string]models.Metrics{}, err
	}
	if err := rows.Err(); err != nil {
		return map[string]models.Metrics{}, err
	}
	defer rows.Close()

	result := make(map[string]models.Metrics)
	for rows.Next() {
		var m models.Metrics
		if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
			continue
		}
		result[m.ID] = m
	}
	return result, nil
}

func (p *PostgresStorage) SetAll(metrics []models.Metrics) error {
	err := retries.ExecuteWithRetry(func() error {
		tx, err := p.db.Begin()
		if err != nil {
			return err
		}

		defer tx.Rollback()

		stmt, err := tx.PrepareContext(p.ctx, fmt.Sprintf(`
			INSERT INTO %s (name, type, delta, value, hash) 
			VALUES ($1, $2, $3, $4, $5) 		
			ON CONFLICT (name)
			DO UPDATE SET 
				type = EXCLUDED.type,
				delta = %s.delta + EXCLUDED.delta,
				value = EXCLUDED.value,
				hash = EXCLUDED.hash;
			`, p.tableName, p.tableName))
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, val := range metrics {
			_, err := stmt.ExecContext(p.ctx, val.ID, val.MType, val.Delta, val.Value, val.Hash)
			if err != nil {
				return err
			}
		}
		return tx.Commit()
	})
	return err
}

/*
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() map[string]models.Metrics
	SetAll([]models.Metrics)
*/
