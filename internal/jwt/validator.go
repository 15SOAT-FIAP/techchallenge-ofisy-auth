package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken cobre qualquer motivo de rejeição do token: assinatura
// incorreta, expirado, issuer inesperado ou malformado. De propósito não
// diferencia o motivo exato - o Lambda Authorizer só precisa saber
// autorizar ou negar, sem vazar detalhe nenhum pro chamador.
var ErrInvalidToken = errors.New("invalid token")

const expectedIssuer = "techchallenge-ofisy-auth"

type Validator struct {
	secret string
}

func NewValidator(secret string) *Validator {
	return &Validator{secret: secret}
}

// ValidateToken confere assinatura (HS256), expiração e issuer. Retorna o
// customerID (claim "sub") quando o token é válido.
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

	if claims.Issuer != expectedIssuer {
		return "", ErrInvalidToken
	}

	if claims.Subject == "" {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}
