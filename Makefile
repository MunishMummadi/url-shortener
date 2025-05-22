# Makefile
.PHONY: help build run test coverage clean docker-build docker-run docker-stop

APP_NAME=url-shortener
DOCKER_IMAGE_NAME=url-shortener-app
DOCKER_CONTAINER_NAME=url-shortener-container

# Default target
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build the application binary"
	@echo "  run            Build and run the application binary"
	@echo "  test           Run all tests with coverage"
	@echo "  coverage       Generate HTML coverage report"
	@echo "  clean          Clean up build artifacts and coverage reports"
	@echo "  docker-build   Build the Docker image"
	@echo "  docker-run     Run the Docker container in detached mode"
	@echo "  docker-stop    Stop and remove the Docker container"

# Build the application
build:
	@echo "Building binary..."
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) main.go

# Run the application (using the built binary)
run: build
	@echo "Running application (binary)..."
	@./bin/$(APP_NAME)

# Run all tests with coverage
test:
	@echo "Running tests..."
	@go test -v ./... -cover

# Generate HTML coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Clean build artifacts and coverage reports
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f $(APP_NAME) # If binary was built in root
	@rm -f coverage.out

# Docker commands
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE_NAME)..."
	@docker build -t $(DOCKER_IMAGE_NAME) .

docker-run:
	@echo "Running Docker container $(DOCKER_CONTAINER_NAME)..."
	@docker run -d --name $(DOCKER_CONTAINER_NAME) -p 8080:8080 --env-file .env $(DOCKER_IMAGE_NAME)
	@echo "Container $(DOCKER_CONTAINER_NAME) started. Access at http://localhost:8080"
	@echo "View logs: docker logs $(DOCKER_CONTAINER_NAME)"
	@echo "To stop: make docker-stop"

docker-stop:
	@echo "Stopping and removing Docker container $(DOCKER_CONTAINER_NAME)..."
	@docker stop $(DOCKER_CONTAINER_NAME) || true
	@docker rm $(DOCKER_CONTAINER_NAME) || true