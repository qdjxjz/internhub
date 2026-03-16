package model

import "time"

// UserProfile 用户扩展资料，与 auth-service 的用户通过 user_id 关联
type UserProfile struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	Nickname  string    `gorm:"size:128" json:"nickname"`
	Avatar    string    `gorm:"size:512" json:"avatar"`
	UpdatedAt time.Time `json:"updated_at"`
}
