package model

import "time"

type Deployment struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	ServerID    uint            `json:"server_id"`
	Type        string          `json:"type"` // nginx, backend, full
	Status      string          `gorm:"default:'pending'" json:"status"` // pending, running, completed, failed
	Config      string          `json:"config"` // JSON: binary_url, service_name, port, domain, etc.
	Logs        string          `json:"logs"`
	TriggeredBy string          `json:"triggered_by"` // agent, manual
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Server      *Server         `gorm:"foreignKey:ServerID" json:"server,omitempty"`
}
