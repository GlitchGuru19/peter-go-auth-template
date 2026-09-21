package main

import (
	"peter-go-auth-template/database"
	"peter-go-auth-template/initializers"
	"peter-go-auth-template/routes"

	"github.com/gin-gonic/gin"
)

// init runs automatically before main(), every time the program starts.
// We use it here to load environment variables before anything else touches them.
func init() {
	initializers.LoadEnvVariables()
	database.ConnectToDB(initializers.MongoURI)
}
func main() {
	router := gin.Default()
	routes.SetupRoutes(router)
	router.Run(":" + initializers.Port)
}
