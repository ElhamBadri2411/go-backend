include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations/

.PHONY: migrate
migrate:
	@goose create -dir $(MIGRATIONS_PATH) $(name)  sql

.PHONY: up
up:
	@goose postgres -dir $(MIGRATIONS_PATH) "postgres://user:password@localhost/social?sslmode=disable" up

.PHONY: down 
down:
	@goose postgres -dir $(MIGRATIONS_PATH) "postgres://user:password@localhost/social?sslmode=disable" down 
