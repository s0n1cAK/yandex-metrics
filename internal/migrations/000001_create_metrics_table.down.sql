-- Создание таблицы метрик
CREATE TABLE metrics (
    name VARCHAR(255) PRIMARY KEY,
    mtype VARCHAR(255) NOT NULL CHECK (mtype IN ('counter', 'gauge')),
    delta INTEGER,q
    value DOUBLE PRECISION,
    hash VARCHAR(255)
);

CREATE INDEX idx_metric_name ON metrics(name);

CREATE INDEX idx_metric_hash ON metrics(hash); 