.PHONY: help  run  docker-up docker-down clean

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

.DEFAULT_GOAL := help