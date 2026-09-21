package auth

import (
	"context"
	"fmt"
	"net/http"

	"auth/database"
	"auth/helpers"
	"auth/initializers"
	"auth/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
)

// Refresh handles POST /refresh – exchanges a valid, still-active refresh
// token for a brand new access token, without requiring the password again.
func Refresh(c *gin.Context) {
	var input models.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse and validate the refresh token.
	token, err := jwt.Parse(input.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(initializers.SecretKey), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Ensure this is a refresh token, not an access token.
	if claims["type"] != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
		return
	}

	userIDHex, ok := claims["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
		return
	}

	userID, err := helpers.PrimitiveObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return
	}

	users := database.Database.Collection("users")

	var user models.User
	err = users.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Check that the presented refresh token matches the one stored on the user.
	// This allows us to revoke tokens by clearing the stored value.
	if user.RefreshToken != input.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has been revoked"})
		return
	}

	// Issue a new access token.
	newAccessToken, err := helpers.GenerateAccessToken(userIDHex, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}
