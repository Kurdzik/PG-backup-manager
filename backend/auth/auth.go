package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	jwt.RegisteredClaims
}
