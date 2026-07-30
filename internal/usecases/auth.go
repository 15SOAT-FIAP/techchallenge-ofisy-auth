package usecases

import (
	"context"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type AuthUseCase struct {
	userRepo UserRepository
}
