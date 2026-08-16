package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type stubValidator struct {
	customerID    string
	err           error
	receivedToken string
}

func (s *stubValidator) ValidateToken(token string) (string, error) {
	s.receivedToken = token
	return s.customerID, s.err
}

func TestHandle_MissingAuthorizationHeader(t *testing.T) {
	h := NewAuthorizerHandler(&stubValidator{})

	resp, err := h.Handle(context.Background(), events.APIGatewayV2CustomAuthorizerV2Request{
		Headers: map[string]string{},
	})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.IsAuthorized {
		t.Error("IsAuthorized = true, esperado false (sem header Authorization)")
	}
}

func TestHandle_HeaderWithoutBearerPrefix(t *testing.T) {
	h := NewAuthorizerHandler(&stubValidator{})

	resp, err := h.Handle(context.Background(), events.APIGatewayV2CustomAuthorizerV2Request{
		Headers: map[string]string{"authorization": "token-sem-prefixo"},
	})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.IsAuthorized {
		t.Error("IsAuthorized = true, esperado false (sem prefixo Bearer)")
	}
}

func TestHandle_InvalidToken(t *testing.T) {
	stub := &stubValidator{err: errors.New("invalid token")}
	h := NewAuthorizerHandler(stub)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2CustomAuthorizerV2Request{
		Headers: map[string]string{"authorization": "Bearer token-invalido"},
	})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.IsAuthorized {
		t.Error("IsAuthorized = true, esperado false (token invalido)")
	}
	if stub.receivedToken != "token-invalido" {
		t.Errorf("token recebido pelo validador = %q, esperado %q", stub.receivedToken, "token-invalido")
	}
}

func TestHandle_ValidToken(t *testing.T) {
	stub := &stubValidator{customerID: "customer-123"}
	h := NewAuthorizerHandler(stub)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2CustomAuthorizerV2Request{
		Headers: map[string]string{"Authorization": "Bearer token-valido"},
	})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if !resp.IsAuthorized {
		t.Fatal("IsAuthorized = false, esperado true (token valido)")
	}
	if got := resp.Context["customerId"]; got != "customer-123" {
		t.Errorf("Context[customerId] = %v, esperado %q", got, "customer-123")
	}
	if stub.receivedToken != "token-valido" {
		t.Errorf("token recebido pelo validador = %q, esperado %q", stub.receivedToken, "token-valido")
	}
}

func TestHandle_HeaderLookupIsCaseInsensitive(t *testing.T) {
	stub := &stubValidator{customerID: "customer-123"}
	h := NewAuthorizerHandler(stub)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2CustomAuthorizerV2Request{
		Headers: map[string]string{"AUTHORIZATION": "Bearer token-valido"},
	})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if !resp.IsAuthorized {
		t.Error("IsAuthorized = false, esperado true (header em caixa alta deveria ser aceito)")
	}
}
