package migrations

import (
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	models "github.com/kevinmranda/AWE-Backend/Models"
)

// migrations for the database are found here
func SyncDatabase() {
	//Migrate Schema
	initializers.DB.AutoMigrate(&models.User{})
	initializers.DB.AutoMigrate(&models.Product{})
	initializers.DB.AutoMigrate(&models.Order{})
	initializers.DB.AutoMigrate(&models.Payment{})
	initializers.DB.AutoMigrate(&models.Token{})
	initializers.DB.AutoMigrate(&models.Customer{})
}
