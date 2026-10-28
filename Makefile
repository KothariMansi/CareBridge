.PHONY: help fmt fmt-check vet test build run check seed sqlc db-up db-down migrate-up migrate-down migrate-version migrate-new

# Read from .env locally; CI sets DATABASE_URL in the environment instead.
DATABASE_URL ?= $(shell grep '^DATABASE_URL=' .env | cut -d= -f2- | tr -d '\r')
MIGRATE = migrate -path db/migration -database "$(DATABASE_URL)"
COMPOSE = docker compose --env-file .env -f deployments/docker-compose.yml

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "%-18s %s\n", $$1, $$2}'

fmt: ## Format all Go files
	gofmt -w .

fmt-check: ## Fail if any Go file is not formatted
	@test -z "$$(gofmt -l .)" || (echo "unformatted files:"; gofmt -l .; exit 1)

vet: ## Run go vet
	go vet ./...

test: ## Run all tests with coverage
	go test ./... -cover

build: ## Build the API into bin/
	go build -o bin/ ./cmd/api

run: ## Run the API
	go run ./cmd/api

check: fmt-check vet test build ## Everything CI checks, locally

seed: ## Load synthetic development data (safe to run repeatedly)
	docker exec -i carebridge-postgres sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1' < db/seed/dev_seed.sql

sqlc: ## Regenerate the database code from db/query
	sqlc generate

db-up: ## Start PostgreSQL
	$(COMPOSE) up -d

db-down: ## Stop PostgreSQL (keeps the data)
	$(COMPOSE) down

migrate-up: ## Apply all pending migrations
	$(MIGRATE) up

migrate-down: ## Undo the last migration
	$(MIGRATE) down 1

migrate-version: ## Show the current migration version
	$(MIGRATE) version

migrate-new: ## Create a migration: make migrate-new name=add_doctors
	@test -n "$(name)" || (echo "usage: make migrate-new name=add_doctors"; exit 1)
	migrate create -ext sql -dir db/migration -seq $(name)