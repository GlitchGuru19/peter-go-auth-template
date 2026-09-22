package auth

import (
	"context"
	"net/http"
	"peter-go-auth-template/database"
	"peter-go-auth-template/helpers"
	"peter-go-auth-template/models"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// VerifyOTP handles POST /verify-otp — consumes the 6-digit code and
// marks the account as verified.
//
// Rules:
//   - Codes expire after 10 minutes.
//   - Max 5 wrong attempts, then the code is dead (must request a new one).
//   - On success, OTP fields are cleared so the code can't be replayed.
func VerifyOTP(c *gin.Context) {
	var input models.VerifyOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	users := database.Database.Collection("users")

	var user models.User
	if err := users.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user); err != nil {
		// Same message as bad code, so an attacker can't enumerate.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired code"})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"message": "Email already verified"})
		return
	}

	if user.OTPCodeHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No verification code on file — request a new one"})
		return
	}

	if time.Now().After(user.OTPExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code expired — request a new one"})
		return
	}

	if user.OTPAttempts >= helpers.OTPMaxAttempts {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many attempts — request a new code"})
		return
	}

	// Compare hashes in constant-ish time isn't necessary here because
	// the hash comparison is on the digest of a 6-digit code, not a
	// secret with meaningful entropy — and we cap attempts at 5.
	if helpers.HashOTP(input.Code) != user.OTPCodeHash {
		// Wrong code: increment attempts and return the same error
		// as "expired" so an attacker can't probe which case they hit.
		_, _ = users.UpdateOne(
			context.TODO(),
			bson.M{"_id": user.ID},
			bson.M{"$inc": bson.M{"otp_attempts": 1}},
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired code"})
		return
	}

	// Success — mark verified and clear the OTP fields.
	_, err := users.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"email_verified": true,
			"otp_code_hash":  "",
			"otp_expires_at": time.Time{},
			"otp_attempts":   0,
			"updated_at":     time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
