package env

import (
	"os"

	"github.com/joho/godotenv"
)

// GetEnv reads environment variables from .env file.
// If file doesn't exists, it's read from system
func GetEnv(key string) string {
	godotenv.Load(".env")

	return os.Getenv(key)
}
