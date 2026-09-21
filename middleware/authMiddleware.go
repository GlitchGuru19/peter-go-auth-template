package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"auth/initializers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// RequireAuth checks for a valid JWT in the Authorization header.
// If valid, it stores the user's ID and role in the request context
// so handlers further down the chain can access who's making the request.
func RequireAuth(c *gin.Context) {
	// Standard convention: "Authorization: Bearer <token>"
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or malformed token"})
		c.Abort() // stops the request here – the handler never runs
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse and validate the token.
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(initializers.SecretKey), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	// Ensure this is an access token, not a refresh token.
	if claims["type"] != "access" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
		c.Abort()
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
		c.Abort()
		return
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token role"})
		c.Abort()
		return
	}

	// Pass the user's identity down to whatever handler runs next.
	c.Set("userID", sub)
	c.Set("userRole", role)
	c.Next()
}