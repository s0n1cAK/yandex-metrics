package audit

import "github.com/s0n1cAK/yandex-metrics/internal/model"

type AuditObserver interface {
	Notify(event model.AuditEvent) error
}
