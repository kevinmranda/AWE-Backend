package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	models "github.com/kevinmranda/AWE-Backend/Models"
)

func AddProduct(c *gin.Context) {
	c.Get("user")
	Uploaded_bystr := c.Param("id")
	Uploaded_by64, _ := strconv.ParseUint(Uploaded_bystr, 10, 32)
	Uploaded_by := uint(Uploaded_by64)
	// struct for contents to be saved
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Filename    string  `json:"filename"` //(path to the high-quality image)
		Price       float64 `json:"price"`
	}

	// Get contents from the body of the request and Bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Create a product instance and save to DB
	product := models.Product{
		Title:       body.Title,
		Description: body.Description,
		Filename:    body.Filename,
		Price:       body.Price,
		User_id:     Uploaded_by,
	}

	result := initializers.DB.Create(&product)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":    2000,
			"error": "Failed to insert the record",
		})
		return
	}

	// Respond
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "Record inserted successfully",
		"data":    product,
	})
}

func Upload(c *gin.Context) {
	c.Get("user")
	// Parse the multipart form
	err := c.Request.ParseMultipartForm(20 << 20) // 20 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get the file from the request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to retrieve the file",
		})
		return
	}
	defer file.Close()

	// Define the file path
	filename := header.Filename
	filePath := fmt.Sprintf("Products/%s", filename)

	// Create the Photos directory if it doesn't exist
	if err := os.MkdirAll("Products", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create directory",
		})
		return
	}

	// Save the file to the products folder
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save the file",
		})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save the file",
		})
		return
	}
}

func GetProduct(c *gin.Context) {

	// Get filename from request
	filename := c.Param("filename")

	// Define the directory where photos are stored
	filepath := "Products/" + filename

	// Check if the file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// If file doesn't exist, return a 404
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "file not present",
		})
		return
	}

	// Serve the file
	c.File(filepath)
}

// get all products
func GetProducts(c *gin.Context) {
	c.Get("user")
	id := c.Param("id")
	var products []models.Product

	// Preload the Orders relationship to include associated orders in the results
	result := initializers.DB.Preload("Orders").Where("user_id = ?", id).Find(&products)

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
		"data":    products,
	})
}

func GetAllProducts(c *gin.Context) {
	var products []models.Product

	initializers.DB.Find(&products)

	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    products,
	})
}

// update a product with id
func UpdateProduct(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	//struct for contents to be updated
	var contentForUpdate struct {
		ID          uint   `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Filename    string `json:"filename"` //(path to the high-quality image)
		Price       string `json:"price"`
	}

	//Get contents from body of request and Bind JSON input to the struct
	if err := c.ShouldBindJSON(&contentForUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Check if the product to be updated exists
	var product models.Product
	result := initializers.DB.Preload("Orders").First(&product, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	//update product with struct provided on request body
	if contentForUpdate.Title != "" {
		product.Title = contentForUpdate.Title
	}

	if contentForUpdate.Description != "" {
		product.Description = contentForUpdate.Description
	}

	if contentForUpdate.Filename != "" {
		product.Filename = contentForUpdate.Filename
	}

	if contentForUpdate.Price != "" {
		iPrice, _ := strconv.ParseFloat(contentForUpdate.Price, 64)
		product.Price = iPrice
	}

	result = initializers.DB.Save(&product)
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
		"data":    product,
	})
}

// delete a product with specified id
func DeleteProduct(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	var product models.Product
	// Check if the photo exists
	initializers.DB.First(&product, id)

	if product.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	// Delete the product
	result := initializers.DB.Delete(&product)

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
