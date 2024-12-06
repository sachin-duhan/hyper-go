#!/bin/bash

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "golang-migrate not found. Installing..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Load environment variables
source .env

# Construct database URL
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Default to "up" if no direction specified
DIRECTION=${1:-up}

echo "Running migrations ${DIRECTION}..."
migrate -database "${DB_URL}" -path migrations "${DIRECTION}"

if [ $? -eq 0 ]; then
    echo "Migrations completed successfully!"
else
    echo "Migration failed!"
    exit 1
fi 