package audit

import "github.com/s0n1cAK/yandex-metrics/internal/model"

type AuditPublisher struct {
	observers []AuditObserver
}

func (p *AuditPublisher) Register(o AuditObserver) {
	p.observers = append(p.observers, o)
}

func (p *AuditPublisher) Publish(event model.AuditEvent) error {
	for _, obs := range p.observers {
		err := obs.Notify(event)
		if err != nil {
			return err
		}
	}
	return nil
}
