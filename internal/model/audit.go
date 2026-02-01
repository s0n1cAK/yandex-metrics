package model

type AuditEvent struct {
	TS        int64     `json:"ts"`
	Metrics   []Metrics `json:"metrics"`
	IPAddress string    `json:"ip_address"`
}
