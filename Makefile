.PHONY: ci build test test-race vet tidy run db-up db-down docker-build clean

ci: tidy vet test-race build

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

test:
	@go test ./... -cover

test-race:
	@go test ./... -race -cover

vet:
	@go vet ./...

tidy:
	@go mod tidy

run:
	@go run ./cmd/lambda

db-up:
	@docker compose -f compose.db.yaml up -d

db-down:
	@docker compose -f compose.db.yaml down

docker-build:
	@docker build -t ofisy-auth-lambda .

clean:
	@rm -f bootstrap