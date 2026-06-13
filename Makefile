## go-http-server — build, test, and dev targets

BINARY     := server
BUILD_DIR  := bin
CMD_PATH   := ./cmd/server
GO         := $(shell which go 2>/dev/null || echo /usr/local/go/bin/go)
GOFLAGS    := -trimpath
LDFLAGS    := -s -w

.PHONY: all build run test test-race lint vet fmt tidy clean docker-build docker-run db-up db-down db-logs help

all: build

## Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

## Run the server locally loading .env if present
run:
	@[ -f .env ] && export $$(grep -v '^#' .env | grep -v '^$$' | xargs) ; \
	LOG_FORMAT=text LOG_LEVEL=debug $(GO) run $(CMD_PATH)

## Run all tests
test:
	$(GO) test ./... -count=1 -timeout 60s

## Run tests with the race detector
test-race:
	$(GO) test -race ./... -count=1 -timeout 60s

## Run tests with coverage report
test-cover:
	$(GO) test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

## Run go vet
vet:
	$(GO) vet ./...

## Format source code
fmt:
	$(GO) fmt ./...

## Tidy and vendor dependencies
tidy:
	$(GO) mod tidy

## Remove build artifacts
clean:
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

## Start the local database (Postgres) in the background
db-up:
	docker compose up -d

## Stop the local database (keeps the data volume)
db-down:
	docker compose down

## Tail the database logs
db-logs:
	docker compose logs -f postgres

## Build Docker image
docker-build:
	docker build -t go-http-server:latest .

## Run in Docker
docker-run:
	docker run --rm -p 8080:8080 \
		-e LOG_FORMAT=text \
		-e LOG_LEVEL=info \
		go-http-server:latest

## Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
