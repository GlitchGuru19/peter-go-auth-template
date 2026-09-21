package routes

import (
	"net/http"

	"auth/handlers/auth"
	"auth/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes registers all routes on the given Gin engine.
func SetupRoutes(router *gin.Engine) {
	// Public, but rate-limited — these are the endpoints someone would
	// try to brute-force (guessing passwords, spamming signups).
	router.POST("/signup", middleware.RateLimit(), auth.Signup)
	router.POST("/login", middleware.RateLimit(), auth.Login)
	router.POST("/refresh", middleware.RateLimit(), auth.Refresh)

	// Protected — requires a valid access token.
	router.GET("/me", middleware.RequireAuth, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.MustGet("userID"),
			"role":   c.MustGet("userRole"),
		})
	})
	router.POST("/logout", middleware.RequireAuth, auth.Logout)

	// Admin-only.
	router.GET("/admin/users", middleware.RequireAuth, middleware.RequireRole("ADMIN"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome, admin — this route is role-protected"})
	})
}