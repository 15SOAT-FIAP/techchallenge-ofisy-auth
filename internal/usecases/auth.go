package usecases

import (
	"context"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
)

type CustomerRepository interface {
	GetCustomerByCpfCnpj(ctx context.Context, cpfCnpj string) (*models.Customer, error)
}

type AuthUseCase struct {
	customerRepo CustomerRepository
}

func NewAuthUseCase(customerRepo CustomerRepository) *AuthUseCase {
	return &AuthUseCase{
		customerRepo: customerRepo,
	}
}
