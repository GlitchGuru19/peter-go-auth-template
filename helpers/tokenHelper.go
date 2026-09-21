package helpers

import (
	"peter-go-auth-template/initializers"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateAccessToken creates a short-lived token used to authenticate
// normal API requests. It includes a "type" claim so middleware can
// distinguish it from a refresh token.
func GenerateAccessToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"type": "access",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // 24h for demo; shorten in production
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(initializers.SecretKey))
}

// GenerateRefreshToken creates a long-lived token whose only purpose is to
// be exchanged for a new access token via /refresh. It also includes a
// "type" claim for validation.
func GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(initializers.SecretKey))
}

// PrimitiveObjectIDFromHex converts a hex string (like the "sub" claim)
// back into the ObjectID type MongoDB queries expect.
func PrimitiveObjectIDFromHex(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}
