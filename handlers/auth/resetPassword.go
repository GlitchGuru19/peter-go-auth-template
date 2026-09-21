package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"peter-go-auth-template/database"
	"peter-go-auth-template/initializers"
	"peter-go-auth-template/models"
)

// ResetPassword handles POST /reset-password.
//
// Flow:
//   1. Verify the JWT is validly signed and not expired.
//   2. Confirm the claim type is "reset" (not access/refresh).
//   3. Look up the user and check the token matches what's stored.
//   4. Hash the new password, update the user.
//   5. Clear the reset fields so the token can't be replayed.
func ResetPassword(c *gin.Context) {
	var input models.ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify signature and expiry.
	token, err := jwt.Parse(input.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(initializers.SecretKey), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		return
	}

	// Only accept reset tokens here — an access or refresh token must not work.
	if claims["type"] != "reset" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		return
	}

	userIDHex, _ := claims["sub"].(string)
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		return
	}

	users := database.Database.Collection("users")

	var user models.User
	if err := users.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		return
	}

	// The token must be the exact one we stored. This makes the flow
	// single-use: after a successful reset we clear the field, and any
	// further attempt with the same token fails this check.
	if user.PasswordResetToken != input.Token {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token already used or invalid"})
		return
	}
	if time.Now().After(user.PasswordResetExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token expired"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	_, err = users.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"password":                  string(hashed),
			"password_reset_token":      "",
			"password_reset_expires_at": time.Time{},
			"updated_at":                time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}
