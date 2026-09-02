//go:build integration

package customers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	code, err := runWithPostgres(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "falha ao preparar o postgres de teste: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runWithPostgres(m *testing.M) (int, error) {
	ctx := context.Background()

	seed, err := filepath.Abs(filepath.Join("..", "..", "..", "scripts", "seed.sql"))
	if err != nil {
		return 0, fmt.Errorf("failed to resolve seed script path: %w", err)
	}

	container, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("ofisy_auth_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(seed),
		postgres.BasicWaitStrategies(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintf(os.Stderr, "falha ao encerrar o container: %v\n", err)
		}
	}()
	if err != nil {
		return 0, fmt.Errorf("failed to start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return 0, fmt.Errorf("failed to build connection string: %w", err)
	}

	db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		return 0, fmt.Errorf("failed to open db connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return 0, fmt.Errorf("failed to ping db: %w", err)
	}

	return m.Run(), nil
}
