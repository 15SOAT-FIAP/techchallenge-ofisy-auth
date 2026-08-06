package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
	"github.com/google/uuid"
)

type stubCustomerRepository struct {
	customer *models.Customer
	err      error
}

func (s *stubCustomerRepository) GetCustomerByCpfCnpj(ctx context.Context, cpfCnpj string) (*models.Customer, error) {
	return s.customer, s.err
}

type stubJWTGenerator struct {
	token              string
	err                error
	receivedCustomerID string
}

func (s *stubJWTGenerator) GenerateToken(customerID string) (string, error) {
	s.receivedCustomerID = customerID
	return s.token, s.err
}

func TestAuthenticate_CustomerNotFound(t *testing.T) {
	repo := &stubCustomerRepository{err: ErrCustomerNotFound}
	jwtGen := &stubJWTGenerator{}
	uc := NewAuthUseCase(repo, jwtGen)

	_, err := uc.Authenticate(context.Background(), "52998224725")

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("erro = %v, esperado ErrInvalidCredentials", err)
	}
}

func TestAuthenticate_InactiveCustomer(t *testing.T) {
	repo := &stubCustomerRepository{customer: &models.Customer{ID: uuid.New(), Active: false}}
	jwtGen := &stubJWTGenerator{}
	uc := NewAuthUseCase(repo, jwtGen)

	_, err := uc.Authenticate(context.Background(), "52998224725")

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("erro = %v, esperado ErrInvalidCredentials", err)
	}
}

func TestAuthenticate_RepositoryUnexpectedError(t *testing.T) {
	unexpected := errors.New("connection refused")
	repo := &stubCustomerRepository{err: unexpected}
	jwtGen := &stubJWTGenerator{}
	uc := NewAuthUseCase(repo, jwtGen)

	_, err := uc.Authenticate(context.Background(), "52998224725")

	if err == nil {
		t.Fatal("esperado erro, obtido nil")
	}
	if errors.Is(err, ErrInvalidCredentials) {
		t.Error("erro inesperado de infraestrutura nao deveria virar ErrInvalidCredentials")
	}
	if !errors.Is(err, unexpected) {
		t.Errorf("erro = %v, esperado que envolvesse %v", err, unexpected)
	}
}

func TestAuthenticate_TokenGenerationFails(t *testing.T) {
	repo := &stubCustomerRepository{customer: &models.Customer{ID: uuid.New(), Active: true}}
	genErr := errors.New("signing failed")
	jwtGen := &stubJWTGenerator{err: genErr}
	uc := NewAuthUseCase(repo, jwtGen)

	_, err := uc.Authenticate(context.Background(), "52998224725")

	if err == nil {
		t.Fatal("esperado erro, obtido nil")
	}
	if !errors.Is(err, genErr) {
		t.Errorf("erro = %v, esperado que envolvesse %v", err, genErr)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	customerID := uuid.New()
	repo := &stubCustomerRepository{customer: &models.Customer{ID: customerID, Active: true}}
	jwtGen := &stubJWTGenerator{token: "signed-token"}
	uc := NewAuthUseCase(repo, jwtGen)

	resp, err := uc.Authenticate(context.Background(), "52998224725")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.Token != "signed-token" {
		t.Errorf("Token = %q, esperado %q", resp.Token, "signed-token")
	}
	if jwtGen.receivedCustomerID != customerID.String() {
		t.Errorf("GenerateToken recebeu customerID = %q, esperado %q", jwtGen.receivedCustomerID, customerID.String())
	}
}
