package initializers

import (
	"log"

	"github.com/joho/godotenv"
)

// loading all Environment variables to the project
func LoadEnvVariables() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Environment variable error")
	}
}
