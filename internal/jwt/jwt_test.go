package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_ValidatesWithCorrectSecret(t *testing.T) {
	g := NewGenerator("my-secret", time.Hour)

	tokenString, err := g.GenerateToken("customer-123")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("my-secret"), nil
	})
	if err != nil {
		t.Fatalf("token nao validou com o secret correto: %v", err)
	}
	if !token.Valid {
		t.Fatal("token esperado valido, mas Valid = false")
	}
}

func TestGenerateToken_FailsWithWrongSecret(t *testing.T) {
	g := NewGenerator("my-secret", time.Hour)

	tokenString, err := g.GenerateToken("customer-123")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}

	_, err = jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Fatal("esperado erro ao validar token com secret errado, obtido nil")
	}
}

func TestGenerateToken_ClaimsContent(t *testing.T) {
	expiration := 2 * time.Hour
	g := NewGenerator("my-secret", expiration)

	before := time.Now()
	tokenString, err := g.GenerateToken("customer-456")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}
	after := time.Now()

	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("my-secret"), nil
	})
	if err != nil {
		t.Fatalf("falha ao parsear token: %v", err)
	}

	if claims.Subject != "customer-456" {
		t.Errorf("Subject = %q, esperado %q", claims.Subject, "customer-456")
	}
	if claims.Issuer != "techchallenge-ofisy-auth" {
		t.Errorf("Issuer = %q, esperado %q", claims.Issuer, "techchallenge-ofisy-auth")
	}

	const tolerance = 2 * time.Second
	wantExpMin := before.Add(expiration).Add(-tolerance)
	wantExpMax := after.Add(expiration).Add(tolerance)
	gotExp := claims.ExpiresAt.Time
	if gotExp.Before(wantExpMin) || gotExp.After(wantExpMax) {
		t.Errorf("ExpiresAt = %v, esperado entre %v e %v", gotExp, wantExpMin, wantExpMax)
	}
}

func TestGenerateToken_UsesHS256(t *testing.T) {
	g := NewGenerator("my-secret", time.Hour)

	tokenString, err := g.GenerateToken("customer-123")
	if err != nil {
		t.Fatalf("GenerateToken() erro inesperado: %v", err)
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		t.Fatalf("falha ao parsear token sem verificar: %v", err)
	}
	if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		t.Errorf("algoritmo de assinatura = %q, esperado %q", token.Method.Alg(), jwt.SigningMethodHS256.Alg())
	}
}
