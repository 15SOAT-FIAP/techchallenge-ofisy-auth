package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
)

type CustomerRepository interface {
	GetCustomerByCpfCnpj(ctx context.Context, cpfCnpj string) (*models.Customer, error)
}

type JWTGenerator interface {
	GenerateToken(customerID string) (string, error)
}

type AuthUseCase struct {
	customerRepo CustomerRepository
	jwtGenerator JWTGenerator
}

func NewAuthUseCase(customerRepo CustomerRepository, jwtGenerator JWTGenerator) *AuthUseCase {
	return &AuthUseCase{
		customerRepo: customerRepo,
		jwtGenerator: jwtGenerator,
	}
}

func (a *AuthUseCase) Authenticate(ctx context.Context, cpfCnpj string) (*models.AuthResponse, error) {
	customer, err := a.customerRepo.GetCustomerByCpfCnpj(ctx, cpfCnpj)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			slog.Warn("customer not found")
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get customer by cpf_cnpj: %w", err)
	}
	if !customer.Active {
		slog.Warn("customer is inactive")
		return nil, ErrInvalidCredentials
	}

	token, err := a.jwtGenerator.GenerateToken(customer.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token: token,
	}, nil
}
