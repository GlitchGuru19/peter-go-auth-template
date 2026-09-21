package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole returns a middleware that only allows requests through
// if the authenticated user's role matches one of the allowed roles.
// It must run AFTER RequireAuth, since it depends on "userRole" already
// being set in the context.
//
// Usage: router.GET("/admin", middleware.RequireAuth, middleware.RequireRole("ADMIN"), handler)
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	// We return a closure (a function that remembers allowedRoles) so this
	// one function can be reused for different roles on different routes.
	return func(c *gin.Context) {
		// Safely retrieve the user's role from context.
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		// Check if the user's role is in the list of allowed roles.
		for _, allowed := range allowedRoles {
			if roleStr == allowed {
				c.Next() // role matches – let the request continue
				return
			}
		}

		// If we reach here, no match was found – reject.
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
		c.Abort()
	}
}