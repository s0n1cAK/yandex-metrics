package audit

import (
	"encoding/json"
	"os"

	"github.com/s0n1cAK/yandex-metrics/internal/model"
)

type FileAuditObserver struct {
	path string
}

func NewFileAuditObserver(path string) *FileAuditObserver {
	return &FileAuditObserver{path: path}
}

func (f *FileAuditObserver) Notify(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(append(data, '\n'))
	return err
}
