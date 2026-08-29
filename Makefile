.PHONY: ci pre-commit build build-authorizer test test-race test-integration cover-check vet fmt fmt-check lint sec tidy run \
	dev-up dev-down docker-build docker-build-authorizer clean lambda-seed lambda-deploy lambda-deploy-authorizer lambda-invoke lambda-invoke-authorizer sonar sonar-up sonar-down sonar-logs

LOCALSTACK_ENDPOINT := http://localhost:4566
LAMBDA_FUNCTION_NAME := ofisy-auth
AUTHORIZER_FUNCTION_NAME := ofisy-auth-authorizer
SONAR_COMPOSE_FILE := compose.sonar.yaml
SONAR_HOST_URL := http://host.docker.internal:9000
SONAR_SCANNER_IMAGE := sonarsource/sonar-scanner-cli:latest
COVERAGE_THRESHOLD := 70
COVERAGE_EXCLUDE := /cmd/|/internal/models/

ci: tidy vet test-race build build-authorizer

pre-commit: tidy fmt-check vet lint test-race cover-check sec
	@echo "pre-commit checks passed"

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

build-authorizer:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap-authorizer ./cmd/authorizer

test:
	@go test ./... -cover

test-race:
	@CGO_ENABLED=1 go test -tags=integration ./... -race -coverprofile=coverage.txt -covermode=atomic

test-integration:
	@go test -tags=integration ./internal/repositories/... -count=1 -v

cover-check: test-race
	@grep -Ev "$(COVERAGE_EXCLUDE)" coverage.txt > coverage.filtered.txt
	@COVERAGE=$$(go tool cover -func=coverage.filtered.txt | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if awk "BEGIN {exit !($$COVERAGE < $(COVERAGE_THRESHOLD))}"; then \
		echo "Coverage $${COVERAGE}% is below the $(COVERAGE_THRESHOLD)% threshold"; \
		exit 1; \
	fi

vet:
	@go vet ./...

fmt:
	@gofmt -w .

fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted (run 'make fmt'):"; \
		gofmt -l .; \
		exit 1; \
	fi

lint:
	@golangci-lint run ./...

sec:
	@gosec -quiet ./...

tidy:
	@go mod tidy

run:
	@go run ./cmd/lambda

dev-up:
	@docker compose up -d

dev-down:
	@docker compose down

sonar-up:
	@docker compose -f $(SONAR_COMPOSE_FILE) up -d

sonar-down:
	@docker compose -f $(SONAR_COMPOSE_FILE) down

sonar-logs:
	@docker compose -f $(SONAR_COMPOSE_FILE) logs -f sonarqube

sonar: cover-check
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

docker-build:
	@docker build -t ofisy-auth-lambda .

docker-build-authorizer:
	@docker build -f Dockerfile.authorizer -t ofisy-auth-authorizer .

clean:
	@rm -f bootstrap bootstrap-authorizer function.zip function-authorizer.zip

lambda-seed:
	@set -a && . ./.env && set +a && \
	docker exec -i postgres-auth-local-db psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" < scripts/seed.sql

lambda-deploy: build
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

lambda-deploy-authorizer: build-authorizer
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

lambda-invoke:
	@aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda invoke \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--cli-binary-format raw-in-base64-out \
		--payload '{"body": "{\"cpfCnpj\": \"46808813051\"}"}' \
		--region us-east-1 \
		output.json
	@cat output.json

lambda-invoke-authorizer:
	@aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda invoke \
		--function-name $(AUTHORIZER_FUNCTION_NAME) \
		--cli-binary-format raw-in-base64-out \
		--payload '{"headers": {"authorization": "Bearer token-de-teste"}}' \
		--region us-east-1 \
		output-authorizer.json
	@cat output-authorizer.json