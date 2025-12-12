.PHONY: help up down logs restart backup db clean test build

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

test: ## Run tests (if any)
	go test -v ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

