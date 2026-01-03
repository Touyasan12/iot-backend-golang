# History Value Field Explanation

Dokumentasi penjelasan field `value` pada `ActionHistory`.

## 📊 Field Value

Field `value` pada response history memiliki arti yang berbeda tergantung pada `device_type`:

### 🐟 FEEDER

- **Unit**: Gram
- **Deskripsi**: Jumlah pakan yang diberikan
- **Contoh nilai**:
  - `10` = 10 gram pakan
  - `15` = 15 gram pakan
  - `20` = 20 gram pakan

### 💡 UV

- **Unit**: Detik (seconds)
- **Deskripsi**: Durasi UV menyala
- **Contoh nilai**:
  - `7200` = 2 jam (2 × 60 × 60)
  - `14400` = 4 jam (4 × 60 × 60)
  - `28800` = 8 jam (8 × 60 × 60)

## 🔍 Contoh Response

### Feeder Manual

```json
{
  "id": 100,
  "device_type": "FEEDER",
  "trigger_source": "MANUAL",
  "start_time": "2025-11-26T08:00:00Z",
  "end_time": "2025-11-26T08:00:05Z",
  "status": "SUCCESS",
  "value": 15, // ← 15 gram pakan
  "created_at": "2025-11-26T08:00:00Z",
  "updated_at": "2025-11-26T08:00:05Z"
}
```

### Feeder Schedule

```json
{
  "id": 101,
  "device_type": "FEEDER",
  "trigger_source": "SCHEDULE",
  "start_time": "2025-11-26T08:00:00Z",
  "end_time": "2025-11-26T08:00:05Z",
  "status": "SUCCESS",
  "value": 10, // ← 10 gram pakan (default schedule)
  "created_at": "2025-11-26T08:00:00Z",
  "updated_at": "2025-11-26T08:00:05Z"
}
```

### UV Manual

```json
{
  "id": 200,
  "device_type": "UV",
  "trigger_source": "MANUAL",
  "start_time": "2025-11-26T10:00:00Z",
  "end_time": "2025-11-26T12:00:00Z",
  "status": "SUCCESS",
  "value": 7200, // ← 7200 detik = 2 jam
  "created_at": "2025-11-26T10:00:00Z",
  "updated_at": "2025-11-26T12:00:00Z"
}
```

### UV Schedule

```json
{
  "id": 201,
  "device_type": "UV",
  "trigger_source": "SCHEDULE",
  "start_time": "2025-11-25T20:00:00Z",
  "end_time": "2025-11-26T04:00:00Z",
  "status": "RUNNING",
  "value": 28800, // ← 28800 detik = 8 jam (20:00 - 04:00)
  "created_at": "2025-11-25T20:00:00Z",
  "updated_at": "2025-11-25T20:00:00Z"
}
```

## 🔄 Kalkulasi Value untuk UV Schedule

### Scenario 1: Normal Range (Start < End)

Jadwal: `20:00 - 23:00` (3 jam)

- Jika scheduler trigger pada `20:00`
- Duration = `23:00 - 20:00` = 3 jam = 10,800 detik
- **value = 10800**

### Scenario 2: Overnight Range (Start > End)

Jadwal: `20:00 - 04:00` (8 jam melewati tengah malam)

#### Case A: Trigger sebelum tengah malam (20:00)

- Current time: `20:00`
- End time: `04:00` (besok)
- Duration = (4 jam sampai midnight) + (4 jam setelah midnight) = 8 jam = 28,800 detik
- **value = 28800**

#### Case B: Trigger setelah tengah malam (01:00)

- Current time: `01:00`
- End time: `04:00` (hari yang sama)
- Duration = `04:00 - 01:00` = 3 jam = 10,800 detik
- **value = 10800**

## 💻 Frontend Usage

### Display Duration (UV)

```typescript
// Convert seconds to hours
const durationHours = history.value / 3600;

// Display: "2 jam", "8 jam", etc.
return `${durationHours} jam`;
```

### Display dengan Format HH:MM

```typescript
function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}:${minutes.toString().padStart(2, "0")}`;
}

// Usage
formatDuration(7200); // "2:00"
formatDuration(28800); // "8:00"
formatDuration(14430); // "4:00"
```

### Display Amount (Feeder)

```typescript
// Display: "10 gram", "15 gram", etc.
return `${history.value} gram`;
```

## 📝 Summary

| Device Type | Unit  | Typical Values           | Example                 |
| ----------- | ----- | ------------------------ | ----------------------- |
| **FEEDER**  | Gram  | 5, 10, 15, 20, 25        | `value: 10` = 10g pakan |
| **UV**      | Detik | 3600, 7200, 14400, 28800 | `value: 28800` = 8 jam  |

## ⚠️ Historical Note

**Sebelum fix (versi lama):**

- UV schedule memiliki `value: 0`
- Ini tidak informatif dan tidak konsisten dengan dokumentasi

**Setelah fix (versi baru):**

- UV schedule memiliki `value` yang berisi durasi aktual dalam detik
- Konsisten dengan dokumentasi: "seconds untuk UV"
- Lebih informatif untuk frontend dan analytics

## 🔧 Implementation Reference

File: `scheduler/scheduler.go`

```go
// Calculate duration in minutes first
durationMinutes := endTimeMinutes - currentTimeMinutes

// Create action with value in SECONDS
action := models.ActionHistory{
    DeviceType:    "UV",
    TriggerSource: "SCHEDULE",
    Value:         durationMinutes * 60, // Convert to seconds
    // ... other fields
}
```

File: `handlers/uv.go` (Manual UV)

```go
durationSec := req.DurationMinutes * 60

action := models.ActionHistory{
    DeviceType:    "UV",
    TriggerSource: "MANUAL",
    Value:         durationSec, // Already in seconds
    // ... other fields
}
```
