# MQTT Troubleshooting Guide - Railway Deployment

## 🔍 Checklist: Mengapa MQTT Tidak Kirim ke Hardware?

### 1️⃣ Cek Log Railway

Buka Railway Dashboard > Deployments > View Logs, pastikan muncul:

```
✅ Connected to MQTT broker successfully!
✅ Subscribed to topic: aquarium/feeder/status
✅ Subscribed to topic: aquarium/uv/status
✅ Subscribed to topic: aquarium/device/report
✅ Subscribed to topic: aquarium/sensor/dht
📡 All MQTT subscriptions completed!
```

**Jika TIDAK muncul**, lanjut ke step 2.

---

### 2️⃣ Cek Environment Variables di Railway

Pastikan semua variable ini sudah di-set di **Railway Dashboard > Variables**:

```
MQTT_BROKER=5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud
MQTT_PORT=8883
MQTT_USER=ghalytsarhivemq
MQTT_PASS=Ember1233
MQTT_CLIENT_ID=aquarium-backend-prod
DEMO_MODE=false    👈 PENTING! Harus false
```

**Jika DEMO_MODE=true**, backend akan pakai Mock MQTT (tidak kirim ke HiveMQ).

---

### 3️⃣ Test Kirim Command Manual

Setelah deploy, coba trigger feeding/UV dari frontend. Lihat log Railway:

**Yang Diharapkan:**

```
📤 Publishing to aquarium/feeder/command: {"action":"FEED","dose":1}
✅ Published feeder command: dose=1
```

**Jika muncul ini = MQTT ERROR:**

```
❌ MQTT Client not connected!
❌ Failed to publish feeder command: ...
```

**Jika muncul ini = DEMO MODE aktif:**

```
⚠️  MOCK MODE: Simulating feeder command
```

---

### 4️⃣ Cek ESP32 Subscribed ke Topic yang Benar?

ESP32 harus subscribe ke:

- `aquarium/feeder/command` (untuk terima perintah pakan)
- `aquarium/uv/command` (untuk terima perintah UV)

**Test di HiveMQ Web Client:**

1. Login ke https://console.hivemq.cloud/
2. Buka Web Client
3. Subscribe ke `aquarium/feeder/command`
4. Dari frontend, klik tombol "Feed Now"
5. Harusnya muncul message: `{"action":"FEED","dose":1}`

**Jika tidak muncul** = Backend tidak publish (cek log Railway)
**Jika muncul tapi ESP32 tidak respon** = ESP32 belum subscribe / offline

---

### 5️⃣ Cek ClientID Conflict

Jika ada **2 client dengan ClientID sama**, HiveMQ akan disconnect salah satu.

**Solusi:**

- Railway backend: `MQTT_CLIENT_ID=aquarium-backend-prod`
- ESP32: `clientId = "ESP32-Aquarium"` (harus beda!)

---

## 🛠️ Quick Fix Commands

### Rebuild & Redeploy Railway

```bash
git add .
git commit -m "feat: add detailed MQTT logging"
git push origin feat/add-mqtt
```

### Check Railway Logs

```bash
railway logs
```

---

## 📊 Expected Log Output (Success)

```
2025/12/11 10:00:00 ✅ Timezone set to Asia/Jakarta (WIB)
2025/12/11 10:00:01 Database connected successfully
2025/12/11 10:00:02 🔌 Connecting to MQTT broker: tls://5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud:8883 (Client ID: aquarium-backend-prod)
2025/12/11 10:00:02 🔐 MQTT User: ghalytsarhivemq
2025/12/11 10:00:02 🔑 MQTT Password: [SET]
2025/12/11 10:00:02 🔒 MQTT TLS enabled for secure connection
2025/12/11 10:00:02 ⏳ Connecting to MQTT broker...
2025/12/11 10:00:03 ✅ Connected to MQTT broker successfully!
2025/12/11 10:00:03 📡 Subscribing to MQTT topics...
2025/12/11 10:00:03 ✅ Subscribed to topic: aquarium/feeder/status
2025/12/11 10:00:03 ✅ Subscribed to topic: aquarium/uv/status
2025/12/11 10:00:03 ✅ Subscribed to topic: aquarium/device/report
2025/12/11 10:00:03 ✅ Subscribed to topic: aquarium/sensor/dht
2025/12/11 10:00:03 📡 All MQTT subscriptions completed!
2025/12/11 10:00:03 Scheduler started
2025/12/11 10:00:03 Server starting on port 8080
```

---

## 🐛 Common Issues

### Issue 1: "MQTT Client not connected"

**Penyebab:** Environment variables salah atau DEMO_MODE=true
**Solusi:** Set ulang env vars di Railway

### Issue 2: Backend connected tapi ESP32 tidak terima

**Penyebab:** ESP32 tidak subscribe atau topic salah
**Solusi:** Cek code ESP32, pastikan subscribe ke `aquarium/feeder/command`

### Issue 3: "Failed to connect to MQTT broker"

**Penyebab:**

- Username/password salah
- Broker URL salah
- Port salah (harus 8883 untuk TLS)
  **Solusi:** Cek credential di Railway Variables

---

## 📞 Next Steps

1. Push perubahan code ini ke Railway
2. Check log Railway setelah deploy
3. Copy & paste log ke chat jika masih error
4. Test trigger feeding dari frontend
5. Monitor log saat trigger untuk melihat MQTT publish
