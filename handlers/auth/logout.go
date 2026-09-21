package auth

import (
	"context"
	"net/http"
	"peter-go-auth-template/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Logout handles POST /logout – clears the stored refresh token, so it
// can no longer be exchanged for new access tokens even though the JWT
// itself hasn't technically expired yet. Requires RequireAuth to have run
// first, so we know who "userID" is.
func Logout(c *gin.Context) {
	// Safely retrieve the userID from context.
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userIDHex, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	users := database.Database.Collection("users")
	_, err = users.UpdateOne(
		context.TODO(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"refresh_token": ""}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log out"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
