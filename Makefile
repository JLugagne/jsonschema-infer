.PHONY: test build clean

# Run tests
test:
	go test -v

# Build
build:
	go build

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run tests with race detector
test-race:
	go test -v -race

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html

# Run go mod tidy
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run all checks
check: fmt test lint
