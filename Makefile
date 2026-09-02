.PHONY: help ci pre-commit build build-authorizer test test-race test-integration cover-check vet fmt fmt-check lint sec tidy run \
	dev-up dev-down docker-build docker-build-authorizer clean lambda-seed lambda-deploy lambda-deploy-authorizer lambda-invoke lambda-invoke-authorizer sonar sonar-up sonar-down sonar-logs

.DEFAULT_GOAL := help

LOCALSTACK_ENDPOINT := http://localhost:4566
LAMBDA_FUNCTION_NAME := ofisy-auth
AUTHORIZER_FUNCTION_NAME := ofisy-auth-authorizer
SONAR_COMPOSE_FILE := compose.sonar.yaml
SONAR_HOST_URL := http://host.docker.internal:9000
SONAR_SCANNER_IMAGE := sonarsource/sonar-scanner-cli:latest
COVERAGE_THRESHOLD := 70
COVERAGE_EXCLUDE := /cmd/|/internal/models/

help: ## Lista os alvos disponíveis
	@echo "Uso: make <alvo>"
	@awk 'BEGIN {FS = ":.*##"} \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next } \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

##@ Verificação

ci: tidy vet test-race build build-authorizer ## Roda a verificação da pipeline de CI

pre-commit: tidy fmt-check vet lint test-race cover-check sec ## Roda todas as checagens antes de commitar (exige Docker)
	@echo "pre-commit checks passed"

##@ Build

build: ## Compila o binário da Lambda de autenticação (linux/arm64)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

build-authorizer: ## Compila o binário da Lambda authorizer (linux/arm64)
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap-authorizer ./cmd/authorizer

##@ Testes

test: ## Roda os testes unitários com cobertura (rápido, sem Docker)
	@go test ./... -cover

test-race: ## Roda todos os testes com -race e gera o coverage.txt (exige Docker)
	@CGO_ENABLED=1 go test -tags=integration ./... -race -coverprofile=coverage.txt -covermode=atomic

test-integration: ## Roda os testes de integração do repositório com Testcontainers (exige Docker)
	@go test -tags=integration ./internal/repositories/... -count=1 -v

cover-check: test-race ## Verifica se a cobertura atinge o mínimo exigido (exige Docker)
	@grep -Ev "$(COVERAGE_EXCLUDE)" coverage.txt > coverage.filtered.txt
	@COVERAGE=$$(go tool cover -func=coverage.filtered.txt | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if awk "BEGIN {exit !($$COVERAGE < $(COVERAGE_THRESHOLD))}"; then \
		echo "Coverage $${COVERAGE}% is below the $(COVERAGE_THRESHOLD)% threshold"; \
		exit 1; \
	fi

##@ Qualidade de código

vet: ## Roda o go vet em todos os pacotes
	@go vet ./...

fmt: ## Formata os arquivos com gofmt
	@gofmt -w .

fmt-check: ## Verifica se há arquivos fora do padrão do gofmt
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted (run 'make fmt'):"; \
		gofmt -l .; \
		exit 1; \
	fi

lint: ## Roda o golangci-lint
	@golangci-lint run ./...

sec: ## Roda a análise de segurança com gosec
	@gosec -quiet ./...

tidy: ## Sincroniza as dependências do go.mod
	@go mod tidy

##@ Ambiente local

run: ## Executa a Lambda de autenticação localmente
	@go run ./cmd/lambda

dev-up: ## Sobe o LocalStack e o Postgres local
	@docker compose up -d

dev-down: ## Derruba o LocalStack e o Postgres local
	@docker compose down

##@ Análise estática (SonarQube)

sonar-up: ## Sobe o SonarQube em http://localhost:9000
	@docker compose -f $(SONAR_COMPOSE_FILE) up -d

sonar-down: ## Derruba o SonarQube
	@docker compose -f $(SONAR_COMPOSE_FILE) down

sonar-logs: ## Acompanha os logs do SonarQube
	@docker compose -f $(SONAR_COMPOSE_FILE) logs -f sonarqube

sonar: cover-check ## Envia o código e a cobertura para o SonarQube
	@set -a && . ./.env && set +a && \
	if [ -z "$$SONAR_TOKEN" ]; then \
		echo "SONAR_TOKEN não definido. Suba o SonarQube com 'make sonar-up', gere um token em My Account > Security e defina SONAR_TOKEN no .env"; \
		exit 1; \
	fi; \
	docker run --rm \
		--add-host=host.docker.internal:host-gateway \
		-e SONAR_HOST_URL=$(SONAR_HOST_URL) \
		-e SONAR_TOKEN="$$SONAR_TOKEN" \
		-v "$(PWD):/usr/src" \
		$(SONAR_SCANNER_IMAGE)

##@ Docker

docker-build: ## Builda a imagem da Lambda de autenticação
	@docker build -t ofisy-auth-lambda .

docker-build-authorizer: ## Builda a imagem da Lambda authorizer
	@docker build -f Dockerfile.authorizer -t ofisy-auth-authorizer .

clean: ## Remove os binários e pacotes gerados
	@rm -f bootstrap bootstrap-authorizer function.zip function-authorizer.zip

##@ Lambda no LocalStack

lambda-seed: ## Popula o Postgres local com os dados do scripts/seed.sql
	@set -a && . ./.env && set +a && \
	docker exec -i postgres-auth-local-db psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" < scripts/seed.sql

lambda-deploy: build ## Empacota e registra a Lambda de autenticação no LocalStack
	@zip -j function.zip bootstrap
	@set -a && . ./.env && set +a && \
	aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda create-function \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--runtime provided.al2023 \
		--architectures arm64 \
		--handler bootstrap \
		--role arn:aws:iam::000000000000:role/lambda-role \
		--zip-file fileb://function.zip \
		--timeout 30 \
		--environment "Variables={POSTGRES_HOST=postgres,POSTGRES_PORT=5432,POSTGRES_DB=$$POSTGRES_DB,POSTGRES_USER=$$POSTGRES_USER,POSTGRES_PASSWORD=$$POSTGRES_PASSWORD,POSTGRES_SSL_MODE=disable,JWT_SECRET=$$JWT_SECRET,JWT_EXPIRATION=$$JWT_EXPIRATION}" \
		--region us-east-1 \
	|| aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda update-function-code \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--zip-file fileb://function.zip \
		--region us-east-1

lambda-deploy-authorizer: build-authorizer ## Empacota e registra a Lambda authorizer no LocalStack
	@zip -j function-authorizer.zip bootstrap-authorizer
	@set -a && . ./.env && set +a && \
	aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda create-function \
		--function-name $(AUTHORIZER_FUNCTION_NAME) \
		--runtime provided.al2023 \
		--architectures arm64 \
		--handler bootstrap-authorizer \
		--role arn:aws:iam::000000000000:role/lambda-role \
		--zip-file fileb://function-authorizer.zip \
		--timeout 10 \
		--environment "Variables={JWT_SECRET=$$JWT_SECRET}" \
		--region us-east-1 \
	|| aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda update-function-code \
		--function-name $(AUTHORIZER_FUNCTION_NAME) \
		--zip-file fileb://function-authorizer.zip \
		--region us-east-1

lambda-invoke: ## Invoca a Lambda de autenticação com um payload de exemplo
	@aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda invoke \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--cli-binary-format raw-in-base64-out \
		--payload '{"body": "{\"cpfCnpj\": \"46808813051\"}"}' \
		--region us-east-1 \
		output.json
	@cat output.json

lambda-invoke-authorizer: ## Invoca a Lambda authorizer com um token de exemplo
	@aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda invoke \
		--function-name $(AUTHORIZER_FUNCTION_NAME) \
		--cli-binary-format raw-in-base64-out \
		--payload '{"headers": {"authorization": "Bearer token-de-teste"}}' \
		--region us-east-1 \
		output-authorizer.json
	@cat output-authorizer.json