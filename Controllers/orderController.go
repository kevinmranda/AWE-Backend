package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	models "github.com/kevinmranda/AWE-Backend/Models"
)

// add order to db
func AddOrder(c *gin.Context) {
	c.Get("user")
	// Bind JSON input to the struct
	var body struct {
		Customer_email string `json:"customer_email" binding:"required"`
		ProductIDs     []uint `json:"product_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	if ValidateEmail(body.Customer_email) {
		// Retrieve the associated products based on PhotoIDs
		var products []models.Product
		if err := initializers.DB.Where("id IN ?", body.ProductIDs).Find(&products).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"id":    2011,
				"error": "Failed to retrieve products",
			})
			return
		}

		var customer models.Customer
		if err := initializers.DB.Where("customer_email = ?", body.Customer_email).First(&customer).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"id":    2011,
				"error": "Failed to retrieve customer",
			})
			return
		}
		// Create the order record
		total_amount := 0.0
		for _, product := range products {
			total_amount += product.Price
		}

		order := models.Order{
			// Customer_email: body.Customer_email,
			Total_amount: total_amount,
			Products:     products,
			CustomerID:   customer.ID,
		}

		if err := initializers.DB.Create(&order).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":    2000,
				"error": "Failed to insert the record",
			})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"id":      2001,
			"message": "Order created successfully",
			"order":   order,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2002,
			"error": "Invalid email format",
		})
	}

}

// delete order
func RemoveOrder(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	var order models.Order
	// Check if the product exists
	initializers.DB.First(&order, id)

	if order.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	// Delete the order
	result := initializers.DB.Delete(&order)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":    2013,
			"error": "failed to delete record",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2012,
		"message": "record deleted successfully",
	})
}

// retrieve an order with id
func GetOrder(c *gin.Context) {
	c.Get("user")
	id := c.Param("id")
	var order models.Order
	// Preload the many-to-many relationship with Products and the one-to-one relationship with Payment
	result := initializers.DB.
		Joins("JOIN order_products ON orders.id = order_products.order_id").
		Joins("JOIN products ON order_products.product_id = products.id").
		Joins("JOIN customers ON orders.customer_id = customers.id"). // Join the customer table
		Preload("Products").
		Preload("Payment").
		Select("orders.*, customers.customer_email"). // Select customer email
		Find(&order, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "records not present",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    order,
	})
}

// retrieve all orders
func GetOrders(c *gin.Context) {
	c.Get("user")
	id := c.Param("id")
	var orders []models.Order
	// Preload the many-to-many relationship with Products and the one-to-one relationship with Payment
	result := initializers.DB.
		Joins("JOIN order_products ON orders.id = order_products.order_id").
		Joins("JOIN products ON order_products.product_id = products.id").
		Joins("JOIN customers ON orders.customer_id = customers.id").
		Where("products.user_id = ?", id).
		Preload("Products").
		Preload("Payment").
		Group("orders.id").
		Select("orders.*, MAX(customers.customer_email) as Customer_email").
		Find(&orders)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "records not present",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    orders,
	})
}

// update an order with id
func UpdateOrder(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	//struct for contents to be updated
	var contentForUpdate struct {
		Customer_email string  `json:"customer_email"`
		Total_amount   float64 `json:"total_amount"`
		Status         string  `json:"status"`
		ProductIDs     []uint  `json:"product_ids"`
	}

	//Get contents from body of request and Bind JSON input to the struct
	if err := c.ShouldBindJSON(&contentForUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Check if the order to be updated exists
	var order models.Order
	result := initializers.DB.Preload("Products").Preload("Payment").First(&order, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	//update order with struct provided on request body
	// if contentForUpdate.Customer_email != "" {
	// 	order.Customer_email = contentForUpdate.Customer_email
	// }

	if contentForUpdate.Total_amount != 0 {
		order.Total_amount = contentForUpdate.Total_amount
	}

	if contentForUpdate.Status != "" {
		order.Status = contentForUpdate.Status
	}

	if contentForUpdate.ProductIDs != nil {
		// Retrieve the associated products based on ProductIDs
		var products []models.Product
		if err := initializers.DB.Where("id IN ?", contentForUpdate.ProductIDs).Find(&products).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"id":    2011,
				"error": "Failed to retrieve photos",
			})
			return
		}
		order.Products = products
	}

	result = initializers.DB.Save(&order)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2014,
			"error": "Failed to update the record",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    order,
	})
}
