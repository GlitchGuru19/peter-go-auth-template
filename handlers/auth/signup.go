package auth

import (
	"context"
	"errors"
	"net/http"
	"peter-go-auth-template/database"
	"peter-go-auth-template/models"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Signup handles POST /signup – create a new user account.
func Signup(c *gin.Context) {
	var input models.SignupInput

	// Validate request body against the SignupInput struct tags.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize name and email.
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	users := database.Database.Collection("users")

	// Check if a user with this email already exists.
	var existing models.User
	err := users.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&existing)

	if err == nil {
		// Found a user -> conflict.
		c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists."})
		return
	}

	// If the error is something other than "no documents found", it's a real DB error.
	if !errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash the password.
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Build the new user document.
	newUser := models.User{
		ID:       primitive.NewObjectID(),
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
		Role:     "USER", // default role
	}

	// Insert into MongoDB.
	result, err := users.InsertOne(context.TODO(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	// Extract the inserted ID.
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read new user ID"})
		return
	}

	// Return success with the new user's ID.
	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"user_id": insertedID.Hex(),
	})
}
