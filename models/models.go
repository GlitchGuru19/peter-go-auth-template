// Package models defines the data structure stored in the database and provides methods to interact with it.
// and useed to validate incoming requests.
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User represents an account stored in the "users" collection.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name" binding:"required"`
	Email        string             `bson:"email" json:"email" binding:"required,email"`
	Password     string             `bson:"password" json:"-"` // json :"-" never send the hash back to the client.
	Role         string             `bson:"role" json:"role"`
	RefreshToken string             `bson:"refresh_token" json:"-"` // never expose
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

// RefreshToken represents a stored refresh token linked to a user.
type RefreshToken struct {
	Token     string             `bson:"token" json:"token"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ExpiresAt int64              `bson:"expires_at" json:"expires_at"`
	Revoked   bool               `bson:"revoked" json:"revoked"`
}

// RefreshInput is what we expect in the request body when refreshing a token.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
