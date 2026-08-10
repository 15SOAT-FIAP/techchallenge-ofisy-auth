.PHONY: ci pre-commit build test test-race cover-check vet fmt fmt-check lint sec tidy run \
	dev-up dev-down docker-build clean lambda-seed lambda-deploy lambda-invoke

LOCALSTACK_ENDPOINT := http://localhost:4566
LAMBDA_FUNCTION_NAME := ofisy-auth

COVERAGE_THRESHOLD := 70
COVERAGE_EXCLUDE := cmd/api/main.go|internal/auth/dto.go|internal/auth/types.go|internal/admin/dto.go

ci: tidy vet test-race build

pre-commit: tidy fmt-check vet lint test-race cover-check sec
	@echo "pre-commit checks passed"

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

test:
	@go test ./... -cover

test-race:
	@CGO_ENABLED=1 go test ./... -race -coverprofile=coverage.txt -covermode=atomic

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

docker-build:
	@docker build -t ofisy-auth-lambda .

clean:
	@rm -f bootstrap function.zip

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

lambda-invoke:
	@aws --endpoint-url=$(LOCALSTACK_ENDPOINT) lambda invoke \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--cli-binary-format raw-in-base64-out \
		--payload '{"body": "{\"cpfCnpj\": \"46808813051\"}"}' \
		--region us-east-1 \
		output.json
	@cat output.json