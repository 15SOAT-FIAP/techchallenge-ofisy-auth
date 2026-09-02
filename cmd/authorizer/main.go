package main

import (
	"log"
	"os"

	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/handlers"
	"github.com/15SOAT-FIAP/techchallenge-ofisy-auth/internal/jwt"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("missing required env var: JWT_SECRET")
	}

	validator := jwt.NewValidator(secret)
	authorizerHandler := handlers.NewAuthorizerHandler(validator)

	lambda.Start(authorizerHandler.Handle)
}
