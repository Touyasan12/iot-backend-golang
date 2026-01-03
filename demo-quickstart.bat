@echo off
REM Quick Start Demo Script for Windows

echo === Smart Aquarium Controller - Demo Mode ===
echo.

REM Check if .env exists
if not exist .env (
    echo Creating .env file...
    (
        echo SERVER_PORT=8080
        echo DB_TYPE=sqlite
        echo DB_NAME=aquarium_db
        echo DEMO_MODE=true
    ) > .env
    echo .env file created!
)

echo Starting server in demo mode...
echo Server will be available at http://localhost:8080
echo.
echo After server starts, you can:
echo 1. Seed demo data: curl -X POST http://localhost:8080/api/v1/demo/seed
echo 2. Check dashboard: curl http://localhost:8080/api/v1/dashboard
echo 3. Manual feed: curl -X POST http://localhost:8080/api/v1/feeder/manual
echo.
echo Press Ctrl+C to stop the server
echo.

go run main.go


