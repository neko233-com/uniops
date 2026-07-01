package model

import "time"

type SSHKey struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	PublicKey   string    `gorm:"type:text" json:"public_key"`
	PrivateKey  string    `gorm:"type:text" json:"private_key"`
	Fingerprint string    `json:"fingerprint"`
	ServerID    uint      `json:"server_id"`
	Status      string    `gorm:"default:active" json:"status"` // active, deployed, pending
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
