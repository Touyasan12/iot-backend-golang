# Setup Guide

## Prerequisites

1. **Go 1.21+** - Install from https://golang.org/dl/
2. **PostgreSQL** (recommended) atau **SQLite** (untuk development)
3. **MQTT Broker** - Install Mosquitto atau gunakan MQTT broker lainnya

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Setup Database

#### Option A: PostgreSQL (Recommended)

1. Install PostgreSQL
2. Create database:

```sql
CREATE DATABASE aquarium_db;
```

#### Option B: SQLite (Development)

Tidak perlu setup, database akan dibuat otomatis.

### 3. Setup MQTT Broker

#### Install Mosquitto (Linux/Mac)

```bash
# Ubuntu/Debian
sudo apt-get install mosquitto mosquitto-clients

# Mac
brew install mosquitto
```

#### Install Mosquitto (Windows)

Download dari: https://mosquitto.org/download/

#### Start Mosquitto

```bash
# Linux/Mac
mosquitto -c /etc/mosquitto/mosquitto.conf

# Windows (dari install directory)
mosquitto.exe
```

### 4. Configure Environment

Copy `.env.example` ke `.env` dan edit sesuai kebutuhan:

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Server
SERVER_PORT=8080

# Database (PostgreSQL)
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=aquarium_db

# Atau SQLite (untuk development)
# DB_TYPE=sqlite
# DB_NAME=aquarium_db

# MQTT
MQTT_BROKER=localhost
MQTT_PORT=1883
MQTT_USER=
MQTT_PASS=
MQTT_CLIENT_ID=aquarium-backend
```

### 5. Run Application

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## Testing API

### Test Dashboard

```bash
curl http://localhost:8080/api/v1/dashboard
```

### Create Feeder Schedule

```bash
curl -X POST http://localhost:8080/api/v1/feeder/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "day_name": "Monday",
    "time": "08:00",
    "amount_gram": 10,
    "is_active": true
  }'
```

### Manual Feed

```bash
curl -X POST http://localhost:8080/api/v1/feeder/manual
```

### Create UV Schedule

```bash
curl -X POST http://localhost:8080/api/v1/uv/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "day_name": "Monday",
    "start_time": "20:00",
    "end_time": "04:00",
    "is_active": true
  }'
```

### Manual UV

```bash
curl -X POST http://localhost:8080/api/v1/uv/manual \
  -H "Content-Type: application/json" \
  -d '{
    "duration_minutes": 120
  }'
```

## Troubleshooting

### Database Connection Error

- Pastikan database sudah dibuat (untuk PostgreSQL)
- Periksa kredensial di `.env`
- Pastikan PostgreSQL service berjalan

### MQTT Connection Error

- Pastikan MQTT broker berjalan
- Periksa `MQTT_BROKER` dan `MQTT_PORT` di `.env`
- Test koneksi dengan:

```bash
mosquitto_pub -h localhost -t test -m "hello"
```

### Port Already in Use

- Ubah `SERVER_PORT` di `.env`
- Atau hentikan aplikasi yang menggunakan port tersebut
