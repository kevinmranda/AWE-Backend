package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	models "github.com/kevinmranda/AWE-Backend/Models"
)

// AuthMiddleware checks for the presence of a valid JWT token in the user's cookies and validates it.
func AuthMiddleware(c *gin.Context) {
	// Retrieve the token from the "Auth" cookie
	tokenString, err := c.Cookie("Auth")
	if err != nil {
		// If the cookie is not found, return an unauthorized status
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization cookie not found",
		})
		c.Abort()
		return
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key used to sign the token
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		// Log the parsing error and return an unauthorized status
		log.Println("Failed to parse token:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		c.Abort()
		return
	}

	// Check if the token is valid and extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verify token expiration
		if exp, ok := claims["exp"].(float64); !ok || float64(time.Now().Unix()) > exp {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			c.Abort()
			return
		}

		// Retrieve the user based on the token's subject (sub)
		var user models.User
		if err := initializers.DB.First(&user, claims["sub"]).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Attach the user to the context for access in subsequent handlers
		c.Set("user", user)

		// Proceed to the next middleware or handler
		c.Next()
	} else {
		// If the token is invalid, return an unauthorized status
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
	}
}
