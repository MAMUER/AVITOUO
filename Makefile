.PHONY: tidy fmt vet lint test check

tidy:
	@echo "Tidying Go modules..."
	@go mod tidy
	@echo "Tidy complete."

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Format complete."

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete."

lint:
	@echo "Running golangci-lint..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --max-issues-per-linter=0
	@echo "Lint complete."

test:
	@echo "Running unit tests..."
	@go test -v -timeout 5m ./...
	@echo "Tests complete."

check: tidy fmt vet lint test
	@echo "========================================"
	@echo "  LOCAL CHECKS PASSED!"
	@echo "========================================"

run:
	@go run .

help:
	@echo "Available commands:"
	@echo "  make tidy       - Tidy Go modules"
	@echo "  make fmt        - Format Go code"
	@echo "  make vet        - Run go vet"
	@echo "  make lint       - Run golangci-lint"
	@echo "  make test       - Run unit tests"
	@echo "  make check      - Run all checks"
	@echo "  make run        - Run application"