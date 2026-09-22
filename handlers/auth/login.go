package auth

import (
	"context"
	"net/http"
	"strings"

	"peter-go-auth-template/database"
	"peter-go-auth-template/helpers"
	"peter-go-auth-template/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Login handles POST /login — verifies credentials and issues both tokens.
func Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	users := database.Database.Collection("users")

	var user models.User
	err := users.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// Don't reveal whether the email exists or the password is wrong.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Block unverified accounts. If you'd rather let them in and gate
	// certain features later, delete this block and check EmailVerified
	// in the handlers that need it.
	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Email not verified — check your inbox for the code",
		})
		return
	}

	userIDHex := user.ID.Hex()

	accessToken, err := helpers.GenerateAccessToken(userIDHex, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := helpers.GenerateRefreshToken(userIDHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	_, err = users.UpdateOne(
		context.TODO(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"refresh_token": refreshToken}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
