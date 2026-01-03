Ya, saya sudah membaca dan memahami isi dari **`openapi.yaml`** dan **`PRD.md` (Versi 1.1)** yang Anda unggah. Dokumen tersebut sudah sangat rapi mendefinisikan sistem Feeder dan UV.

Terkait penambahan komponen **DHT (Suhu & Kelembaban)** dan **RTC (Real Time Clock)**, jawabannya adalah: **YA, PERLU PENAMBAHAN BACKEND**, terutama untuk sensor DHT.

Berikut adalah analisis dampak dan perubahan yang perlu dilakukan pada **Schema Database**, **MQTT**, dan **API Backend**:

-----

### 1\. Analisis Perubahan per Komponen

#### A. Sensor DHT (Suhu & Kelembaban)

Ini berdampak besar pada backend karena data suhu biasanya perlu **dipantau (monitoring)** secara real-time dan **disimpan (logging)** untuk melihat grafik perubahan suhu air/sekitar.

  * **Backend:** Perlu tabel baru untuk menyimpan riwayat suhu.
  * **API:** Perlu endpoint untuk mengambil suhu saat ini (untuk dashboard) dan history (untuk grafik).
  * **MQTT:** Perlu topik baru agar ESP32 bisa mengirim data sensor secara berkala.

#### B. RTC (Real Time Clock)

Modul ini dampaknya lebih banyak di **Hardware (ESP32)** daripada Backend.

  * **Fungsi Utama:** Agar ESP32 tetap tahu jam berapa sekarang meskipun **OFFLINE** (tidak ada internet/koneksi ke Backend). Ini krusial agar jadwal pakan/UV tetap jalan saat internet mati.
  * **Relasi ke Backend:** Backend tidak perlu menyimpan data RTC. Backend hanya perlu mengirimkan "Waktu Server" sekali-sekali (via MQTT) untuk *sinkronisasi* agar jam di RTC akurat (mencegah *time drift*).

-----

### 2\. Update Schema Database

Kita tidak perlu mengubah tabel `pakan_schedules` atau `uv_schedules`. Kita hanya perlu menambahkan **satu tabel baru** untuk log sensor.

**Tabel Baru: `sensor_logs`**
Tabel ini akan menyimpan data yang dikirim oleh DHT secara berkala (misal setiap 10 atau 30 menit).

| Kolom | Tipe Data | Keterangan |
| :--- | :--- | :--- |
| `id` | Integer (PK) | ID unik |
| `temperature` | Float/Decimal | Suhu (Celcius), misal `28.5` |
| `humidity` | Float/Decimal | Kelembaban (%), misal `60.2` (Opsional jika hanya butuh suhu) |
| `recorded_at` | Timestamp | Waktu pencatatan |

**SQL Draft (PostgreSQL/MySQL):**

```sql
CREATE TABLE sensor_logs (
    id SERIAL PRIMARY KEY,
    temperature DECIMAL(5,2) NOT NULL,
    humidity DECIMAL(5,2),
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexing penting untuk query grafik agar cepat
CREATE INDEX idx_sensor_time ON sensor_logs(recorded_at);
```

-----

### 3\. Update Arsitektur MQTT

Perlu penambahan topik komunikasi untuk sensor dan sinkronisasi waktu.

**Topik Baru:**

| Topic | Method | Payload Example | Deskripsi |
| :--- | :--- | :--- | :--- |
| `aquarium/sensor/dht` | PUBLISH (Alat) | `{"temp": 28.5, "hum": 65, "rtc_time": "14:00:00"}` | Alat melapor suhu & waktu RTC lokal ke server. |
| `aquarium/time/sync` | SUBSCRIBE (Alat) | `{"epoch": 1732630000}` | Server mengirim waktu standar ke alat untuk update RTC (agar akurat). |

*Catatan:* Saya menyarankan `rtc_time` dikirim bersamaan dengan data sensor hanya untuk *monitoring* di dashboard (memastikan jam di alat tidak ngaco), tapi tidak perlu disimpan ke database.

-----

### 4\. Update API Spec (OpenAPI)

Anda perlu menambahkan/memodifikasi endpoint berikut di `openapi.yaml`:

**A. Modifikasi Dashboard (`GET /dashboard`)**
Tambahkan field `environment` untuk menampilkan suhu terkini.

```yaml
# Tambahan di response schema DashboardResponse
environment:
  temperature: 28.5
  humidity: 60
  last_updated: "2025-11-26T10:00:00Z"
  device_time_status: "SYNCED" # Opsional, hasil cek RTC vs Server Time
```

**B. Endpoint Baru: Sensor History (`GET /sensors/history`)**
Digunakan untuk membuat grafik suhu di Frontend.

```yaml
paths:
  /sensors/history:
    get:
      summary: Get temperature history
      parameters:
        - name: period
          in: query
          schema:
            type: string
            enum: [24h, 7d, 30d] # Filter rentang waktu
      responses:
        '200':
          content:
            application/json:
              example:
                data:
                  - { time: "08:00", temp: 27.5 }
                  - { time: "09:00", temp: 28.0 }
                  - { time: "10:00", temp: 28.2 }
```

-----

### 5\. Logika Kerja Hardware (ESP32) dengan RTC & DHT

Agar Anda mendapat gambaran alur coding di ESP32 nanti:

1.  **Saat Booting (Nyalakan Alat):**
      * Cek koneksi WiFi/MQTT.
      * Jika **Connected**: Minta jam dari Server (`aquarium/time/sync`) -\> Update jam RTC.
      * Jika **Disconnected**: Baca jam dari modul RTC.
2.  **Loop Utama:**
      * Baca sensor DHT.
      * Tampilkan Jam RTC & Suhu di layar LCD/OLED (Local Display).
      * Cek Jadwal Pakan/UV (Bandingkan jam RTC dengan Jadwal).
3.  **Reporting (Misal tiap 5 menit):**
      * Kirim data suhu ke MQTT `aquarium/sensor/dht` untuk disimpan Backend.

### Kesimpulan untuk Developer

1.  **Backend Dev:** Buat tabel `sensor_logs` dan worker MQTT untuk subscribe ke `aquarium/sensor/dht` lalu insert datanya ke tabel tersebut.
2.  **API:** Update endpoint Dashboard dan buat endpoint History Suhu.
3.  **Hardware Dev:** Tambahkan library `DHT` dan `RTClib`. Pastikan logika jadwal sekarang mengacu pada waktu RTC, bukan hanya waktu server (agar fitur *offline* berjalan).
 