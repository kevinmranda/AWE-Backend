package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	First_name   string
	Last_name    string
	Password     string
	Gender       string
	Birthdate    time.Time
	Address      string
	Email        string `gorm:"unique"`
	Mobile       string
	Role         string  `gorm:"default:regular_user"`
	Products       []Product // One-to-Many relationship with Photo
	Token        Token
	Subscription bool   `gorm:"default:false"`
	Theme        string `gorm:"default:bootstrap4-light-purple"`
	Language     string `gorm:"default:en"`
}
