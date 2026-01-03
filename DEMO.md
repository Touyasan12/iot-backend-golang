# Demo Mode Guide

Mode demo memungkinkan Anda menjalankan backend tanpa memerlukan MQTT broker atau hardware. Semua komunikasi dengan device disimulasikan.

## Cara Menjalankan Demo

### 1. Setup Environment

Buat file `.env` dengan konfigurasi berikut:

```env
# Server
SERVER_PORT=8080

# Database (gunakan SQLite untuk kemudahan)
DB_TYPE=sqlite
DB_NAME=aquarium_db

# Demo Mode - PENTING: Set ke true
DEMO_MODE=true
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Jalankan Server

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080` dengan mode demo aktif.

### 4. Seed Demo Data

Setelah server berjalan, seed data demo dengan mengirim request:

```bash
curl -X POST http://localhost:8080/api/v1/demo/seed
```

Ini akan membuat:

- Stock pakan: 1000 gram (1kg)
- Sample jadwal pakan untuk beberapa hari
- Sample jadwal UV untuk setiap hari
- Sample history untuk 7 hari terakhir

## Testing API Endpoints

### 1. Dashboard

```bash
curl http://localhost:8080/api/v1/dashboard
```

### 2. Manual Feed

```bash
curl -X POST http://localhost:8080/api/v1/feeder/manual
```

**Hasil:**

- Feeder akan mensimulasikan proses feeding (5 detik)
- Stock akan berkurang 10 gram
- History akan terupdate dengan status SUCCESS

### 3. Get Last Feed Info

```bash
curl http://localhost:8080/api/v1/feeder/last-feed
```

### 4. Manual UV

```bash
curl -X POST http://localhost:8080/api/v1/uv/manual \
  -H "Content-Type: application/json" \
  -d '{"duration_minutes": 5}'
```

**Hasil:**

- UV akan menyala selama 5 menit
- Status akan terupdate setiap detik (countdown)
- Setelah 5 menit, UV akan mati otomatis

### 5. Get UV Status

```bash
curl http://localhost:8080/api/v1/uv/status
```

### 6. Get Feeder Schedules

```bash
curl http://localhost:8080/api/v1/feeder/schedules
```

### 7. Create New Feeder Schedule

```bash
curl -X POST http://localhost:8080/api/v1/feeder/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "day_name": "Monday",
    "time": "14:00",
    "amount_gram": 10,
    "is_active": true
  }'
```

### 8. Get History

```bash
# Get all history
curl http://localhost:8080/api/v1/history

# Filter by device type
curl http://localhost:8080/api/v1/history?device_type=FEEDER

# Filter by trigger source
curl http://localhost:8080/api/v1/history?trigger_source=MANUAL

# Limit results
curl http://localhost:8080/api/v1/history?limit=10
```

### 9. Update Stock

```bash
curl -X PUT http://localhost:8080/api/v1/stock \
  -H "Content-Type: application/json" \
  -d '{"amount_gram": 2000}'
```

### 10. Get Stock

```bash
curl http://localhost:8080/api/v1/stock
```

## Simulasi Scheduler

Scheduler akan berjalan setiap menit dan:

- Mengecek jadwal pakan yang sesuai dengan waktu saat ini
- Mengecek jadwal UV yang sesuai dengan waktu saat ini
- Jika ada jadwal yang cocok, akan trigger command secara otomatis

**Catatan:** Untuk testing scheduler, Anda bisa:

1. Buat jadwal dengan waktu yang dekat (misal: 1-2 menit dari sekarang)
2. Tunggu hingga waktu tersebut
3. Cek history untuk melihat apakah scheduler berjalan

## Clear Demo Data

Untuk menghapus semua data demo:

```bash
curl -X POST http://localhost:8080/api/v1/demo/clear
```

## Perbedaan Demo Mode vs Production Mode

| Fitur             | Demo Mode                | Production Mode     |
| ----------------- | ------------------------ | ------------------- |
| MQTT Connection   | ❌ Tidak diperlukan      | ✅ Diperlukan       |
| Device Simulation | ✅ Otomatis              | ❌ Real device      |
| Response Time     | ⚡ Instant (simulated)   | ⏱️ Real device time |
| Stock Update      | ✅ Otomatis setelah feed | ✅ Via MQTT report  |

## Contoh Workflow Demo

1. **Seed data:**

   ```bash
   curl -X POST http://localhost:8080/api/v1/demo/seed
   ```

2. **Cek dashboard:**

   ```bash
   curl http://localhost:8080/api/v1/dashboard
   ```

3. **Manual feed:**

   ```bash
   curl -X POST http://localhost:8080/api/v1/feeder/manual
   ```

   Tunggu 5 detik, lalu cek dashboard lagi untuk melihat stock berkurang.

4. **Manual UV:**

   ```bash
   curl -X POST http://localhost:8080/api/v1/uv/manual \
     -H "Content-Type: application/json" \
     -d '{"duration_minutes": 2}'
   ```

   Monitor status UV setiap beberapa detik untuk melihat countdown.

5. **Cek history:**
   ```bash
   curl http://localhost:8080/api/v1/history?limit=5
   ```

## Troubleshooting

### Server tidak start

- Pastikan port 8080 tidak digunakan
- Atau ubah `SERVER_PORT` di `.env`

### Database error

- Pastikan `DB_TYPE=sqlite` untuk demo mode
- Database file akan dibuat otomatis di `aquarium_db.db`
- Jika mendapat error "CGO_ENABLED=0", pastikan menggunakan driver SQLite pure Go (sudah diupdate di kode)

### Scheduler tidak jalan

- Scheduler berjalan setiap menit
- Pastikan waktu server sesuai dengan jadwal yang dibuat
- Cek log untuk melihat aktivitas scheduler
