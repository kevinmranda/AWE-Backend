package models

import (
	"time"

	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	UserID uint
	Token  string
	Expiry time.Time
}
