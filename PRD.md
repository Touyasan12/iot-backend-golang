**Product Requirements Document (PRD)**
**Nama Proyek:** Smart Aquarium Controller (Auto Feeder & UV System)
**Versi:** 1.2
**Tanggal:** 26 November 2025
**Last Updated:** 26 November 2025

## 1. Ringkasan Proyek (Executive Summary)

Sistem IoT berbasis web untuk memantau dan mengontrol akuarium secara otomatis. Sistem ini memiliki tiga fitur utama:

1.  **Smart Feeder:** Sistem pemberian pakan dengan mekanisme "Double Gate" (Volumetric Dosing) untuk takaran akurat (~10g) dan keamanan stok.
2.  **Smart UV Sterilizer:** Pengontrolan lampu UV pembersih air dengan sistem penjadwalan fleksibel dan mode manual yang memiliki prioritas _override_.
3.  **Environment Monitoring (BARU):** Monitoring temperature dan humidity menggunakan sensor DHT dengan history tracking.

Sistem menggunakan protokol **MQTT** untuk komunikasi antara Hardware (ESP32/Arduino) dan Server.

**Mode Operasi:**

- **Production Mode:** Menggunakan MQTT broker untuk komunikasi dengan hardware
- **Demo Mode:** Simulasi responses tanpa hardware (untuk development/testing)

---

## 2. Fitur Fungsional (Functional Requirements)

### A. Modul Automatic Feeder (Pemberi Pakan)

**1. Mekanisme Fisik (Double Gate Logic)**
Sistem harus mengikuti alur kerja berikut untuk menjatuhkan 1 dosis pakan:

- **State Awal:** Gate Atas (P2) & Bawah (P4) tertutup.
- **Step 1 (Isi Takaran):** Buka P2 selama **X detik** (estimasi 4-6 detik) untuk mengisi ruang takar (P3) dari Stok (P1). P4 tetap tertutup.
- **Step 2 (Kunci):** Tutup P2. Jeda sebentar untuk stabilisasi.
- **Step 3 (Tuang):** Buka P4 untuk menjatuhkan isi P3 (~10g) ke kolam.
- **Step 4 (Reset):** Tutup P4. Kembali ke State Awal.
- _Output:_ Mengurangi nilai stok pakan di database sebesar jumlah yang diberikan (default 10 gram).

**⚡ IMPLEMENTASI TERBARU:**

- Proses feeding dilakukan **INSTANT** tanpa delay (untuk demo mode)
- Stock langsung dikurangi setelah command dikirim
- Status berubah: `PENDING` → `DISPENSING` → `IDLE` → `SUCCESS`

**2. Penjadwalan Pakan (Scheduling)**

- User dapat mengatur jadwal berdasarkan **Hari** (Mon, Tue, Wed, Thu, Fri, Sat, Sun).
- Limitasi: Maksimal 5 kali pemberian pakan per hari.
- Input user: Jam eksekusi (format HH:MM, misal: 08:00, 12:00).
- Input tambahan: Jumlah pakan (amount_gram, default 10g).

**3. Manual Feed & Validasi**

- Terdapat tombol "Beri Pakan Sekarang".
- User dapat menentukan jumlah pakan (amount_gram) saat manual feed.
- **Response API:** Saat tombol ditekan, backend mengembalikan data "last_feed":
  ```json
  {
    "message": "Feeding command sent",
    "action_id": 26,
    "amount_gram": 15,
    "last_feed": {
      "day": "Wednesday",
      "time": "08:00"
    }
  }
  ```
- Frontend dapat menampilkan konfirmasi berdasarkan data `last_feed` ini.
- Jika belum pernah ada feeding, `last_feed` tidak disertakan dalam response.

**4. Last Feed Info Endpoint**

- Endpoint khusus: `GET /api/v1/feeder/last-feed`
- Response format:
  ```json
  {
    "exists": true,
    "day": "Wednesday",
    "time": "08:00",
    "date": "2025-11-26"
  }
  ```
- Digunakan untuk menampilkan informasi pemberian pakan terakhir di UI.

---

### B. Modul UV Sterilizer (Lampu UV)

**1. Penjadwalan UV**

- User dapat mengatur jadwal berdasarkan **Hari** (Mon-Sun) dan **Rentang Waktu** (Start Time - End Time).
- Format waktu: HH:MM (24-hour format).
- Support jadwal overnight (misal: 20:00 - 04:00 melewati tengah malam).
- Default: 7 Hari seminggu, durasi 8 jam (misal 20:00 - 04:00).
- Satu schedule per hari (tidak ada limitasi multiple schedule per hari seperti feeder).

**2. Manual Override (Prioritas)**

- User dapat menyalakan UV secara manual dengan input **Durasi dalam Menit**.
- Input: `duration_minutes` (minimum 1 menit).
- API akan mengkonversi ke detik: `duration_sec = duration_minutes * 60`.
- Response akan memberikan `end_time` sebagai referensi kapan UV akan mati otomatis.
- **Logika Konflik:** Jika Mode Manual sedang aktif (status RUNNING), maka jadwal otomatis yang seharusnya berjalan di jam tersebut harus **diabaikan/dipauses** hingga manual selesai.

**3. Stop Manual UV**

- Endpoint: `POST /api/v1/uv/manual/stop`
- Fungsi: Menghentikan UV yang sedang running (baik MANUAL maupun SCHEDULE)
- ⚡ **IMPLEMENTASI TERBARU:**
  - Dapat menghentikan UV SCHEDULE yang sedang running (tidak hanya MANUAL)
  - Status action berubah menjadi `STOPPED` (bukan `OVERRIDDEN`)
  - Response menyertakan `trigger_source` untuk informasi apakah yang dihentikan MANUAL atau SCHEDULE

**4. UV Status**

- Endpoint: `GET /api/v1/uv/status`
- Menampilkan status real-time UV:
  - `state`: ON/OFF
  - `remaining`: Sisa waktu dalam detik
  - `manual_active`: Boolean, true jika manual UV sedang running
  - `manual_end_time`: Timestamp kapan manual UV akan selesai (jika aktif)

---

### C. Modul Environment Monitoring (BARU)

**1. Sensor DHT (Temperature & Humidity)**

- Hardware: Sensor DHT11/DHT22 terhubung ke ESP32/Arduino
- Data dikirim ke backend melalui MQTT topic `aquarium/sensor/dht`
- Payload format: `{"temperature": 27.5, "humidity": 65.0}`
- Frekuensi pengiriman: Setiap 5 menit
- Data disimpan ke tabel `sensor_logs` untuk history tracking

**2. Current Sensor Readings**

- Endpoint: `GET /api/v1/sensors/current`
- Response: Data sensor terbaru (temperature, humidity, recorded_at)
- Digunakan untuk menampilkan status environment real-time

**3. Sensor History**

- Endpoint: `GET /api/v1/sensors/history`
- Support pagination (page, page_size)
- Support filter time range: 24h, 7d, 30d
- Response format:
  ```json
  {
    "data": [
      {
        "id": 100,
        "temperature": 27.5,
        "humidity": 65.0,
        "recorded_at": "2025-11-19T09:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 50,
      "total": 250,
      "total_pages": 5
    }
  }
  ```

**4. Dashboard Integration**

- Dashboard response termasuk field `environment` dengan data sensor terbaru
- Jika belum ada data sensor, field `environment` bernilai `null`

**5. Demo Mode**

- Mock sensor data generator menghasilkan data simulasi:
  - Temperature: 25-30°C (random realistic values)
  - Humidity: 60-80% (random realistic values)
  - Interval: Setiap 5 menit

---

### D. Dashboard & History

## 3. Arsitektur Sistem & Data Flow

### Alur Komunikasi (MQTT)

Sistem menggunakan arsitektur Pub/Sub. Backend bertindak sebagai pengendali logika.

**1. Topik MQTT (Draft Standar)**

| Topic                     | Method           | Payload (JSON Example)                                   | Deskripsi                                                                 |
| :------------------------ | :--------------- | :------------------------------------------------------- | :------------------------------------------------------------------------ |
| `aquarium/feeder/command` | PUBLISH (Server) | `{"action": "FEED", "dose": 1}`                          | Perintah dari server ke alat untuk memulai siklus Double Gate.            |
| `aquarium/feeder/status`  | PUBLISH (Alat)   | `{"status": "IDLE"}` atau `{"status": "DISPENSING"}`     | Status alat saat ini.                                                     |
| `aquarium/uv/command`     | PUBLISH (Server) | `{"state": "ON", "duration_sec": 7200}`                  | Perintah nyalakan UV. Jika schedule, duration bisa diset 0 (ikut jadwal). |
| `aquarium/uv/status`      | PUBLISH (Alat)   | `{"state": "ON", "remaining": 3600}`                     | Laporan status UV real-time.                                              |
| `aquarium/sensor/dht`     | PUBLISH (Alat)   | `{"temperature": 27.5, "humidity": 65.0}`                | Data sensor DHT (temperature & humidity) dikirim setiap 5 menit.          |
| `aquarium/device/report`  | PUBLISH (Alat)   | `{"result": "SUCCESS", "type": "FEED", "feed_gram": 10}` | Laporan akhir setelah aksi selesai untuk dicatat ke DB.                   |

---

## 4. Spesifikasi Database (Schema Overview)

Sesuai diskusi sebelumnya, berikut adalah struktur tabel yang akan digunakan di Backend.

**1. `pakan_schedules`**

- Menyimpan jadwal rutin pakan.
- Kolom: `id`, `day_name` (Mon-Sun), `time` (HH:MM), `amount_gram` (default 10), `is_active`.

**2. `uv_schedules`**

- Menyimpan jadwal rutin UV.
- Kolom: `id`, `day_name`, `start_time`, `end_time`, `is_active`.

**3. `action_history` (Log Utama)**

- Mencatat semua aktivitas dan digunakan untuk validasi logic.
- Kolom:
  - `id`
  - `device_type`: ('FEEDER', 'UV')
  - `trigger_source`: ('SCHEDULE', 'MANUAL')
  - `start_time`: Timestamp
  - `end_time`: Timestamp (Diisi saat selesai / estimasi selesai untuk UV manual)
  - `status`: ('PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'STOPPED')
  - `value`: Integer dengan arti berbeda per device:
    - **FEEDER**: Jumlah pakan dalam **gram** (10, 15, 20, dll)
    - **UV**: Durasi dalam **detik** (7200 = 2 jam, 28800 = 8 jam, dll)
  - `created_at`: Timestamp
  - `updated_at`: Timestamp
  - `deleted_at`: Timestamp (soft delete, nullable)

**4. `sensor_logs`**

- Menyimpan history pembacaan sensor DHT (temperature & humidity).
- Kolom:
  - `id`
  - `temperature`: Float (Celsius)
  - `humidity`: Float (Percentage)
  - `recorded_at`: Timestamp

**5. `stocks`**

- Menyimpan informasi stock pakan.
- Kolom: `id`, `amount_gram`, `updated_at`.

**6. `device_statuses`**

- Menyimpan status real-time device (UV & Feeder).
- Kolom: `device_type`, `status`, `remaining`, `last_updated`.

**⚡ PERUBAHAN PENTING:**

- Status `OVERRIDDEN` diganti dengan `STOPPED` untuk lebih jelas
- Field `value` untuk UV schedule sekarang berisi durasi aktual dalam detik (bukan 0)
- Perhitungan `value` untuk UV schedule: `(end_time - start_time) in seconds`
- **BARU:** Tabel `sensor_logs` untuk monitoring environment (temperature & humidity)

---

## 5. Kebutuhan Antarmuka (UI/UX)

**Dashboard Utama:**

1.  **Card Status Stok:** Menampilkan sisa pakan (dalam Gram atau Persentase) dengan grafik visual.
2.  **Card Status UV:** Indikator apakah UV sedang Nyala/Mati dan sisa waktu (jika manual).
3.  **Card Status Feeder:** Status feeder (IDLE/DISPENSING) dan last updated.
4.  **Card Environment:** Menampilkan data sensor DHT (temperature & humidity) terbaru.

**⚡ PERUBAHAN:**

- Dashboard **TIDAK** menampilkan history terbaru
- History harus diakses melalui endpoint terpisah: `GET /api/v1/history`
- Dashboard fokus pada status real-time saja
- **BARU:** Dashboard menampilkan environment data (temperature, humidity, last_updated)

**History Page (Terpisah):**

1. **Pagination:** Support pagination dengan parameters:
   - `page`: Nomor halaman (default: 1)
   - `page_size`: Jumlah item per halaman (default: 50, max: 200)
2. **Filter:** Support filtering berdasarkan:
   - `device_type`: FEEDER atau UV
   - `trigger_source`: SCHEDULE atau MANUAL
   - `status`: PENDING, RUNNING, SUCCESS, FAILED, STOPPED
3. **Response Format:**
   ```json
   {
     "data": [...],
     "pagination": {
       "page": 1,
       "page_size": 50,
       "total": 156,
       "total_pages": 4
     }
   }
   ```

**Kontrol Panel:**

1.  **Tombol Manual Feed:**
    - Input: `amount_gram` (optional, default 10g)
    - Response berisi `last_feed` info
    - Frontend dapat menampilkan konfirmasi atau info last feed
2.  **Tombol Manual UV:**
    - Input: `duration_minutes` (required, minimum 1)
    - Menampilkan end_time di response
3.  **Tombol Stop UV:**
    - Muncul ketika UV sedang running (manual atau schedule)
    - Dapat menghentikan UV yang sedang aktif
4.  **Setting Jadwal:**
    - Form untuk menambah/edit/hapus jadwal Pakan dan UV
    - Toggle ON/OFF per schedule (is_active)
    - Pagination untuk list schedules:
      - Feeder schedules: default 20, max 100
      - UV schedules: default 20, max 100

---

## 6. Logika Backend (Business Logic)

**Cron Job / Scheduler (Berjalan setiap menit):**

1.  **Cek Jadwal Pakan:**

    - Apakah ada jadwal di menit ini (day_name + time match)?
    - Cek duplikasi: Apakah sudah ada action dengan status PENDING/RUNNING/SUCCESS dalam 1 menit terakhir?
    - Jika belum -> Create action dengan status PENDING
    - Publish MQTT `aquarium/feeder/command` dengan payload `{"dose": calculated_doses}`
    - Update action status menjadi RUNNING
    - Kurangi stock setelah feeding SUCCESS

2.  **Cek Jadwal UV:**
    - Apakah saat ini masuk rentang waktu jadwal UV (start_time - end_time)?
    - Support jadwal overnight (misal 20:00 - 04:00)
    - _Cek Konflik:_ Apakah ada data di `action_history` dengan:
      - device_type = 'UV'
      - trigger_source = 'MANUAL'
      - status = 'RUNNING'
      - end_time masih di masa depan
    - Jika TIDAK ADA konflik:
      - Cek apakah sudah ada schedule yang running
      - Jika belum -> Create action dengan status RUNNING
      - Calculate end_time dan duration
      - Set value = duration in seconds
      - Publish MQTT `aquarium/uv/command` (ON)
    - Jika ADA konflik -> Skip (biarkan manual berjalan)
    - Di luar rentang jadwal -> Turn OFF UV jika schedule masih running

**⚡ IMPLEMENTASI TERBARU:**

- Value untuk UV schedule berisi durasi aktual (bukan 0)
- Kalkulasi durasi memperhitungkan overnight schedule
- Stop endpoint dapat menghentikan UV schedule (tidak hanya manual)

---

## 7. API Endpoints (REST API)

### Dashboard

- `GET /api/v1/dashboard` - Get dashboard data (stock, UV status, feeder status, environment)
  - **TIDAK** termasuk history (untuk performa)
  - **BARU:** Termasuk environment data (temperature, humidity)

### Feeder

- `GET /api/v1/feeder/schedules` - Get all schedules (dengan pagination)
- `POST /api/v1/feeder/schedules` - Create schedule
- `PUT /api/v1/feeder/schedules/:id` - Update schedule
- `DELETE /api/v1/feeder/schedules/:id` - Delete schedule
- `POST /api/v1/feeder/manual` - Manual feed (dengan amount_gram optional)
- `GET /api/v1/feeder/last-feed` - Get last feed info

### UV

- `GET /api/v1/uv/schedules` - Get all schedules (dengan pagination)
- `POST /api/v1/uv/schedules` - Create schedule
- `PUT /api/v1/uv/schedules/:id` - Update schedule
- `DELETE /api/v1/uv/schedules/:id` - Delete schedule
- `POST /api/v1/uv/manual` - Manual UV (dengan duration_minutes)
- `POST /api/v1/uv/manual/stop` - Stop running UV (manual atau schedule)
- `GET /api/v1/uv/status` - Get UV status

### History

- `GET /api/v1/history` - Get action history (dengan pagination & filters)
  - Params: `page`, `page_size`, `device_type`, `trigger_source`, `status`

### Stock

- `GET /api/v1/stock` - Get current stock
- `PUT /api/v1/stock` - Update stock

### Sensors (BARU)

- `GET /api/v1/sensors/current` - Get current sensor readings (temperature & humidity)
- `GET /api/v1/sensors/history` - Get sensor history (dengan pagination)
  - Params: `page`, `page_size`, `range` (24h, 7d, 30d)

### Demo

- `POST /api/v1/demo/seed` - Seed demo data
- `POST /api/v1/demo/clear` - Clear demo data

**Dokumentasi Lengkap:**

- OpenAPI Spec: `/openapi.yaml`
- Swagger UI: `/docs`
- Pagination Guide: `docs/PAGINATION.md`
- History Value Field: `docs/HISTORY_VALUE_FIELD.md`

---

## 8. MQTT Topics (Hardware Communication)

| Topic                     | Method             | Payload Example                                          | Deskripsi                                      |
| :------------------------ | :----------------- | :------------------------------------------------------- | :--------------------------------------------- |
| `aquarium/feeder/command` | PUBLISH (Server)   | `{"action": "FEED", "dose": 1}`                          | Perintah feeding ke hardware                   |
| `aquarium/feeder/status`  | PUBLISH (Hardware) | `{"status": "IDLE"}` atau `{"status": "DISPENSING"}`     | Status feeder real-time                        |
| `aquarium/uv/command`     | PUBLISH (Server)   | `{"state": "ON", "duration_sec": 7200}`                  | Perintah UV (duration_sec 0 = follow schedule) |
| `aquarium/uv/status`      | PUBLISH (Hardware) | `{"state": "ON", "remaining": 3600}`                     | Status UV real-time dengan remaining time      |
| `aquarium/sensor/dht`     | PUBLISH (Hardware) | `{"temperature": 27.5, "humidity": 65.0}`                | Data sensor DHT (BARU)                         |
| `aquarium/device/report`  | PUBLISH (Hardware) | `{"result": "SUCCESS", "type": "FEED", "feed_gram": 10}` | Report hasil eksekusi                          |

---

## 9. Pagination Implementation

Pagination telah diimplementasikan pada endpoints berikut:

### History

- Default page_size: 50
- Max page_size: 200
- Sorting: `created_at DESC`

### Feeder Schedules

- Default page_size: 20
- Max page_size: 100
- Sorting: `day_name, time`

### UV Schedules

- Default page_size: 20
- Max page_size: 100
- Sorting: `day_name, start_time`

### Sensor History (BARU)

- Default page_size: 50
- Max page_size: 200
- Sorting: `recorded_at DESC`
- Filter: `range` parameter (24h, 7d, 30d)

**Response Format:**

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

---

## 10. Next Step

Dokumen ini siap digunakan sebagai acuan.

1.  **Backend Dev:** Setup database dan MQTT Broker, buat Cron Job.
2.  **Hardware Dev:** Coding ESP32 untuk subscribe topic MQTT dan menggerakkan servo (P2 & P4) sesuai urutan logic.
3.  **Frontend Dev:** Buat UI Dashboard dan logika fetch data history untuk popup.
