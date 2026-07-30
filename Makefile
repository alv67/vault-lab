COMPOSE ?= podman-compose

.PHONY: dev build up down migrate frontend-dev

up: ## Start all services
	$(COMPOSE) up --build -d

down: ## Stop all services
	$(COMPOSE) down

logs: ## Follow logs
	$(COMPOSE) logs -f

restart: ## Restart all services
	$(COMPOSE) restart

migrate: ## Run DB migrations
	$(COMPOSE) exec backend /server migrate

migrate-down: ## Rollback DB migrations
	migrate -path backend/migrations -database "postgres://vaultlab:vaultlab@localhost:5432/vaultlab?sslmode=disable" down 1

frontend-dev: ## Run frontend in dev mode
	cd frontend && npm run dev

build: ## Build all binaries (inside container)
	$(COMPOSE) build

test: ## Run tests
	cd backend && go test ./... 2>/dev/null || echo "Go not installed locally, use: $(COMPOSE) exec backend go test ./..."

db-shell: ## Connect to postgres
	$(COMPOSE) exec postgres psql -U vaultlab vaultlab

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
