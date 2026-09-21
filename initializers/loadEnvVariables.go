// Package initializers contains setup logic that needs to run
// before the rest of the application starts (env vars, DB connections, etc).
package initializers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Global variables holding our loaded environment values.
// Other packages import these directly (e.g. initializers.Port)
// instead of calling os.Getenv everywhere.
var (
	Port      string
	MongoURI  string
	SecretKey string
)

// LoadEnvVariables reads the .env file and populates the package-level
// variables above. It should be called once, before anything else runs.
func LoadEnvVariables() {
	log.Println("Loading environment variables...")
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file.")
	}

	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "8000" // Default port if not set
	}
	MongoURI = os.Getenv("MONGODB_URI")
	if MongoURI == "" {
		log.Fatal("MONGODB_URI is not set in the environment variables.")
	}
	SecretKey = os.Getenv("SECRET_KEY")
	if SecretKey == "" {
		log.Fatal("SECRET_KEY is not set in the environment variables.")
	}

	log.Println("Environment variables loaded successfully")
}
