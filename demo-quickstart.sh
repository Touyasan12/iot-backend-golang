#!/bin/bash

# Quick Start Demo Script
echo "=== Smart Aquarium Controller - Demo Mode ==="
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
SERVER_PORT=8080
DB_TYPE=sqlite
DB_NAME=aquarium_db
DEMO_MODE=true
EOF
    echo ".env file created!"
fi

echo "Starting server in demo mode..."
echo "Server will be available at http://localhost:8080"
echo ""
echo "After server starts, you can:"
echo "1. Seed demo data: curl -X POST http://localhost:8080/api/v1/demo/seed"
echo "2. Check dashboard: curl http://localhost:8080/api/v1/dashboard"
echo "3. Manual feed: curl -X POST http://localhost:8080/api/v1/feeder/manual"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

go run main.go


