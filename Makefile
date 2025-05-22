include .env

GO := /usr/local/go/bin/go
MIGRATIONS_PATH = ./cmd/migrate/migrations/
SEED_PATH = ./cmd/migrate/seed/main.go
main_package_path = ./cmd/api/
binary_name = DevSocial

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


## build: build the application
.PHONY: build
build:
	go build -o=/tmp/bin/${binary_name} ${main_package_path}

.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
        --build.cmd "make build" --build.bin "/tmp/bin/${binary_name}" --build.delay "100" \
        --build.exclude_dir "" \
        --build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
        --misc.clean_on_exit "true"


.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: cannon
cannon:
	npx autocannon -r 4000 -d 2 -c 10 --renderStatusCodes http://localhost:3000/v1/health
