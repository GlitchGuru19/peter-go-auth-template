package helpers

import (
	"peter-go-auth-template/initializers"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateAccessToken creates a short-lived (24h) token used to authenticate
// normal API requests.
func GenerateAccessToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"type": "access",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(initializers.SecretKey))
}

// GenerateRefreshToken creates a long-lived (30 day) token whose only purpose
// is to be exchanged for a new access token via /refresh.
func GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(initializers.SecretKey))
}

// GeneratePasswordResetToken creates a 15-minute JWT used only for the
// forgot-password flow. The "type": "reset" claim ensures it can't be
// used as an access or refresh token.
func GeneratePasswordResetToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "reset",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(initializers.SecretKey))
}

// PrimitiveObjectIDFromHex converts a hex string (like the "sub" claim)
// back into the ObjectID type MongoDB queries expect.
func PrimitiveObjectIDFromHex(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}
