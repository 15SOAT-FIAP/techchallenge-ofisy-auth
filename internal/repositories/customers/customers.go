package customers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/usecases"
	"github.com/jmoiron/sqlx"
)

type Customers struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Customers {
	return &Customers{db: db}
}

func (c *Customers) GetCustomerByCpfCnpj(ctx context.Context, cpfCnpj string) (*models.Customer, error) {
	var customer models.Customer
	err := c.db.GetContext(ctx, &customer, `
	SELECT id, active
	FROM customers
	WHERE cpf_cnpj = $1
	`, cpfCnpj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("customer not found", "cpf_cnpj", cpfCnpj)
			return nil, usecases.ErrCustomerNotFound
		}
		slog.Error("failed to get customer by cpf_cnpj", "cpf_cnpj", cpfCnpj, "error", err)
		return nil, fmt.Errorf("failed to get customer by cpf_cnpj: %w", err)
	}
	return &customer, nil
}
