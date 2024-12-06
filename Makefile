.PHONY: build run test clean docker-up docker-down dev dev-backend dev-analytics dev-audit-logs dev-dashboard dev-all install-tools setup-env kill-ports

# Build all services
build:
	@echo "Building all services..."
	cd services/backend && go build -o ../../bin/backend
	cd services/analytics && go build -o ../../bin/analytics
	cd services/audit-logs && go build -o ../../bin/audit-logs

# Kill processes using our ports
kill-ports:
	@echo "Killing processes using our ports..."
	-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	-lsof -ti:8081 | xargs kill -9 2>/dev/null || true
	-lsof -ti:8082 | xargs kill -9 2>/dev/null || true
	-lsof -ti:5173 | xargs kill -9 2>/dev/null || true

# Setup environment files
setup-env:
	@echo "Setting up environment files..."
	@cp .env services/backend/.env || true
	@cp .env services/analytics/.env || true
	@cp .env services/audit-logs/.env || true
	@cp .env services/dashboard/.env || true

# Run all services locally
run: kill-ports setup-env
	@echo "Starting all services..."
	docker-compose up -d postgres clickhouse rabbitmq
	@echo "Waiting for databases to be ready..."
	sleep 5
	./bin/backend & ./bin/analytics & ./bin/audit-logs &

# Run tests
test:
	@echo "Running tests..."
	go test ./pkg/... ./services/... -v

# Clean build artifacts
clean: kill-ports
	@echo "Cleaning build artifacts..."
	rm -rf bin/ tmp/
	rm -f services/*/.env

# Start docker services
docker-up:
	@echo "Starting docker services..."
	docker-compose up -d

# Stop docker services
docker-down: kill-ports
	@echo "Stopping docker services..."
	docker-compose down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	cd pkg && go mod tidy
	cd services/backend && go mod tidy
	cd services/analytics && go mod tidy
	cd services/audit-logs && go mod tidy
	cd services/dashboard && npm install

# Create development database
init-db: docker-up
	@echo "Creating development database..."
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	@docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE go_turbo;" || true
	@docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE go_turbo_test;" || true
	@echo "Running migrations..."
	@make migrate-up

# Show service status
status:
	@echo "Service Status:"
	@docker-compose ps

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	npm install -g pnpm

# Run all services in development mode
dev: docker-up setup-env kill-ports
	@echo "Starting all services in development mode..."
	make dev-backend & make dev-analytics & make dev-audit-logs

# Run backend service in development mode
dev-backend: setup-env kill-ports
	@echo "Starting backend service in development mode..."
	cd services/backend && air

# Run analytics service in development mode
dev-analytics: setup-env kill-ports
	@echo "Starting analytics service in development mode..."
	cd services/analytics && air

# Run audit-logs service in development mode
dev-audit-logs: setup-env kill-ports
	@echo "Starting audit-logs service in development mode..."
	cd services/audit-logs && air

# Run dashboard in development mode
dev-dashboard:
	@echo "Starting dashboard in development mode..."
	cd services/dashboard && bun dev

# Run all services including dashboard
dev-all: clean docker-up init-db setup-env kill-ports
	@echo "Starting all services and dashboard..."
	@echo "Waiting for infrastructure services..."
	@sleep 5
	@mkdir -p tmp/logs
	@touch tmp/logs/backend.log tmp/logs/analytics.log tmp/logs/audit-logs.log tmp/logs/dashboard.log
	@(cd services/backend && air > ../../tmp/logs/backend.log 2>&1 & echo "Backend started") & \
	(cd services/analytics && air > ../../tmp/logs/analytics.log 2>&1 & echo "Analytics started") & \
	(cd services/audit-logs && air > ../../tmp/logs/audit-logs.log 2>&1 & echo "Audit-logs started") & \
	(cd services/dashboard && pnpm dev > ../../tmp/logs/dashboard.log 2>&1 & echo "Dashboard started") & \
	echo "All services started. Logs are in tmp/logs/"
	@tail -f tmp/logs/*.log

# Migration commands
migrate-up:
	@echo "Running migrations up..."
	@bash ./scripts/migrate.sh up

migrate-down:
	@echo "Running migrations down..."
	@bash ./scripts/migrate.sh down

migrate-force:
	@echo "Forcing migration version..."
	@bash ./scripts/migrate.sh force $(version)