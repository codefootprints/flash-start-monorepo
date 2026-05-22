package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Status      string         `gorm:"default:'todo'" json:"status"` // todo, in_progress, done
	UserID      uint           `json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user"`
	ResourceID  uint           `json:"resource_id"`
	Resource    Resource       `gorm:"foreignKey:ResourceID" json:"resource"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
