# OpenRAGLecture/Makefile

# ==============================================================================
# VARIABLES
# ==============================================================================

# Go parameters
BINARY_NAME=main
BINARY_DIR=./tmp
API_CMD_PATH=./cmd/api
BATCH_CMD_PATH=./cmd/batch

# Go binary paths
GOPATH ?= $(shell go env GOPATH)
ifeq ($(GOPATH),)
GOPATH = $(HOME)/go
endif
GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN = $(GOPATH)/bin
endif

# Docker Compose parameters
DOCKER_COMPOSE_CMD=docker compose

# Migration parameters
# ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
# 修正点: `pwd` を使ってプロジェクトルートの絶対パスを取得し、
# MIGRATE_PATH を絶対パスで定義します。
# これで、どこから `make` が呼ばれてもパスの解釈が常に一定になります。
# ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
PROJECT_DIR := $(shell pwd)
MIGRATE_CMD=$(GOBIN)/migrate
MIGRATE_PATH=$(PROJECT_DIR)/scripts/migrations
MIGRATE_DB_URL=mysql://$(shell grep MYSQL_USER .env | cut -d '=' -f2):$(shell grep MYSQL_PASSWORD .env | cut -d '=' -f2)@tcp(localhost:$(shell grep MYSQL_HOST_PORT .env | cut -d '=' -f2))/$(shell grep MYSQL_DATABASE .env | cut -d '=' -f2)

# Coverage parameters
COVERAGE_PKGS=./internal/domain/...,./internal/interface/...,./internal/usecase/...,./pkg/...

# ==============================================================================
# HELP
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo "Usage: make <command>"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==============================================================================
# DEVELOPMENT ENVIRONMENT
# ==============================================================================

.PHONY: setup up down logs clean
setup: build-migrate-tool ## Setup the development environment (build, start, and migrate)
	@echo "Setting up development environment..."
	@$(DOCKER_COMPOSE_CMD) build
	@make up

up: ## Start all services and run migrations
	@echo "Starting Docker containers..."
	@$(DOCKER_COMPOSE_CMD) up -d
	@echo "Waiting for DB to be ready..."
	@./scripts/wait-for-it.sh localhost:$(shell grep MYSQL_HOST_PORT .env | cut -d '=' -f2) --timeout=30
	@echo "Applying migrations to the development database..."
	@make migrate-up

down: ## Stop and remove all services
	@echo "Stopping Docker containers..."
	@$(DOCKER_COMPOSE_CMD) down

logs: ## Follow logs for all services
	@echo "Following logs..."
	@$(DOCKER_COMPOSE_CMD) logs -f

clean: ## Stop containers and remove all volumes
	@echo "Cleaning up Docker environment..."
	@$(DOCKER_COMPOSE_CMD) down -v --remove-orphans
	@echo "Removing tmp directory..."
	@rm -rf $(BINARY_DIR)


# ==============================================================================
# BUILD & RUN
# ==============================================================================

.PHONY: build run-dev
build: ## Build the Go API binary
	@echo "Building Go binary..."
	@go build -o $(BINARY_DIR)/$(BINARY_NAME) $(API_CMD_PATH)

run-dev: ## Run the API server with hot-reloading using air
	@echo "Starting server with hot-reloading..."
	@air


# ==============================================================================
# TEST COMMANDS
# ==============================================================================

.PHONY: test-unit test-e2e
test-unit: ## Run all unit tests and show coverage
	@echo "Running unit tests with coverage..."
	@go test -v -cover -coverpkg=$(COVERAGE_PKGS) ./internal/tests/...

test-e2e: ## Run end-to-end tests
	@echo "Starting E2E test setup..."
	@$(DOCKER_COMPOSE_CMD) up -d db
	@echo "Waiting for DB to be ready..."
	@./scripts/wait-for-it.sh localhost:$(shell grep MYSQL_HOST_PORT .env | cut -d '=' -f2) --timeout=30
	@echo "Running migrations for test DB..."
	@$(MIGRATE_CMD) -path file://$(MIGRATE_PATH) -database "$(MIGRATE_DB_URL)" up
	@echo "Running E2E tests..."
	@go test -v ./cmd/api -run TestE2ETestSuite -timeout 5m
	@echo "Cleaning up E2E test setup..."
	@echo "Rolling back migrations for test DB..."
	@$(MIGRATE_CMD) -path file://$(MIGRATE_PATH) -database "$(MIGRATE_DB_URL)" down -all
	@$(DOCKER_COMPOSE_CMD) down


# ==============================================================================
# DATABASE MIGRATION
# ==============================================================================

.PHONY: migrate-create migrate-up migrate-down migrate-force build-migrate-tool
migrate-create: build-migrate-tool ## Create a new migration file (e.g., make migrate-create name=add_users_table)
	@echo "Creating migration file: $(name)"
	@$(MIGRATE_CMD) create -ext sql -dir $(MIGRATE_PATH) -seq $(name)

migrate-up: build-migrate-tool ## Apply all up migrations
	@echo "Applying migrations..."
	# ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	# 修正点: file:// スキーマを明示的に指定
	# 絶対パスを使う場合は、これが推奨される形式です。
	# ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	@$(MIGRATE_CMD) -path file://$(MIGRATE_PATH) -database "$(MIGRATE_DB_URL)" up

migrate-down: build-migrate-tool ## Rollback the last migration
	@echo "Rolling back last migration..."
	@$(MIGRATE_CMD) -path file://$(MIGRATE_PATH) -database "$(MIGRATE_DB_URL)" down 1

migrate-force: build-migrate-tool ## Force a specific migration version (e.g., make migrate-force version=20240101120000)
	@echo "Forcing migration version to $(version)..."
	@$(MIGRATE_CMD) -path file://$(MIGRATE_PATH) -database "$(MIGRATE_DB_URL)" force $(version)

build-migrate-tool: ## Install the golang-migrate CLI tool if not present
	@if [ ! -f "$(MIGRATE_CMD)" ]; then \
		echo "golang-migrate CLI not found, installing..."; \
		go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi

# ==============================================================================
# BATCH COMMANDS
# ==============================================================================
.PHONY: batch-sync-documents
batch-sync-documents: ## Run the 'sync-documents' batch job
	@echo "Running 'sync-documents' batch job inside a container..."
	@$(DOCKER_COMPOSE_CMD) run --rm batch sync-documents