package users

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/models"
	"github.com/jmoiron/sqlx"
)

type Users struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Users {
	return &Users{db: db}
}

func (u *Users) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := u.db.GetContext(ctx, &user, `
	SELECT id, name, email, password, role, active
	FROM users
	WHERE email = $1
	`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("email not found", "email", email)
			return nil, sql.ErrNoRows
		}
		slog.Error("failed to get user by email", "email", email, "error", err)
		return nil, err
	}
	return &user, nil
}
