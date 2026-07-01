package model

import "time"

type Session struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `json:"user_id"`
	ServerID  uint       `json:"server_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Status    string     `gorm:"default:active" json:"status"` // active, closed
	Replay    string     `json:"replay"` // base64 encoded
}
