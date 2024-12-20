package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	Customer_email string
	Total_amount   float64
	Status         string  `gorm:"default:pending"` //(enum: "pending", "completed", "canceled")
	Products         []Product `gorm:"many2many:order_products;"`
	Payment        Payment // One-to-One relationship with Payment
	CustomerID     uint
}
