package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Generator struct {
	secret     string
	expiration time.Duration
}

func NewGenerator(secret string, expiration time.Duration) *Generator {
	return &Generator{
		secret:     secret,
		expiration: expiration,
	}
}

func (j *Generator) GenerateToken(customerID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    "techchallenge-ofisy-auth",
		Subject:   customerID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.expiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.secret))
}
