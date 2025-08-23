package filestorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	models "github.com/s0n1cAK/yandex-metrics/internal/model"
)

type (
	Producer struct {
		file          *os.File
		storeInterval time.Duration
		writer        *bufio.Writer
	}

	Consumer struct {
		file *os.File
	}
)

func NewProducer(filename string, storeInterval time.Duration) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{file: file, storeInterval: storeInterval, writer: bufio.NewWriter(file)}, nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func (p *Producer) WriteMetrics(metrics map[string]models.Metrics) error {
	p.file.Truncate(0)

	var storage []models.Metrics

	for _, value := range metrics {
		storage = append(storage, value)
	}

	data, err := json.MarshalIndent(storage, "", " ")
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) WriteMetric(metric models.Metrics) error {
	consumer, err := NewConsumer(p.file.Name())
	if err != nil {
		return err
	}

	oldMetrics, err := consumer.ReadFile()
	if err != nil {
		return err
	}

	storageMap := make(map[string]models.Metrics, len(oldMetrics))
	for _, m := range oldMetrics {
		storageMap[m.ID] = m
	}

	storageMap[metric.ID] = metric

	return p.WriteMetrics(storageMap)
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{file: file}, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (c *Consumer) ReadFile() ([]models.Metrics, error) {
	op := "ReadFile"
	var metrics []models.Metrics

	_, err := c.file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("%s: seek error: %w", op, err)
	}

	data, err := io.ReadAll(c.file)
	if err != nil {
		return nil, fmt.Errorf("%s: read error: %w", op, err)
	}

	if len(data) == 0 {
		return []models.Metrics{}, nil
	}

	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("%s: json unmarshal error: %w", op, err)
	}

	return metrics, nil
}
