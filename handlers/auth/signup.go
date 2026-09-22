package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"peter-go-auth-template/database"
	"peter-go-auth-template/helpers"
	"peter-go-auth-template/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Signup handles POST /signup — creates the account and emails a
// verification code. The account starts unverified; login is blocked
// until the code is consumed via /verify-otp.
func Signup(c *gin.Context) {
	var input models.SignupInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	users := database.Database.Collection("users")

	var existing models.User
	err := users.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&existing)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists."})
		return
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Generate the first verification code up front so we can store it
	// on the new user document in a single insert (avoids a follow-up update).
	rawOTP, hashOTP, err := helpers.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification code"})
		return
	}

	now := time.Now()
	newUser := models.User{
		ID:            primitive.NewObjectID(),
		Name:          input.Name,
		Email:         input.Email,
		Password:      string(hashed),
		Role:          "USER",
		EmailVerified: false,
		OTPCodeHash:   hashOTP,
		OTPExpiresAt:  now.Add(helpers.OTPValidFor),
		OTPAttempts:   0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := users.InsertOne(context.TODO(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read new user ID"})
		return
	}

	// Send the verification email. If it fails we still return 201 —
	// the account exists, and the user can hit /send-otp to retry.
	if err := helpers.SendOTPEmail(newUser.Email, rawOTP); err != nil {
		log.Printf("[SIGNUP] failed to send OTP to %s: %v", newUser.Email, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created — check your email for the verification code",
		"user_id": insertedID.Hex(),
	})
}
