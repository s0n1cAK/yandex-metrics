package dbstorage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/s0n1cAK/yandex-metrics/internal/config"
	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

type PostgresStorage struct {
	db        *sql.DB
	tableName string
	ctx       context.Context
}

func NewPostgresStorage(ctx context.Context, DSN config.DSN) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", DSN.String())
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
	q := fmt.Sprintf(`
		INSERT INTO %s (name, mtype, delta, value, hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name)
		DO UPDATE SET mtype = $2, delta = $3, value = $4, hash = $5;
		`, p.tableName)
	_, err := p.db.ExecContext(p.ctx, q, key, value.MType, value.Delta, value.Value, value.Hash)
	return err
}

func (p *PostgresStorage) Get(key string) (models.Metrics, bool) {
	q := `SELECT name, mtype, delta, value, hash FROM metrics WHERE = $1`
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

func (p *PostgresStorage) GetAll() map[string]models.Metrics {
	q := `SELECT name, mtype, delta, value, hash FROM metrics`
	rows, err := p.db.QueryContext(p.ctx, q)
	if err != nil {
		return nil
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
	return result
}

func (p *PostgresStorage) SetAll([]models.Metrics) {
	// заглушка
}

/*
	Set(key string, value models.Metrics) error
	Get(key string) (models.Metrics, bool)
	GetAll() map[string]models.Metrics
	SetAll([]models.Metrics)
*/
