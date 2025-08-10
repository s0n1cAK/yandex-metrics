-- Для тестов
CREATE TABLE praktikum (
    name VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL CHECK (type IN ('counter', 'gauge')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255)
);


-- Создание таблицы метрик
CREATE TABLE metrics (
    name VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL CHECK (type IN ('counter', 'gauge')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255)
);

CREATE INDEX idx_metric_name ON metrics(name);

CREATE INDEX idx_metric_hash ON metrics(hash); 
