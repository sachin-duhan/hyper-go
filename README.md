# Go Hyper - Microservices Project

A modern microservices architecture with Go backend services and React TypeScript frontend, featuring real-time analytics and audit logging.

## Features

- 🔐 Authentication & Authorization
- 📊 Real-time Analytics
- 📝 Audit Logging
- 🎯 Event Tracking
- 🖥️ Modern Dashboard
- 🔄 Hot Reload Development
- 🐳 Docker Support

## Services

- **Backend**: Main API service (Go/Gin)
  - Authentication & Authorization
  - User Management
  - Event Publishing
  
- **Analytics**: Event tracking service (Go/ClickHouse)
  - Page Views
  - User Actions
  - API Requests
  - Custom Events

- **Audit Logs**: Activity logging service (Go/ClickHouse)
  - User Actions
  - Resource Access
  - Security Events
  - System Changes

- **Dashboard**: Admin interface (React/TypeScript)
  - User Management
  - Analytics Visualization
  - Audit Log Viewer
  - Real-time Updates

## Tech Stack

- **Backend**
  - Go 1.21+
  - Gin Web Framework
  - JWT Authentication
  - PostgreSQL 16
  - ClickHouse
  - RabbitMQ

- **Frontend**
  - React 18
  - TypeScript
  - Tailwind CSS
  - React Query
  - Zustand
  - React Router 6

- **Infrastructure**
  - Docker & Docker Compose
  - Air (Hot Reload)
  - Make
  - Migrate

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
make docker-up
make init-db
```

3. Start all services in development mode:

```bash
make dev-all
```

4. Access the services:
- Dashboard: http://localhost:5173
- Backend API: http://localhost:8080
- Analytics API: http://localhost:8081
- Audit Logs API: http://localhost:8082

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
- `make kill-ports`: Kill processes using service ports

### Service Ports

- Backend: 8080
- Analytics: 8081
- Audit Logs: 8082
- Dashboard: 5173 (dev) / 80 (prod)
- PostgreSQL: 5432
- ClickHouse: 8123/9000
- RabbitMQ: 5672/15672 (Management: 15672)

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
│   ├── events/         # Event publishing
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
- GET `/api/admin/users`: List all users (admin only)

### Analytics
- GET `/api/analytics/events`: Get user analytics events
- POST `/track`: Track events

### Audit Logs
- GET `/api/audit/logs`: Get user audit logs
- POST `/audit`: Log audit events

## Event Tracking

### Analytics Events
- Page Views
- API Requests
- User Login/Logout
- User Registration
- Custom Events

### Audit Logs
- Authentication attempts
- Resource access
- User actions
- System changes

## Environment Variables

Key environment variables (see `.env` for full list):
```
# Service Ports
BACKEND_PORT=8080
ANALYTICS_PORT=8081
AUDIT_LOGS_PORT=8082

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_turbo

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672

# ClickHouse
CLICKHOUSE_HOST=localhost:9000
CLICKHOUSE_DATABASE=default
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=

# Frontend
VITE_API_URL=http://localhost:8080
```

## Docker Support

Build and run with Docker:

```bash
docker-compose up -d
```

## License

MIT
