# Go Hyper - Microservices Project

A modern microservices architecture with Go backend services and React TypeScript frontend.

## Services

- **Backend**: Main API service (Go/Gin)
- **Analytics**: Event tracking service (Go/ClickHouse)
- **Audit Logs**: Activity logging service (Go/ClickHouse)
- **Dashboard**: Admin interface (React/TypeScript)

## Tech Stack

- Go 1.21+
- PostgreSQL 16
- ClickHouse
- RabbitMQ
- React 18
- TypeScript
- Tailwind CSS
- Docker & Docker Compose

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make

## Quick Start

1. Clone and setup:

```bash
git clone <repository-url>
cd go-hyper
cp .env.example .env
```

2. Start infrastructure and run migrations:

```bash
make init-db
```

3. Start all services in development mode:

```bash
make dev-all
```

## Development

### Available Make Commands

- `make dev-all`: Start all services with hot reload
- `make dev-backend`: Start backend service only
- `make dev-analytics`: Start analytics service only
- `make dev-audit-logs`: Start audit logs service only
- `make dev-dashboard`: Start dashboard only
- `make build`: Build all services
- `make test`: Run tests
- `make docker-up`: Start infrastructure services
- `make docker-down`: Stop infrastructure services
- `make migrate-up`: Run database migrations
- `make migrate-down`: Revert migrations

### Service Ports

- Backend: 8080
- Analytics: 8081
- Audit Logs: 8082
- Dashboard: 5173 (dev) / 80 (prod)
- PostgreSQL: 5432
- ClickHouse: 8123/9000
- RabbitMQ: 5672/15672

### Test Users

```
Admin:
- Email: admin@example.com
- Password: password123

User:
- Email: user@example.com
- Password: password123
```

## Project Structure

```
.
├── migrations/           # Database migrations
├── pkg/                 # Shared packages
│   ├── auth/           # Authentication utilities
│   ├── database/       # Database clients
│   ├── models/         # Shared data models
│   ├── queue/          # Message queue utilities
│   └── utils/          # Common utilities
├── scripts/            # Utility scripts
├── services/           # Microservices
│   ├── analytics/      # Analytics service
│   ├── audit-logs/     # Audit logging service
│   ├── backend/        # Main API service
│   └── dashboard/      # Frontend application
└── docker-compose.yml  # Infrastructure setup
```

## API Endpoints

### Auth
- POST `/api/auth/login`: User login
- POST `/api/auth/register`: User registration

### User
- GET `/api/user/profile`: Get user profile

### Admin
- GET `/api/admin/users`: List all users (admin only)

### Analytics
- POST `/track`: Track events

### Audit Logs
- POST `/audit`: Log audit events

## Environment Variables

Key environment variables (see `.env` for full list):
```
BACKEND_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_turbo
RABBITMQ_URL=amqp://guest:guest@localhost:5672
CLICKHOUSE_HOST=localhost:9000
```

## Docker Support

Build and run with Docker:

```bash
docker-compose up -d
```
