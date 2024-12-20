package models

import (
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	Order_id       uint //(Foreign Key to Order) Order ID
	Amount         float64
	Status         string //(enum: "pending", "completed", "failed")
	Payment_method string //(e.g., "credit_card", "paypal")
	Transaction_id string //(Identifier from payment gateway)
}
