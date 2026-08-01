package customers

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
	"github.com/jmoiron/sqlx"
)

type Customers struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Customers {
	return &Customers{db: db}
}

func (u *Customers) GetCustomerByCpfCnpj(ctx context.Context, cpfCnpj string) (*models.Customer, error) {
	var customer models.Customer
	err := u.db.GetContext(ctx, &customer, `
	SELECT id, active
	FROM customers
	WHERE cpf_cnpj = $1
	`, cpfCnpj)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("customer not found", "cpf_cnpj", cpfCnpj)
			return nil, sql.ErrNoRows
		}
		slog.Error("failed to get customer by cpf_cnpj", "cpf_cnpj", cpfCnpj, "error", err)
		return nil, err
	}
	return &customer, nil
}
