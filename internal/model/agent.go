package model

import "time"

type Agent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // claude, openai, custom
	Endpoint  string    `json:"endpoint"`
	APIKey    string    `json:"api_key"` // encrypted
	Config    string    `json:"config"`  // JSON
	Status    string    `gorm:"default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
