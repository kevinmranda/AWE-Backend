package models

import (
	"gorm.io/gorm"
)

type ActivityLog struct {
	gorm.Model
	Level   string
	Message string
	// Body    string
}
