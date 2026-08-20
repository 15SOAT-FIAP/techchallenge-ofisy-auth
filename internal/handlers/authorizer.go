package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const bearerPrefix = "Bearer "

// TokenValidator confere assinatura, expiração e emissor do token.
type TokenValidator interface {
	ValidateToken(token string) (customerID string, err error)
}

type AuthorizerHandler struct {
	validator TokenValidator
}

func NewAuthorizerHandler(validator TokenValidator) *AuthorizerHandler {
	return &AuthorizerHandler{validator: validator}
}

// Handle implementa um Lambda Authorizer do tipo REQUEST, com payload format
// version 2.0 (resposta simples: só true/false + contexto opcional). Nunca
// retorna erro pra fora - qualquer motivo de rejeição vira "não autorizado",
// sem detalhar o porquê pro chamador (o API Gateway só repassa isAuthorized
// adiante). O motivo real de cada rejeição fica só nos logs, via slog -
// mesmo padrão usado no AuthHandler.
func (h *AuthorizerHandler) Handle(ctx context.Context, req events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	token, ok := extractBearerToken(req.Headers)
	if !ok {
		slog.Warn("authorization denied: missing or malformed bearer token", "routeArn", req.RouteArn)
		return deny(), nil
	}

	customerID, err := h.validator.ValidateToken(token)
	if err != nil {
		slog.Warn("authorization denied: invalid token", "error", err, "routeArn", req.RouteArn)
		return deny(), nil
	}

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context: map[string]interface{}{
			"customerId": customerID,
		},
	}, nil
}

// extractBearerToken procura o header "Authorization" ignorando
// maiúsculas/minúsculas, já que a origem do map pode variar.
func extractBearerToken(headers map[string]string) (token string, ok bool) {
	for key, value := range headers {
		if !strings.EqualFold(key, "authorization") {
			continue
		}
		if !strings.HasPrefix(value, bearerPrefix) {
			return "", false
		}
		return strings.TrimPrefix(value, bearerPrefix), true
	}
	return "", false
}

func deny() events.APIGatewayV2CustomAuthorizerSimpleResponse {
	return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}
}