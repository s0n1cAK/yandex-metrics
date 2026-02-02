package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/s0n1cAK/yandex-metrics/internal/model"
)

type FileAuditObserver struct {
	path string
	mu   sync.Mutex
}

func NewFileAuditObserver(path string) *FileAuditObserver {
	return &FileAuditObserver{path: path}
}

func (f *FileAuditObserver) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Открытие файла вне критической секции
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", f.path, err)
	}
	defer file.Close()

	f.mu.Lock()
	defer f.mu.Unlock()

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", f.path, err)
	}

	return nil
}
