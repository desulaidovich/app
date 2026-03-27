PROTO_FILES := $(shell find proto -name '*.proto')

.PHONY: init init-force env-to-json docker-up proto build run

ifneq (,$(wildcard .env))
    include .env
    export
endif

init:
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env"; \
		cp example.env .env; \
		echo "✅ .env created"; \
	else \
		echo "✅ .env already exists"; \
	fi
	@$(MAKE) show-config
	@echo ""
	@$(MAKE) docker-up

init-force:
	@echo "📝 Force creating .env"
	@cp example.env .env
	@echo "✅ .env overwritten"
	@$(MAKE) show-config
	@echo ""
	@$(MAKE) docker-up

show-config:
	@echo ""
	@echo "📋 Current configuration:"
	@echo "  APP_NAME:       $(APP_NAME)"
	@echo "  APP_ENV:        $(APP_ENV)"
	@echo "  HTTP_PORT:      $(HTTP_PORT)"
	@echo "  GRPC_PORT:      $(GRPC_PORT)"
	@echo "  DATABASE_HOST:  $(DATABASE_HOST)"
	@echo "  DATABASE_PORT:  $(DATABASE_PORT)"
	@echo "  DATABASE_NAME:  $(DATABASE_NAME)"
	@echo "  LOG_LEVEL:      $(LOG_LEVEL)"

env-to-json:
	@echo "{"
	@cat .env | grep -v '^#' | grep -v '^$$' | while IFS='=' read -r key value; do \
		echo "  \"$$key\": \"$$value\","; \
	done | sed '$$ s/,$$//'
	@echo "}"

docker-up:
	@echo "🚀 Starting PostgreSQL for $(APP_NAME)..."
	@docker-compose up -d
	@echo "✅ PostgreSQL is running on port $(DATABASE_PORT)"
	@echo "   Database: $(DATABASE_NAME)"
	@echo "   User: $(DATABASE_USER_NAME)"
	@echo "   Container: $(APP_NAME)_postgres"
	@echo ""
	@echo "✨ Project initialized and ready!"

docker-logs:
	@docker-compose logs -f

db-shell:
	@docker exec -it $(APP_NAME)_postgres psql -U $(DATABASE_USER_NAME) -d $(DATABASE_NAME)

docker-down:
	@docker-compose down

proto:
	@echo "⚙️  Generating protobuf code..."
	@echo "   Files: $(PROTO_FILES)"
	@mkdir -p api
	@protoc \
		-I proto \
		-I third_party \
		--go_out=api --go_opt=paths=source_relative \
		--go-grpc_out=api --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=api --grpc-gateway_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "✅ Proto generated → api/"


build:
	@go build \
		-ldflags "-X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) \
		          -X main.build=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
		-o bin/app ./cmd/app
	@echo "✅ Binary → bin/app"

run:
	@go run ./cmd/app

help:
	@echo "Available commands:"
	@echo "  make init        - Create .env, show config, and start PostgreSQL"
	@echo "  make init-force  - Force recreate .env and start PostgreSQL"
	@echo "  make docker-up   - Start PostgreSQL"
	@echo "  make docker-down - Stop PostgreSQL"
	@echo "  make docker-logs - View PostgreSQL logs"
	@echo "  make db-shell    - Open psql shell"
	@echo "  make proto       - Generate code from .proto files"
	@echo "  make build       - Build binary with version from git"
	@echo "  make run         - Run application"

.DEFAULT_GOAL := help
