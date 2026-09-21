package routes

import (
	"net/http"
	"peter-go-auth-template/handlers/auth"
	"peter-go-auth-template/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Public, rate-limited.
	router.POST("/signup", middleware.RateLimit(), auth.Signup)
	router.POST("/login", middleware.RateLimit(), auth.Login)
	router.POST("/refresh", middleware.RateLimit(), auth.Refresh)
	router.POST("/forgot-password", middleware.RateLimit(), auth.ForgotPassword)
	router.POST("/reset-password", middleware.RateLimit(), auth.ResetPassword)

	// Protected.
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