# Setup Environment Variables

## Untuk Development (Lokal)

1. **Copy file `.env.example` menjadi `.env`**:

   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` dan sesuaikan credential Anda**:

   ```env
   MQTT_BROKER=5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud
   MQTT_PORT=8883
   MQTT_USER=ghalytsarhivemq
   MQTT_PASS=Ember1233
   ```

3. **Jalankan backend**:

   ```bash
   go run main.go
   ```

4. **Pastikan muncul log**:
   ```
   ✅ Timezone set to Asia/Jakarta (WIB)
   MQTT TLS enabled for secure connection
   Connected to MQTT broker
   Subscribed to topic: aquarium/feeder/status
   Subscribed to topic: aquarium/uv/status
   Subscribed to topic: aquarium/device/report
   Subscribed to topic: aquarium/sensor/dht
   ```

---

## Untuk Production (Railway)

### Tidak perlu file `.env`! Set langsung di Railway Dashboard:

1. Buka project Anda di Railway
2. Klik tab **"Variables"**
3. Tambahkan environment variables:

```
MQTT_BROKER=5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud
MQTT_PORT=8883
MQTT_USER=ghalytsarhivemq
MQTT_PASS=Ember1233
MQTT_CLIENT_ID=aquarium-backend-prod
DEMO_MODE=false
```

4. **Deploy ulang** atau Railway akan auto-deploy

---

## Keamanan

⚠️ **PENTING**:

- File `.env` sudah masuk `.gitignore`, jadi **tidak akan ter-commit ke Git**
- Jangan pernah commit credential ke repository!
- Untuk production, **gunakan Railway environment variables**

---

## Mode Demo (Tanpa MQTT)

Jika ingin test backend tanpa koneksi MQTT (untuk demo/testing):

```env
DEMO_MODE=true
```

Backend akan menggunakan Mock MQTT (tidak benar-benar kirim perintah).
