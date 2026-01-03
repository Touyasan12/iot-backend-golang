# UV Control Fix - Active Low Relay

## 🔴 Problem Yang Ditemukan:

### 1. **ESP publishStatusUV() Terbalik**

```cpp
// ❌ SALAH (sebelum fix):
doc["state"] = (digitalRead(RELAY_PIN) == HIGH) ? "ON" : "OFF";

// ✅ BENAR (setelah fix):
doc["state"] = (digitalRead(RELAY_PIN) == LOW) ? "ON" : "OFF";
```

**Kenapa?** Relay Anda **Active LOW**:

- `digitalWrite(RELAY_PIN, LOW)` = Relay ON
- `digitalWrite(RELAY_PIN, HIGH)` = Relay OFF

### 2. **Backend Tidak Auto-OFF Setelah Duration Habis**

Backend kirim command:

```json
{ "state": "ON", "duration_sec": 3600 }
```

ESP nyalakan relay, tapi **tidak ada yang matikan** setelah 3600 detik.

---

## ✅ Solusi yang Diterapkan:

### 1. **Fix ESP Code**

- ✅ `publishStatusUV()` - Sekarang benar (LOW=ON, HIGH=OFF)
- ✅ Display OLED - Sekarang benar
- ✅ Tambah logging untuk debugging

### 2. **Backend Auto-OFF**

Tambah scheduler baru yang jalan setiap 10 detik:

```go
func checkManualUVExpiration() {
    // Cek apakah manual UV sudah expire
    // Jika sudah, kirim command OFF ke ESP
}
```

### 3. **Ignore ESP Status Update**

Backend **tidak lagi update database** dari `aquarium/uv/status` topic.
Backend scheduler yang jadi **source of truth**.

---

## 🔄 Alur Kerja Baru:

### Manual UV (dari Frontend):

```
1. User klik "Turn ON UV" (duration: 60 menit)
2. Backend:
   - Create ActionHistory (status: RUNNING, end_time: now + 60 menit)
   - Publish MQTT: {"state": "ON", "duration_sec": 3600}
   - Update DeviceStatus: state=ON
3. ESP:
   - Terima command
   - digitalWrite(RELAY_PIN, LOW) → Relay ON
   - Publish status: {"state": "ON"}
4. Backend Scheduler (check setiap 10 detik):
   - Cek apakah end_time sudah lewat?
   - Jika YA:
     * Publish MQTT: {"state": "OFF"}
     * Update ActionHistory: status=SUCCESS
     * Update DeviceStatus: state=OFF
5. ESP:
   - Terima command OFF
   - digitalWrite(RELAY_PIN, HIGH) → Relay OFF
```

---

## 🧪 Testing:

### Test Manual UV:

```bash
# POST ke backend
curl -X POST http://localhost:8080/api/uv/manual \
  -H "Content-Type: application/json" \
  -d '{"duration_minutes": 1}'

# Tunggu 1 menit
# Relay harusnya auto-off setelah 1 menit
```

### Expected Log Backend:

```
[MQTT] Publishing UV command: state=ON, duration=60
Manual UV expired (Action ID: 123), sending OFF command
[MQTT] Publishing UV command: state=OFF, duration=0
Manual UV turned OFF successfully
```

### Expected Log ESP:

```
Pesan masuk [aquarium/uv/command]: {"state":"ON","duration_sec":60}
>>> UV MANUAL ON
UV Status Published: {"state":"ON","remaining":0}

... (tunggu 60 detik) ...

Pesan masuk [aquarium/uv/command]: {"state":"OFF","duration_sec":0}
>>> UV MANUAL OFF
UV Status Published: {"state":"OFF","remaining":0}
```

---

## 📝 Files yang Diubah:

1. ✅ `esp.ino` - Fix active-low logic
2. ✅ `scheduler/scheduler.go` - Tambah `checkManualUVExpiration()`
3. ✅ `mqtt/client.go` - Ignore UV status update dari ESP
4. ✅ `handlers/uv.go` - Update device status langsung setelah send command

---

## 🚀 Deploy ke Railway:

```bash
git add .
git commit -m "fix: UV active-low relay and auto-off"
git push origin feat/add-mqtt
```

Pastikan environment variables di Railway sudah benar:

- `DEMO_MODE=false`
- MQTT credentials sudah di-set
