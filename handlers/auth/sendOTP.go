package auth

import (
	"context"
	"log"
	"net/http"
	"peter-go-auth-template/database"
	"peter-go-auth-template/helpers"
	"peter-go-auth-template/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// SendOTP handles POST /send-otp — (re)sends a fresh verification code.
//
// Security rules:
//   - Always return the same response whether or not the email exists.
//   - Don't send if the account is already verified (prevents spam).
//   - Overwrites any previous OTP, so only the latest code works.
//   - Resets the attempt counter each time a new code is issued.
func SendOTP(c *gin.Context) {
	var input models.SendOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	const generic = "If that email needs verification, a code has been sent"

	users := database.Database.Collection("users")

	var user models.User
	err := users.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// Not registered — return the same response anyway.
		c.JSON(http.StatusOK, gin.H{"message": generic})
		return
	}

	// Already verified — silently skip. Returning the same generic
	// response avoids telling an attacker which emails are verified.
	if user.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"message": generic})
		return
	}

	raw, hash, err := helpers.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
		return
	}

	_, err = users.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"otp_code_hash":  hash,
			"otp_expires_at": time.Now().Add(helpers.OTPValidFor),
			"otp_attempts":   0,
			"updated_at":     time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save code"})
		return
	}

	if err := helpers.SendOTPEmail(user.Email, raw); err != nil {
		log.Printf("[SEND-OTP] failed to email %s: %v", user.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": generic})
}
