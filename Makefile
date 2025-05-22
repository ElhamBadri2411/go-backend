include .env

GO := /usr/local/go/bin/go
MIGRATIONS_PATH = ./cmd/migrate/migrations/
SEED_PATH = ./cmd/migrate/seed/main.go

.PHONY: migrate
migrate:
	@goose create -dir $(MIGRATIONS_PATH) $(name) sql

.PHONY: up
up:
	@goose postgres -dir $(MIGRATIONS_PATH) "postgres://user:password@localhost/social?sslmode=disable" up

.PHONY: down 
down:
	@goose postgres -dir $(MIGRATIONS_PATH) "postgres://user:password@localhost/social?sslmode=disable" down 

.PHONY: seed
seed:
	@$(GO) run $(SEED_PATH)

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: run
run:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt && $(GO) run ./cmd/api/*.go

.PHONY: test-api
test-api:
	@$(GO) test -v ./cmd/api/

.PHONY: test
test:
	@$(GO) test -v -race ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy -v
	$(GO) fmt ./...

.PHONY: test/cover
test/cover:
	$(GO) test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	$(GO) tool cover -html=/tmp/coverage.out
