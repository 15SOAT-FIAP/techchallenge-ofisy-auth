//go:build integration

package customers

import (
	"context"
	"errors"
	"testing"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/usecases"
	"github.com/google/uuid"
)

const inactiveCustomerID = "a1b2c3d4-e5f6-7890-abcd-ef1234567899"

func TestGetCustomerByCpfCnpj_Found(t *testing.T) {
	insertInactiveCustomer(t)
	repo := New(db)

	tests := []struct {
		name       string
		cpfCnpj    string
		wantID     string
		wantActive bool
	}{
		{
			name:       "cpf de cliente ativo",
			cpfCnpj:    "46808813051",
			wantID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567801",
			wantActive: true,
		},
		{
			name:       "cnpj de cliente ativo",
			cpfCnpj:    "14986024000176",
			wantID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567808",
			wantActive: true,
		},
		{
			name:       "cliente inativo retorna sem erro",
			cpfCnpj:    "63764546076",
			wantID:     inactiveCustomerID,
			wantActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := repo.GetCustomerByCpfCnpj(context.Background(), tt.cpfCnpj)
			if err != nil {
				t.Fatalf("erro = %v, esperado nil", err)
			}
			if customer.ID != uuid.MustParse(tt.wantID) {
				t.Errorf("id = %v, esperado %v", customer.ID, tt.wantID)
			}
			if customer.Active != tt.wantActive {
				t.Errorf("active = %v, esperado %v", customer.Active, tt.wantActive)
			}
		})
	}
}

func TestGetCustomerByCpfCnpj_NotFound(t *testing.T) {
	repo := New(db)

	tests := []struct {
		name    string
		cpfCnpj string
	}{
		{name: "cpf inexistente", cpfCnpj: "00000000000"},
		{name: "string vazia", cpfCnpj: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := repo.GetCustomerByCpfCnpj(context.Background(), tt.cpfCnpj)
			if !errors.Is(err, usecases.ErrCustomerNotFound) {
				t.Errorf("erro = %v, esperado ErrCustomerNotFound", err)
			}
			if customer != nil {
				t.Errorf("customer = %v, esperado nil", customer)
			}
		})
	}
}

func TestGetCustomerByCpfCnpj_ContextCancelado(t *testing.T) {
	repo := New(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	customer, err := repo.GetCustomerByCpfCnpj(ctx, "46808813051")
	if err == nil {
		t.Fatal("erro = nil, esperado falha por contexto cancelado")
	}
	if errors.Is(err, usecases.ErrCustomerNotFound) {
		t.Errorf("erro = ErrCustomerNotFound, esperado erro de infraestrutura")
	}
	if customer != nil {
		t.Errorf("customer = %v, esperado nil", customer)
	}
}

func insertInactiveCustomer(t *testing.T) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
	INSERT INTO customers (id, cpf_cnpj, name, email, phone, active)
	VALUES ($1, '63764546076', 'Cliente Inativo', 'inativo@email.com', '11900000000', false)
	`, inactiveCustomerID)
	if err != nil {
		t.Fatalf("falha ao inserir cliente inativo: %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM customers WHERE id = $1`, inactiveCustomerID); err != nil {
			t.Errorf("falha ao remover cliente inativo: %v", err)
		}
	})
}
