# Railway Deployment Guide

## 📋 Environment Variables untuk Railway

Copy dan paste environment variables berikut ke **Railway Dashboard** > **Variables**:

### Method 1: Set Manual (One by One)

```
SERVER_PORT=8080
DB_TYPE=postgres
DEMO_MODE=false
MQTT_BROKER=5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud
MQTT_PORT=8883
MQTT_USER=ghalytsarhivemq
MQTT_PASS=Ember1233
MQTT_CLIENT_ID=aquarium-backend
```

**PENTING untuk Database**: Railway sudah auto-inject variable dari Postgres service:

- `${{Postgres.PGHOST}}`
- `${{Postgres.PGPORT}}`
- `${{Postgres.PGUSER}}`
- `${{Postgres.PGPASSWORD}}`
- `${{Postgres.PGDATABASE}}`

Maka tambahkan reference variable berikut di Railway:

```
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_NAME=${{Postgres.PGDATABASE}}
```

### Method 2: Raw Editor (Copy All)

Buka **Raw Editor** di Railway Variables dan paste ini:

```
SERVER_PORT=8080
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_NAME=${{Postgres.PGDATABASE}}
DB_TYPE=postgres
DEMO_MODE=false
MQTT_BROKER=5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud
MQTT_PORT=8883
MQTT_USER=ghalytsarhivemq
MQTT_PASS=Ember1233
MQTT_CLIENT_ID=aquarium-backend
```

---

## 🔍 Verifikasi Setelah Deploy

Cek log Railway, harusnya muncul:

```
✅ Timezone set to Asia/Jakarta (WIB)
MQTT TLS enabled for secure connection
Connected to MQTT broker
Subscribed to topic: aquarium/feeder/status
Subscribed to topic: aquarium/uv/status
Subscribed to topic: aquarium/device/report
Subscribed to topic: aquarium/sensor/dht
Server starting on port 8080
```

---

## 🏠 Local Development vs 🚀 Production

### Local (.env file):

- ✅ Gunakan **SQLite** (ringan, tidak perlu setup DB)
- ✅ File `.env` sudah di-gitignore (aman)

### Railway (Environment Variables):

- ✅ Gunakan **PostgreSQL** (database Railway)
- ✅ MQTT aktif (koneksi ke HiveMQ Cloud)
- ✅ Reference variable `${{Postgres.XXX}}` untuk auto-inject dari Postgres service

---

## 📌 Yang Sudah Lengkap:

✅ **Server Port**: 8080  
✅ **Database**: PostgreSQL (Railway) dengan reference variable  
✅ **MQTT**: HiveMQ Cloud credential  
✅ **Mode**: Production (DEMO_MODE=false)

**Tidak ada yang kurang!** Semua variable sudah lengkap. 🎉
