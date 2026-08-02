package main

import (
	"log"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/config"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/database"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/repositories/customers"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/usecases"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	customerRepo := customers.New(db)
	authUseCase := usecases.NewAuthUseCase(customerRepo)
	_ = authUseCase // por enquanto nao tem handler implementado
}
