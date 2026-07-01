package model

import "time"

type Command struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID uint      `json:"session_id"`
	Command   string    `json:"command"`
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}
