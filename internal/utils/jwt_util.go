package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, email, role, secret string, expireMinutes int) (string, error) {
	return generateToken(userID, email, role, secret, "access", time.Duration(expireMinutes)*time.Minute)
}

func GenerateRefreshToken(userID uint, email, role, secret string) (string, error) {
	return generateToken(userID, email, role, secret, "refresh", 24*time.Hour)
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is empty")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func generateToken(userID uint, email, role, secret, tokenType string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("jwt secret is empty")
	}

	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
