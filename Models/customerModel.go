package models

import "gorm.io/gorm"

type Customer struct {
	gorm.Model
	CustomerEmail string
	Orders        []Order // One-to-Many relationship with order
}
