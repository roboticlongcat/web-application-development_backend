package ds

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uuid.UUID `json:"user_id"`
	Scopes      []string  `json:"scopes"`
	Username    string    `json:"username"`
	IsModerator bool      `json:"is_moderator"`
}
