.PHONY: help install build start dev test clean setup demo load-test stress-test monitor docker-up docker-down docker-logs

help: ## Show this help message
	@echo "Hotel Booking API - Available Commands"
	@echo "====================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $1, $2}'

install: ## Install dependencies
	npm install

build: ## Build the project
	npm run build

start: ## Start the production server
	npm start

dev: ## Start the development server
	npm run dev

test: ## Run tests
	npm test

clean: ## Clean build artifacts
	rm -rf dist/
	rm -rf node_modules/

docker-up: ## Start Docker services
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

setup: ## Complete project setup with Docker
	@chmod +x scripts/*.sh
	@./scripts/setup.sh

demo: ## Run interactive demo
	@./scripts/demo.sh

load-test: ## Run load tests
	@./scripts/load-test.sh

stress-test: ## Run stress tests
	@./scripts/stress-test.sh

monitor: ## Monitor database activity
	@./scripts/monitor.sh

init-db: ## Initialize database
	npm run init-db

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U postgres -d hotel_booking

adminer: ## Open Adminer in browser
	@echo "Opening Adminer at http://localhost:8080"
	@echo "Use these credentials:"
	@echo "  System: PostgreSQL"
	@echo "  Server: postgres"
	@echo "  Username: postgres"
	@echo "  Password: password"
	@echo "  Database: hotel_booking"