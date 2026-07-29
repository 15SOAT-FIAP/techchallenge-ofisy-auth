package database

import (
	"context"
	"fmt"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/config"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresConnection(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", cfg.BuildDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	// Manter um pool pequeno pelo fato de ser uma lambda
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}
