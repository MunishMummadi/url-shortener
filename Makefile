# Makefile
.PHONY: test test-unit test-integration run build clean

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test -v -cover

# Run integration tests (requires server to be running)
test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/...

# Start the application
run:
	@go run main.go

# Build the application
build:
	@mkdir -p bin
	@go build -o bin/url-shortener

# Clean build artifacts
clean:
	@rm -rf bin/

# Run tests with coverage report
coverage:
	@go test -coverprofile=coverage.out
	@go tool cover -html=coverage.out

# Run all tests with coverage
test-all: test coverage