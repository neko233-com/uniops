package model

import "time"

type Server struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `gorm:"default:22" json:"port"`
	Username  string    `json:"username"`
	AuthType  string    `json:"auth_type"` // password, key
	AuthData  string    `json:"auth_data"` // encrypted
	AgentID   uint      `json:"agent_id"`
	GroupID   uint      `json:"group_id"`
	Status    string    `gorm:"default:offline" json:"status"` // online, offline
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
