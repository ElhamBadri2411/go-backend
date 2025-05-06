include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations/
SEED_PATH = ./cmd/migrate/seed/main.go

.PHONY: migrate
migrate:
	@goose create -dir $(MIGRATIONS_PATH) $(name)  sql

.PHONY: up
up:
	@goose postgres -dir $(MIGRATIONS_PATH) "postgres://user:password@localhost/social?sslmode=disable" up

.PHONY: down 
down:
	@goose postgres -dir $(MIGRATIONS_PATH) "postgres://user:password@localhost/social?sslmode=disable" down 

.PHONY: seed
seed:
	@go run $(SEED_PATH)

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: run
run:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt && go run ./cmd/api/*.go
