.PHONY: help  run  docker-up docker-down clean migrate-up migrate-down migrate-version migrate-drop migrate-force migrate-create proto

migrate-up:
	@./scripts/migrate.sh up

migrate-down:
	@./scripts/migrate.sh down

migrate-drop:
	@./scripts/migrate.sh drop

migrate-version:
	@./scripts/migrate.sh version

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
        echo "❌ Usage: make migrate-force VERSION=<version>"; \
        exit 1; \
    fi
	@./scripts/migrate.sh force $(VERSION)

migrate-create:
	@if [ -z "$(NAME)" ]; then \
        echo "❌ Usage: make migrate-create NAME=<migration_name>"; \
        exit 1; \
    fi
	@./scripts/migrate.sh create $(NAME)

help:
	@echo "Available commands:"
	@echo "  make run          - Run the service locally"
	@echo "  make docker-up    - Start all services with Docker Compose"
	@echo "  make docker-down  - Stop all Docker services"
	@echo "  make clean        - Clean build artifacts"



run:
	@echo "Running igm-service..."
	@go run cmd/server/main.go



docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "✅ Services started"
	@docker-compose ps

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "✅ Services stopped"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "✅ Cleaned"
proto:
	@./scripts/proto-gen.sh

.DEFAULT_GOAL := help