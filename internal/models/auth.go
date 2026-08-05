package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Customer struct {
	ID     uuid.UUID
	Active bool
}

type AuthRequest struct {
	CpfCnpj string `json:"cpfCnpj"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type Claims struct {
	jwt.RegisteredClaims
}
