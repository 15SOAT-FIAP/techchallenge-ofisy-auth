package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Validator struct {
	secret string
}

func NewValidator(secret string) *Validator {
	return &Validator{secret: secret}
}

func (v *Validator) ValidateToken(tokenString string) (customerID string, err error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(v.secret), nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	if claims.Issuer != issuer {
		return "", ErrInvalidToken
	}

	if claims.Subject == "" {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}
