package controllers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	models "github.com/kevinmranda/AWE-Backend/Models"
	"golang.org/x/crypto/bcrypt"
)

func CreateAccount(c *gin.Context) {

	//struct of the request body
	var body struct {
		First_name string `json:"first_name" binding:"required"`
		Last_name  string `json:"last_name" binding:"required"`
		Password   string `json:"password" binding:"required"`
		Gender     string `json:"gender" binding:"required"`
		Birthdate  string `json:"birthdate" binding:"required"`
		Address    string `json:"address"`
		Email      string `json:"email" binding:"required"`
		Mobile     string `json:"mobile" binding:"required"`
	}

	//Get contents from body of request and Bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Parse the birthdate string into a time.Time object
	birthdate, err := time.Parse("2006-01-02", body.Birthdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid birthdate format, expected YYYY-MM-DD",
		})
		return
	}

	// Validate user input
	isValidName, isValidEmail, isValidPassword, isValidPhoneNumber := validateUserInput(
		body.First_name,
		body.Last_name,
		body.Email,
		body.Password,
		body.Mobile,
	)

	if isValidName && isValidEmail && isValidPassword && isValidPhoneNumber {
		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to hash password",
			})
			return
		}

		// Create user
		user := models.User{
			First_name: body.First_name,
			Last_name:  body.Last_name,
			Password:   string(hash),
			Gender:     body.Gender,
			Birthdate:  birthdate,
			Address:    body.Address,
			Email:      body.Email,
			Mobile:     body.Mobile,
		}
		result := initializers.DB.Create(&user)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":    2000,
				"error": "Failed to insert the record",
			})
			return
		}

		sendEmail(body.Email, body.First_name, body.Last_name)

		// Respond
		c.JSON(http.StatusOK, gin.H{
			"id":      2001,
			"message": "success",
			"data":    user,
		})
	} else {
		if !isValidName {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":    2002,
				"error": "Invalid first name or last name format",
			})
		}
		if !isValidEmail {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":    2002,
				"error": "Invalid email format",
			})
		}
		if !isValidPassword {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":    2002,
				"error": "Weak password. The password should be at least 8 characters long and include special characters.",
			})
		}
		if !isValidPhoneNumber {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":    2002,
				"error": "Phone number is not valid",
			})
		}
	}
}

func Login(c *gin.Context) {
	//get email and password off body
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//look up for user
	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2005,
			"error": "Invalid email",
			"data":  "",
		})
		return
	}

	//compare hash and password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2006,
			"error": "Invalid password",
			"data":  "",
		})
		return
	}

	//create token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour).Unix(), //1 hour expiry
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2007,
			"error": "Failed to create Token",
		})
		return
	}

	//make and set cookie
	SetAuthCookie(c, tokenString)

	//login
	c.JSON(200, gin.H{
		"id":      2008,
		"message": "Login Successfully",
		"data":    user,
		"token":   tokenString,
	})
}

// sends email to users who have created accounts
func sendEmail(email, firstName, lastName string) {
	// Set up SMTP authentication
	eml := os.Getenv("STMP_EMAIL")
	pswd := os.Getenv("STMP_PASSWORD")
	host := os.Getenv("STMP_HOST")
	addr := os.Getenv("STMP_SERVER_ADDR")

	auth := smtp.PlainAuth("", eml, pswd, host)

	// Set up email details
	from := eml
	to := []string{email}
	subject := "Your ClickMaster Account is Ready!"

	// Email body template
	body := fmt.Sprintf(`
Hello %s %s,

Thank you for creating an account at ClickMaster. Your account is now ready.

Should you have any questions about ClickMaster, visit our support page: https://support.clickmaster.com

Kind regards,

Your ClickMaster Team

---

ClickMaster Inc.
www.clickmaster.com | support@clickmaster.com
Managing Director: David Robert
VAT-Number: DE294776378

ClickMaster Inc., Kinondoni Mkwajuni, Dar es Salaam, Tanzania.
`, firstName, lastName)

	// Create a byte slice for the email message
	msg := bytes.NewBufferString("Subject: " + subject + "\r\n\r\n" + body)

	// Send the email
	err := smtp.SendMail(addr, auth, from, to, msg.Bytes())
	if err != nil {
		log.Fatalf("Error sending email: %v", err)
	}
}

// validates user inputs
func validateUserInput(firstName, lastName, email, password, mobile string) (bool, bool, bool, bool) {
	isValidName := len(firstName) > 0 && len(lastName) > 0
	isValidEmail := ValidateEmail(email)
	isValidPassword := validatePassword(password)
	isValidPhoneNumber := validatePhoneNumber(mobile)

	return isValidName, isValidEmail, isValidPassword, isValidPhoneNumber
}

// validateName
func validateName(name string) bool {
	if len(name) > 0 {
		return true
	}
	return true
}

// ValidateEmail checks if the email is in a valid format.
func ValidateEmail(email string) bool {
	// Simple regex pattern for email validation
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

// validatePassword checks if the password is strong enough
func validatePassword(password string) bool {
	var hasMinLen, hasUpper, hasLower, hasNumber, hasSpecial bool
	hasMinLen = len(password) >= 8

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

// validate phone number
func validatePhoneNumber(mobile string) bool {
	// Regular expression to match the desired phone number formats
	re := regexp.MustCompile(`^(?:\+255|255|0)\d{9}$`)
	// Validate the phone number using the regular expression
	return re.MatchString(mobile)
}

// set cookie
func SetAuthCookie(c *gin.Context, tokenString string) {
	// Clear any existing "Auth" cookie
	if cookie, err := c.Cookie("Auth"); err == nil && cookie != "" {
		c.SetCookie("Auth", "", -1, "/", "localhost", false, true)
	}

	// Set the SameSite attribute
	c.SetSameSite(http.SameSiteLaxMode)

	// Set the new "Auth" cookie with an hour expiration
	// expiration := time.Now().Add(time.Hour)
	c.SetCookie(
		"Auth",      // Name
		tokenString, // Value
		3600,        // MaxAge (24 hours)
		"/",         // Path
		"",          // Domain
		false,       // Secure (false since we're using HTTP)
		true,        // HttpOnly
	)
}

func DeleteUser(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	var user models.User
	// Check if the user exists
	initializers.DB.First(&user, id)

	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "user not found",
		})
		return
	}

	// Delete the user
	result := initializers.DB.Delete(&user)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":    2013,
			"error": "failed to delete user",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2012,
		"message": "user deleted successfully",
	})
}

func GetUser(c *gin.Context) {
	// Get id from request
	id := c.Param("id")

	var user models.User

	// Check if the user exists
	initializers.DB.First(&user, id)

	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    user,
	})
}

func GetUsers(c *gin.Context) {
	var users []models.User

	//retrieve all users
	result := initializers.DB.Preload("Photos").Find(&users)

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
		"data":    users,
	})
}

// update user
func UpdateUser(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	var body struct {
		First_name string `json:"first_name"`
		Last_name  string `json:"last_name"`
		Password   string `json:"password"`
		Gender     string `json:"gender"`
		Birthdate  string `json:"birthdate"`
		Address    string `json:"address"`
		Email      string `json:"email"`
		Mobile     string `json:"mobile"`
	}

	// Get contents from body of request and bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Check if the user to be updated exists
	var user models.User
	result := initializers.DB.Preload("Photos").First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	// Validate and update fields only if they are not empty in the request body
	errors := gin.H{}

	if body.First_name != "" || body.Last_name != "" {
		if !validateName(body.First_name) || !validateName(body.Last_name) {
			errors["name"] = "Invalid first name or last name format"
		} else {
			if body.First_name != "" {
				user.First_name = body.First_name
			}
			if body.Last_name != "" {
				user.Last_name = body.Last_name
			}
		}
	}

	if body.Password != "" {
		if !validatePassword(body.Password) {
			errors["password"] = "Weak password. The password should be at least 8 characters long and include special characters."
		} else {
			hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Failed to hash password",
				})
				return
			}
			user.Password = string(hash)
		}
	}

	if body.Email != "" {
		if !ValidateEmail(body.Email) {
			errors["email"] = "Invalid email format"
		} else {
			user.Email = body.Email
		}
	}

	if body.Mobile != "" {
		if !validatePhoneNumber(body.Mobile) {
			errors["phone_number"] = "Phone number is not valid"
		} else {
			user.Mobile = body.Mobile
		}
	}

	if body.Gender != "" {
		user.Gender = body.Gender
	}

	if body.Birthdate != "" {
		birthdate, err := time.Parse("2006-01-02", body.Birthdate)
		if err != nil {
			errors["birthdate"] = "Invalid birthdate format, expected YYYY-MM-DD"
		} else {
			user.Birthdate = birthdate
		}
	}

	if body.Address != "" {
		user.Address = body.Address
	}

	// If there are validation errors, return them
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2002,
			"error": errors,
		})
		return
	}

	// Save updated user
	result = initializers.DB.Save(&user)
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
		"data":    user,
	})
}

func SendResetPasswordEmail(c *gin.Context) {

	//struct of the request body
	var EmailBody struct {
		Email string `json:"email" binding:"required"`
	}

	//Get contents from body of request and Bind JSON input to the struct
	if err := c.ShouldBindJSON(&EmailBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Look up the user by email
	var user models.User
	initializers.DB.First(&user, "email = ?", EmailBody.Email)
	firstName := user.First_name
	lastName := user.Last_name

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2005,
			"error": "Invalid email",
			"data":  "",
		})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour).Unix(), // 1-hour expiry
		"iat": time.Now().Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2007,
			"error": "Failed to create token",
		})
		return
	}

	// Store hashed token in the database with an expiration time
	saveTokenToDB(user.ID, tokenString, time.Now().Add(time.Hour))

	// Set up SMTP authentication
	eml := os.Getenv("STMP_EMAIL")
	pswd := os.Getenv("STMP_PASSWORD")
	host := os.Getenv("STMP_HOST")
	addr := os.Getenv("STMP_SERVER_ADDR")

	auth := smtp.PlainAuth("", eml, pswd, host)

	// Set up email details
	from := eml
	to := []string{EmailBody.Email}
	subject := "Password Reset Request"

	// Email body template
	body := fmt.Sprintf(`
Hello %s %s,

Click the link to reset your password: http://localhost:4200/resetPassword?token=%s

Should you have any questions about ClickMaster, visit our support page: https://support.clickmaster.com

Kind regards,

Your ClickMaster Team

---

ClickMaster Inc.
www.clickmaster.com | support@clickmaster.com
Managing Director: David Robert
VAT-Number: DE294776378

ClickMaster Inc., Kinondoni Mkwajuni, Dar es Salaam, Tanzania.
`, firstName, lastName, tokenString)

	// Create a byte slice for the email message
	msg := []byte("To: " + EmailBody.Email + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	// Send the email
	err = smtp.SendMail(addr, auth, from, to, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":    2009,
			"error": "Failed to send email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
	})
}

func saveTokenToDB(userID uint, token string, expiry time.Time) {
	// Create user
	tokenToSave := models.Token{
		UserID: userID,
		Token:  token,
		Expiry: expiry,
	}
	initializers.DB.Create(&tokenToSave)
}

func ResetPassword(c *gin.Context) {
	// Get the token from the URL parameter
	tokenParam := c.Param("token")

	// Define a struct to capture the incoming request body
	var body struct {
		Password string `json:"password" binding:"required"`
	}

	// Bind the JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Look up the token in the database
	var token models.Token
	if err := initializers.DB.First(&token, "token = ?", tokenParam).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2005,
			"error": "Invalid or expired token",
			"data":  "",
		})
		return
	}

	// Validate the new password (e.g., minimum length, special characters)
	if !validatePassword(body.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":    2002,
			"error": "Weak password. The password should be at least 8 characters long and include special characters.",
		})
		return
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// Find the user associated with the token and update their password
	var user models.User
	if err := initializers.DB.Model(&user).Where("id = ?", token.UserID).Update("password", string(hash)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reset user password",
		})
		return
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "User password updated successfully",
	})
}

func UpdateUserPassword(c *gin.Context) {
	c.Get("user")
	// Get id from request
	id := c.Param("id")

	var body struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}

	// Get contents from body of request and bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Check if the user to be updated exists
	var user models.User
	result := initializers.DB.Preload("Photos").First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	if body.NewPassword != "" && body.OldPassword != "" {
		if !validatePassword(body.NewPassword) {
			c.JSON(http.StatusNotFound, gin.H{
				"id":    2002,
				"error": "Weak password. The password should be at least 8 characters long and include special characters.",
			})
			return

		} else {
			//compare user password and oldPassword
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.OldPassword))

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"id":    2006,
					"error": "The Old Password is Invalid",
				})
				return
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Failed to hash password",
				})
				return
			}
			user.Password = string(hash)
			// Save updated user
			result = initializers.DB.Save(&user)
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
			})
		}
	}
}

func UpdateUserPreferences(c *gin.Context) {
	c.Get("user")

	id := c.Param("id")

	var body struct {
		Subscription bool   `json:"subscription"`
		Theme        string `json:"theme"`
		Language     string `json:"language"`
	}

	// Get contents from body of request and bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Check if the user to be updated exists
	var user models.User
	result := initializers.DB.Preload("Photos").First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	if body.Subscription {
		user.Subscription = body.Subscription
	}
	if body.Theme != "" {
		user.Theme = body.Theme
	}
	if body.Language != "" {
		user.Language = body.Language
	}

	// Save updated user
	result = initializers.DB.Save(&user)
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
		"data":    user,
	})
}

func GetUserPreferences(c *gin.Context) {
	c.Get("user")

	id := c.Param("id")

	// Check if the user exists
	var user models.User
	result := initializers.DB.Preload("Photos").First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"id":    2011,
			"error": "record not found",
		})
		return
	}

	var Preferences struct {
		Subscription bool
		Theme        string
		Language     string
	}

	Preferences.Subscription = user.Subscription
	Preferences.Theme = user.Theme
	Preferences.Language = user.Language

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"id":      2001,
		"message": "success",
		"data":    Preferences,
	})

}
