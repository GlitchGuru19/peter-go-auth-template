// Package models defines the data structure stored in the database and provides methods to interact with it.
// and used to validate incoming requests.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents an account stored in the "users" collection.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name" binding:"required"`
	Email        string             `bson:"email" json:"email" binding:"required,email"`
	Password     string             `bson:"password" json:"-"` // json:"-" never send the hash back to the client.
	Role         string             `bson:"role" json:"role"`
	RefreshToken string             `bson:"refresh_token" json:"-"` // never expose

	// Password reset fields. Cleared after a successful reset.
	// We store the raw token here for simplicity, matching how the
	// existing RefreshToken field works. When you move to the full
	// production pattern, these move to their own collection with
	// hashed values.
	PasswordResetToken     string    `bson:"password_reset_token" json:"-"`
	PasswordResetExpiresAt time.Time `bson:"password_reset_expires_at" json:"-"`
}

// SignupInput is what we expect in the request body when creating an account.
type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginInput is what we expect in the request body when logging in.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshInput is what we expect in the request body when refreshing a token.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordInput is the body of POST /forgot-password.
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordInput is the body of POST /reset-password.
type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}