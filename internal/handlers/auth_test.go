package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/usecases"
	"github.com/aws/aws-lambda-go/events"
)

type stubAuthenticator struct {
	resp            *models.AuthResponse
	err             error
	receivedCpfCnpj string
}

func (s *stubAuthenticator) Authenticate(ctx context.Context, cpfCnpj string) (*models.AuthResponse, error) {
	s.receivedCpfCnpj = cpfCnpj
	return s.resp, s.err
}

func decodeErrorBody(t *testing.T, body string) string {
	t.Helper()
	var payload map[string]string
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("falha ao parsear body de erro: %v", err)
	}
	msg, ok := payload["error"]
	if !ok {
		t.Fatalf("body de erro nao contem a chave \"error\": %q", body)
	}
	return msg
}

func TestHandle_MalformedBody(t *testing.T) {
	h := NewAuthHandler(&stubAuthenticator{})

	resp, err := h.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: "not-json"})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, esperado %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandle_InvalidCredentials(t *testing.T) {
	stub := &stubAuthenticator{err: usecases.ErrInvalidCredentials}
	h := NewAuthHandler(stub)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: `{"cpfCnpj":"52998224725"}`})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, esperado %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if msg := decodeErrorBody(t, resp.Body); msg != "Credenciais de cliente inválidas" {
		t.Errorf("mensagem de erro = %q, esperado %q", msg, "Credenciais de cliente inválidas")
	}
	if stub.receivedCpfCnpj != "52998224725" {
		t.Errorf("Authenticate recebeu cpfCnpj = %q, esperado %q (campo do JSON nao foi extraido corretamente)", stub.receivedCpfCnpj, "52998224725")
	}
}

func TestHandle_UnexpectedErrorDoesNotLeakDetail(t *testing.T) {
	internalErr := errors.New("connection refused: pg_hba.conf rejects connection")
	h := NewAuthHandler(&stubAuthenticator{err: internalErr})

	resp, err := h.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: `{"cpfCnpj":"52998224725"}`})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, esperado %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if strings.Contains(resp.Body, "connection refused") || strings.Contains(resp.Body, "pg_hba") {
		t.Errorf("Body vazou detalhe interno do erro: %q", resp.Body)
	}
	if msg := decodeErrorBody(t, resp.Body); msg != "Erro interno do servidor" {
		t.Errorf("mensagem de erro = %q, esperado a mensagem generica %q", msg, "Erro interno do servidor")
	}
}

func TestHandle_Success(t *testing.T) {
	stub := &stubAuthenticator{resp: &models.AuthResponse{Token: "signed-token"}}
	h := NewAuthHandler(stub)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: `{"cpfCnpj":"52998224725"}`})

	if err != nil {
		t.Fatalf("erro inesperado do handler: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, esperado %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Headers["Content-Type"]; got != "application/json" {
		t.Errorf("Content-Type = %q, esperado %q", got, "application/json")
	}

	var authResp models.AuthResponse
	if err := json.Unmarshal([]byte(resp.Body), &authResp); err != nil {
		t.Fatalf("falha ao parsear body da resposta: %v", err)
	}
	if authResp.Token != "signed-token" {
		t.Errorf("Token = %q, esperado %q", authResp.Token, "signed-token")
	}
	if stub.receivedCpfCnpj != "52998224725" {
		t.Errorf("Authenticate recebeu cpfCnpj = %q, esperado %q (campo do JSON nao foi extraido corretamente)", stub.receivedCpfCnpj, "52998224725")
	}
}
