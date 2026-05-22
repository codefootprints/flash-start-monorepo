package models

import (
	"time"

	"gorm.io/gorm"
)

type Resource struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Category  string         `json:"category"`
	Status    string         `gorm:"default:'available'" json:"status"` // available, in_use, maintenance
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
