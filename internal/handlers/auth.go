package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/usecases"
	"github.com/aws/aws-lambda-go/events"
)

type Authenticator interface {
	Authenticate(ctx context.Context, cpfCnpj string) (*models.AuthResponse, error)
}

type AuthHandler struct {
	authenticator Authenticator
}

func NewAuthHandler(authenticator Authenticator) *AuthHandler {
	return &AuthHandler{
		authenticator: authenticator,
	}
}

func (h *AuthHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var authRequest models.AuthRequest
	if err := json.Unmarshal([]byte(req.Body), &authRequest); err != nil {
		return errorResponse(http.StatusBadRequest, "Corpo da requisição inválido"), nil
	}

	resp, err := h.authenticator.Authenticate(ctx, authRequest.CpfCnpj)
	if err != nil {
		if errors.Is(err, usecases.ErrInvalidCredentials) {
			return errorResponse(http.StatusUnauthorized, "Credenciais de cliente inválidas"), nil
		}
		slog.Error("failed to authenticate", "error", err)
		return errorResponse(http.StatusInternalServerError, "Erro interno do servidor"), nil
	}
	body, _ := json.Marshal(resp)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func errorResponse(status int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}
