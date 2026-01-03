# Pagination Guide

Dokumentasi penggunaan pagination pada API Smart Aquarium Controller.

## 📋 Endpoints dengan Pagination

Pagination telah diimplementasikan pada endpoint-endpoint berikut:

### 1. **GET /api/v1/history**

- **Default page_size**: 50
- **Max page_size**: 200
- **Sorting**: Descending by `created_at`

### 2. **GET /api/v1/feeder/schedules**

- **Default page_size**: 20
- **Max page_size**: 100
- **Sorting**: By `day_name` and `time`

### 3. **GET /api/v1/uv/schedules**

- **Default page_size**: 20
- **Max page_size**: 100
- **Sorting**: By `day_name` and `start_time`

## 🔧 Query Parameters

Semua endpoint yang support pagination menerima parameter berikut:

| Parameter   | Type    | Required | Default | Description                             |
| ----------- | ------- | -------- | ------- | --------------------------------------- |
| `page`      | integer | No       | 1       | Halaman yang ingin diambil (minimum: 1) |
| `page_size` | integer | No       | varies  | Jumlah item per halaman                 |

## 📊 Response Format

Response dengan pagination menggunakan format wrapper object dengan struktur:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 156,
    "total_pages": 8
  }
}
```

### Field Descriptions

- **data**: Array berisi data yang diminta
- **pagination.page**: Halaman saat ini
- **pagination.page_size**: Jumlah item per halaman
- **pagination.total**: Total jumlah records di database
- **pagination.total_pages**: Total jumlah halaman yang tersedia

## 💡 Contoh Penggunaan

### History dengan Pagination

```bash
# Get first page (default 50 items)
GET /api/v1/history

# Get page 2 with 50 items
GET /api/v1/history?page=2

# Get page 1 with 100 items
GET /api/v1/history?page=1&page_size=100

# With filters
GET /api/v1/history?device_type=FEEDER&page=1&page_size=20
```

**Response:**

```json
{
  "data": [
    {
      "id": 145,
      "device_type": "UV",
      "trigger_source": "SCHEDULE",
      "status": "RUNNING",
      "start_time": "2025-11-22T00:23:00Z",
      "end_time": "2025-11-22T04:00:00Z",
      "value": 13020
    }
    // ... more records
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 145,
    "total_pages": 8
  }
}
```

### Feeder Schedules dengan Pagination

```bash
# Get all schedules (first page, 20 items)
GET /api/v1/feeder/schedules

# Get with custom page size
GET /api/v1/feeder/schedules?page=1&page_size=50
```

**Response:**

```json
{
  "data": [
    {
      "id": 1,
      "day_name": "Mon",
      "time": "08:00",
      "amount_gram": 10,
      "is_active": true,
      "created_at": "2025-11-22T00:00:00Z",
      "updated_at": "2025-11-22T00:00:00Z"
    }
    // ... more schedules
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 11,
    "total_pages": 1
  }
}
```

### UV Schedules dengan Pagination

```bash
# Get UV schedules
GET /api/v1/uv/schedules?page=1&page_size=10
```

**Response:**

```json
{
  "data": [
    {
      "id": 1,
      "day_name": "Mon",
      "start_time": "20:00",
      "end_time": "04:00",
      "is_active": true,
      "created_at": "2025-11-22T00:00:00Z",
      "updated_at": "2025-11-22T00:00:00Z"
    }
    // ... more schedules
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 7,
    "total_pages": 1
  }
}
```

## 🚫 Endpoints Tanpa Pagination

Endpoint berikut **TIDAK** menggunakan pagination karena nature data-nya:

- `GET /api/v1/dashboard` - Sudah limited to 20 latest history
- `GET /api/v1/stock` - Single record
- `GET /api/v1/uv/status` - Single record
- `GET /api/v1/feeder/last-feed` - Single record
- Semua POST/PUT/DELETE endpoints

## ⚙️ Implementation Details

### Backend Utils

Pagination diimplementasikan menggunakan utility functions di `utils/pagination.go`:

```go
// Get pagination params from query string
pagination := utils.GetPaginationParams(c, defaultPageSize, maxPageSize)

// Build response with metadata
paginationMeta := utils.BuildPaginationResponse(
    pagination.Page,
    pagination.PageSize,
    total
)
```

### Frontend Integration

Ketika mengintegrasikan dengan frontend (React/TanStack Query):

```typescript
// Example with TanStack Query
const { data } = useQuery({
  queryKey: ["history", page, pageSize],
  queryFn: () =>
    api.get("/history", {
      params: { page, page_size: pageSize },
    }),
});

// Access data
const items = data.data;
const pagination = data.pagination;
```

## 📝 Notes

- Jika `page` melebihi `total_pages`, response akan berisi empty array
- `page_size` yang melebihi maximum akan di-cap ke maximum value
- Pagination metadata selalu disertakan meskipun data kosong
- Total count dilakukan dengan efficient query menggunakan `COUNT(*)`

## 🔄 Migration from Old API

### Before (without pagination)

```bash
GET /api/v1/history?limit=100
Response: [...]  # Array langsung
```

### After (with pagination)

```bash
GET /api/v1/history?page=1&page_size=100
Response: { data: [...], pagination: {...} }
```

⚠️ **Breaking Change**: Response format berubah dari array langsung menjadi object dengan `data` dan `pagination` fields.

Frontend perlu update untuk mengakses `response.data` instead of `response` directly.
