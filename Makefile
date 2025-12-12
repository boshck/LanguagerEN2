.PHONY: help up down logs restart backup db clean test test-unit test-verbose test-coverage test-watch build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Start all services
	docker-compose up -d --build

down: ## Stop all services
	docker-compose down

logs: ## Show bot logs (follow mode)
	docker-compose logs -f bot

logs-all: ## Show all services logs
	docker-compose logs -f

restart: ## Restart bot service
	docker-compose restart bot

backup: ## Create manual backup
	@echo "Creating manual backup..."
	@docker-compose exec -T postgres pg_dump -U languager languager > backups/manual_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Backup created in backups/"

db: ## Connect to PostgreSQL
	docker-compose exec postgres psql -U languager -d languager

db-shell: ## Open shell in postgres container
	docker-compose exec postgres sh

bot-shell: ## Open shell in bot container
	docker-compose exec bot sh

clean: ## Remove all containers, volumes and backups
	docker-compose down -v
	rm -rf backups/*.sql

ps: ## Show running containers
	docker-compose ps

build: ## Build bot image
	docker-compose build bot

test: ## Run all tests
	@echo "Running tests..."
	@go test ./... -v

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test ./internal/... -v

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@go test ./... -v -race

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

test-watch: ## Run tests in watch mode (requires entr)
	@echo "Watching for changes (requires 'entr')..."
	@find . -name '*.go' | entr -c go test ./...

test-ci: ## Run tests as in CI (with race detector and coverage check)
	@echo "Running CI tests..."
	@go test ./... -v -race -coverprofile=coverage.out
	@go tool cover -func=coverage.out
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${coverage}%"; \
	if [ $$(echo "$$coverage < 80" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $${coverage}% is below minimum 80%"; \
		exit 1; \
	else \
		echo "✅ Coverage $${coverage}% meets minimum requirement"; \
	fi

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

