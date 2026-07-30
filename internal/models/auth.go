package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleAttendant Role = "ATTENDANT"
	RoleMechanic  Role = "MECHANIC"
	RoleStockman  Role = "STOCKMAN"
)

type User struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
	Role     Role
	Active   bool
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Claims struct {
	jwt.RegisteredClaims
}
