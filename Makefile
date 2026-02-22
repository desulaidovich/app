.PHONY: init init-force env-to-json docker-up

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
	@echo "🔐 APP_NAME: $(APP_NAME)"
	@echo "🔐 APP_ENV: $(APP_ENV)"
	@echo "🔐 APP_DEBUG: $(APP_DEBUG)"
	@echo "🔐 HTTP_PORT: $(HTTP_PORT)"
	@echo "🔐 DATABASE_HOST: $(DATABASE_HOST)"
	@echo "🔐 DATABASE_PORT: $(DATABASE_PORT)"
	@echo "🔐 DATABASE_NAME: $(DATABASE_NAME)"
	@echo "🔐 DATABASE_USER_NAME: $(DATABASE_USER_NAME)"
	@echo "🔐 LOG_LEVEL: $(LOG_LEVEL)"
	@echo "🔐 LOG_FORMAT: $(LOG_FORMAT)"

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

help:
	@echo "Available commands:"
	@echo "  make init        - Create .env, show config, and start PostgreSQL"
	@echo "  make init-force  - Force recreate .env and start PostgreSQL"
	@echo "  make docker-up   - Start PostgreSQL"
	@echo "  make docker-down - Stop PostgreSQL"
	@echo "  make docker-logs - View PostgreSQL logs"
	@echo "  make db-shell    - Open psql shell"

.DEFAULT_GOAL := help
