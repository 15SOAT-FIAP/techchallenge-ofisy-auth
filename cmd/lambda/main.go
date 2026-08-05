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
		log.Fatalf("postgres: failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("postgres: failed to close database connection: %v", err)
		}
	}()

	customerRepo := customers.New(db)
	authUseCase := usecases.NewAuthUseCase(customerRepo)
	_ = authUseCase // por enquanto nao tem handler implementado
}
