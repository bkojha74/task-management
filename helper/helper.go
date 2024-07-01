// helper.go
// Author: Bipin Kumar Ojha (Freelancer)

package helper

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file located in the specified directory.
// This function uses the godotenv package to read the .env file and set the environment variables.
// If there is an error loading the .env file, the function panics with an appropriate error message.
//
// Parameters:
// - currentConfigDirectory: The directory where the .env file is located.
func LoadEnv(currentConfigDirectory string) {
	err := godotenv.Load(currentConfigDirectory + "/.env")
	if err != nil {
		panic("Error loading .env file")
	}
}

// GetEnv retrieves the value of the environment variable named by the key.
// It uses the os package to get the value of the specified environment variable.
// If the key is not found, an empty string is returned.
//
// Parameters:
// - key: The name of the environment variable to retrieve.
//
// Returns:
// - string: The value of the environment variable, or an empty string if the key is not found.
func GetEnv(key string) string {
	return os.Getenv(key)
}
