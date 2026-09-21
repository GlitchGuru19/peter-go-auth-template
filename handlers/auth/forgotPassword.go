package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"peter-go-auth-template/database"
	"peter-go-auth-template/helpers"
	"peter-go-auth-template/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// ForgotPassword handles POST /forgot-password.
//
// Security rules:
//   - Always return the same 200 response whether or not the email exists.
//     Any difference lets attackers enumerate registered emails.
//   - Short expiry (15 min).
//   - The token is invalidated as soon as it's used (reset handler clears it).
func ForgotPassword(c *gin.Context) {
	var input models.ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Same response in every non-error case so we don't leak which
	// emails are registered.
	const generic = "If that email is registered, a reset link has been sent"

	users := database.Database.Collection("users")

	var user models.User
	err := users.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// Not registered — return the same response anyway. No email sent.
		c.JSON(http.StatusOK, gin.H{"message": generic})
		return
	}

	token, err := helpers.GeneratePasswordResetToken(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	// Store the token on the user document with a 15-minute expiry.
	// Overwrites any previous reset token, so only the latest link works.
	_, err = users.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"password_reset_token":      token,
			"password_reset_expires_at": time.Now().Add(15 * time.Minute),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
		return
	}

	// Fire the email. If it fails, we log it but still return the generic
	// success — telling the client "email send failed" would leak that the
	// account exists.
	if err := helpers.SendPasswordResetEmail(user.Email, token); err != nil {
		log.Printf("[FORGOT-PASSWORD] failed to send email to %s: %v", user.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": generic})
}
