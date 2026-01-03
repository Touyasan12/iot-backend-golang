# Smart Aquarium Controller - Backend

Backend Golang untuk sistem Smart Aquarium Controller dengan fitur Automatic Feeder dan UV Sterilizer.

## Fitur

- **Automatic Feeder**: Sistem pemberian pakan dengan mekanisme Double Gate
- **UV Sterilizer**: Kontrol lampu UV dengan penjadwalan dan manual override
- **MQTT Communication**: Komunikasi dengan hardware ESP32/Arduino
- **REST API**: API untuk frontend dan kontrol manual
- **Cron Scheduler**: Penjadwalan otomatis untuk feeding dan UV
- **Pagination**: Support pagination untuk list endpoints (history, schedules)

## Requirements

- Go 1.21 atau lebih baru
- PostgreSQL atau SQLite
- MQTT Broker (Mosquitto, etc.)

## Installation

1. Clone repository

```bash
git clone <repository-url>
cd iot-backend-cursor
```

2. Install dependencies

```bash
go mod download
```

3. Setup environment variables

```bash
cp .env.example .env
# Edit .env sesuai konfigurasi Anda
```

4. Run application

```bash
go run main.go
```

## Configuration

Edit file `.env` untuk mengatur konfigurasi:

- **Database**: PostgreSQL atau SQLite
- **MQTT Broker**: Alamat dan port MQTT broker
- **Server Port**: Port untuk REST API (default: 8080)

## API Documentation

Dokumentasi API lengkap menggunakan **OpenAPI 3.1.0** tersedia di:

- **`openapi.yaml`** - Spesifikasi OpenAPI lengkap
- **`swagger-ui.html`** - Swagger UI untuk preview dan testing

### Cara Menggunakan Dokumentasi

1. **Swagger UI (Lokal)**: Buka `swagger-ui.html` di browser
2. **Swagger Editor (Online)**: https://editor.swagger.io/ - Copy paste `openapi.yaml`
3. **Redoc (Online)**: https://redocly.github.io/redoc/ - Upload `openapi.yaml`

Lihat `docs/README.md` untuk panduan lengkap.

## API Endpoints

### Dashboard

- `GET /api/v1/dashboard` - Get dashboard data (stock, UV status, history)

### Feeder

- `GET /api/v1/feeder/schedules` - Get all feeding schedules (with pagination)
- `POST /api/v1/feeder/schedules` - Create new feeding schedule
- `PUT /api/v1/feeder/schedules/:id` - Update feeding schedule
- `DELETE /api/v1/feeder/schedules/:id` - Delete feeding schedule
- `POST /api/v1/feeder/manual` - Trigger manual feed (support `amount_gram`, default 10g)
- `GET /api/v1/feeder/last-feed` - Get last feed information

### UV Sterilizer

- `GET /api/v1/uv/schedules` - Get all UV schedules (with pagination)
- `POST /api/v1/uv/schedules` - Create new UV schedule
- `PUT /api/v1/uv/schedules/:id` - Update UV schedule
- `DELETE /api/v1/uv/schedules/:id` - Delete UV schedule
- `POST /api/v1/uv/manual` - Trigger manual UV (with duration)
- `POST /api/v1/uv/manual/stop` - Stop running manual UV override
- `GET /api/v1/uv/status` - Get current UV status

### History

- `GET /api/v1/history` - Get action history (with pagination)
  - Query params: `device_type`, `trigger_source`, `status`, `page`, `page_size`

### Stock

- `GET /api/v1/stock` - Get current food stock
- `PUT /api/v1/stock` - Update food stock

## MQTT Topics

### Published by Server

- `aquarium/feeder/command` - Command to feeder device
- `aquarium/uv/command` - Command to UV device

### Subscribed by Server

- `aquarium/feeder/status` - Feeder device status
- `aquarium/uv/status` - UV device status
- `aquarium/device/report` - Device action reports

## Database Schema

### pakan_schedules

- `id` (primary key)
- `day_name` (Mon-Sun)
- `time` (HH:MM)
- `amount_gram` (default: 10)
- `is_active` (boolean)
- `created_at`, `updated_at`

### uv_schedules

- `id` (primary key)
- `day_name` (Mon-Sun)
- `start_time` (HH:MM)
- `end_time` (HH:MM)
- `is_active` (boolean)
- `created_at`, `updated_at`

### action_history

- `id` (primary key)
- `device_type` (FEEDER, UV)
- `trigger_source` (SCHEDULE, MANUAL)
- `start_time` (timestamp)
- `end_time` (timestamp, nullable)
- `status` (PENDING, RUNNING, SUCCESS, FAILED, OVERRIDDEN)
- `value` (grams for feeder, seconds for UV)
- `created_at`, `updated_at`

### stock

- `id` (primary key)
- `amount_gram` (integer)
- `updated_at`

### device_status

- `id` (primary key)
- `device_type` (FEEDER, UV)
- `status` (IDLE, DISPENSING, ON, OFF)
- `remaining` (seconds for UV)
- `last_updated`

## Development

### Project Structure

```
.
├── config/         # Configuration management
├── database/       # Database initialization
├── docs/           # Documentation
│   ├── INTEGRATION.md  # Integration guide
│   ├── PAGINATION.md   # Pagination usage guide
│   └── README.md       # Docs overview
├── handlers/       # HTTP request handlers
├── models/         # Database models
├── mqtt/          # MQTT client
├── routes/         # API routes
├── scheduler/      # Cron job scheduler
├── utils/          # Utility functions (pagination, feed, etc.)
├── main.go        # Application entry point
└── go.mod         # Go modules
```

## License

MIT
