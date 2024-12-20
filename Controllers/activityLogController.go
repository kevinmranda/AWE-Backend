package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	models "github.com/kevinmranda/AWE-Backend/Models"
)

func GetLogs(c *gin.Context) {
	c.Get("user")
	var logs []models.ActivityLog
	result := initializers.DB.Find(&logs)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "logs not present",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    logs,
	})
}
