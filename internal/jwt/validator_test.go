package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateToken_AcceptsTokenSignedWithSameSecret(t *testing.T) {
	g := NewGenerator("my-secret", time.Hour)
	v := NewValidator("my-secret")

	tokenString, err := g.GenerateToken("customer-123")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}

	customerID, err := v.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken() erro inesperado: %v", err)
	}
	if customerID != "customer-123" {
		t.Errorf("customerID = %q, esperado %q", customerID, "customer-123")
	}
}

func TestValidateToken_RejectsTokenSignedWithDifferentSecret(t *testing.T) {
	g := NewGenerator("secret-a", time.Hour)
	v := NewValidator("secret-b")

	tokenString, err := g.GenerateToken("customer-123")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}

	if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
		t.Errorf("err = %v, esperado %v", err, ErrInvalidToken)
	}
}

func TestValidateToken_RejectsExpiredToken(t *testing.T) {
	g := NewGenerator("my-secret", -time.Minute)
	v := NewValidator("my-secret")

	tokenString, err := g.GenerateToken("customer-123")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}

	if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
		t.Errorf("err = %v, esperado %v", err, ErrInvalidToken)
	}
}

func TestValidateToken_RejectsUnexpectedIssuer(t *testing.T) {
	v := NewValidator("my-secret")

	claims := jwt.RegisteredClaims{
		Issuer:    "outro-emissor",
		Subject:   "customer-123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("my-secret"))
	if err != nil {
		t.Fatalf("falha ao gerar token de teste: %v", err)
	}

	if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
		t.Errorf("err = %v, esperado %v", err, ErrInvalidToken)
	}
}

func TestValidateToken_RejectsUnexpectedSigningAlgorithm(t *testing.T) {
	v := NewValidator("my-secret")

	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   "customer-123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("falha ao gerar token de teste: %v", err)
	}

	if _, err := v.ValidateToken(tokenString); err != ErrInvalidToken {
		t.Errorf("err = %v, esperado %v", err, ErrInvalidToken)
	}
}

func TestValidateToken_RejectsMalformedToken(t *testing.T) {
	v := NewValidator("my-secret")

	if _, err := v.ValidateToken("isso-nao-e-um-jwt"); err != ErrInvalidToken {
		t.Errorf("err = %v, esperado %v", err, ErrInvalidToken)
	}
}
