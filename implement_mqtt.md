Mantap\! Hardware sudah online, HiveMQ sudah siap. Sekarang saatnya menyambungkan "otak" utama (Backend Go) ke sistem syaraf (MQTT) ini.

Karena backend Anda menggunakan **Go (Golang)**, kita membutuhkan library **Paho MQTT**.

Berikut adalah langkah-langkah detail untuk mengintegrasikan HiveMQ ke Backend Anda:

### 1\. Install Library MQTT untuk Go

Buka terminal di folder project backend Anda, lalu jalankan:

```bash
go get github.com/eclipse/paho.mqtt.golang
```

### 2\. Buat Service MQTT Wrapper

Agar rapi, jangan taruh kode MQTT di `main.go` semua. Buatlah package terpisah.
Misal: buat folder `pkg/mqtt_client/` dan file `mqtt_client.go`.

Berikut adalah kode lengkap untuk meng-handle koneksi **Secure (TLS)** ke HiveMQ Cloud:

**File: `pkg/mqtt_client/client.go`**

```go
package mqtt_client

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var Client mqtt.Client

// InitMQTT menginisialisasi koneksi ke HiveMQ
func InitMQTT() {
	// Ambil config dari Env atau hardcode sementara (sesuai data HiveMQ Anda)
	broker := "5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud"
	port := 8883
	user := "ghalytsarhivemq"
	password := "Ember1233"

	// PENTING: Gunakan tls:// karena port 8883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", broker, port))
	opts.SetClientId("Backend-Service-Golang")
	opts.SetUsername(user)
	opts.SetPassword(password)

	// Config TLS (Set InsecureSkipVerify true agar mirip dengan ESP32 setInsecure)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts.SetTLSConfig(tlsConfig)

	// Auto Reconnect jika putus
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)

	// Handler saat koneksi berhasil
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("[MQTT] Connected to HiveMQ Cloud!")
		
		// Subscribe ke topik laporan dari alat (Report & Sensor)
		SubscribeTopic("aquarium/device/report")
		SubscribeTopic("aquarium/sensor/dht")
	})

	// Handler saat koneksi putus
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("[MQTT] Connection Lost: %v", err)
	})

	// Connect
	Client = mqtt.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[MQTT] Error connecting: %s", token.Error())
	}
}

// PublishCommand mengirim perintah ke alat
func PublishCommand(topic string, payload string) {
	if Client == nil || !Client.IsConnected() {
		log.Println("[MQTT] Client not connected, cannot publish")
		return
	}
	token := Client.Publish(topic, 1, false, payload)
	token.Wait()
	log.Printf("[MQTT] Published to %s: %s", topic, payload)
}

// SubscribeTopic mendengarkan pesan dari alat
func SubscribeTopic(topic string) {
	token := Client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("[MQTT] Received from %s: %s", msg.Topic(), msg.Payload())
		
		// DISINI LOGIKA MASUK DATABASE (Update History / Sensor Log)
		// Nanti kita sambungkan ke Handler Database Anda
		// handleIncomingMessage(msg.Topic(), msg.Payload())
	})
	token.Wait()
	log.Printf("[MQTT] Subscribed to %s", topic)
}
```

### 3\. Panggil Init di `main.go`

Buka file `main.go` Anda, dan tambahkan inisialisasi ini **sebelum** menjalankan server HTTP (Gin/Fiber/Echo).

```go
package main

import (
    "your-project/pkg/mqtt_client" // Sesuaikan nama module Anda
    // ... import lainnya
)

func main() {
    // 1. Init Database
    // database.Connect() ...

    // 2. Init MQTT (Jalankan ini)
    mqtt_client.InitMQTT()

    // 3. Setup Router & Server
    // r := gin.Default() ...
}
```

### 4\. Integrasi: Mengirim Perintah dari Scheduler

Sekarang, mari kita ubah `scheduler.go` Anda agar tidak hanya update database, tapi juga **mengirim perintah nyata** ke ESP32.

**Contoh Update di `scheduler.go`:**

```go
import (
    "encoding/json"
    "your-project/pkg/mqtt_client" // Import wrapper yang kita buat tadi
)

// ... di dalam fungsi pengecekan jadwal Pakan ...

if shouldFeed {
    // 1. Logic Database Anda (Create Action History PENDING)
    // ...
    
    // 2. KIRIM MQTT KE ALAT
    // Buat payload JSON
    payloadMap := map[string]interface{}{
        "action": "FEED",
        "dose":   1, // Atau ambil dari jadwal
    }
    payloadBytes, _ := json.Marshal(payloadMap)
    
    // Publish!
    mqtt_client.PublishCommand("aquarium/feeder/command", string(payloadBytes))
    
    log.Println("Perintah Pakan dikirim via MQTT!")
}

// ... di dalam fungsi pengecekan jadwal UV ...

if shouldTurnOnUV {
    // 1. Logic Database ...
    
    // 2. KIRIM MQTT
    payloadMap := map[string]string{
        "state": "ON",
    }
    payloadBytes, _ := json.Marshal(payloadMap)
    
    mqtt_client.PublishCommand("aquarium/uv/command", string(payloadBytes))
}
```

### 5\. Integrasi: Mengirim Perintah dari API (Manual Button)

Sama seperti scheduler, update juga handler API manual Anda (misal di `handlers/feeder.go`).

```go
func ManualFeed(c *gin.Context) {
    // ... Logic validasi & database ...

    // KIRIM MQTT LANGSUNG
    payload := `{"action":"FEED", "dose":1}`
    mqtt_client.PublishCommand("aquarium/feeder/command", payload)

    c.JSON(200, gin.H{"message": "Feeding command sent via MQTT"})
}
```

### Rangkuman Alur Data Baru:

1.  **Scheduler/API** mendeteksi waktu makan.
2.  Backend memanggil `mqtt_client.PublishCommand()`.
3.  **HiveMQ Cloud** meneruskan pesan ke ESP32.
4.  **ESP32** bergerak (Servo) lalu mengirim report balik.
5.  **Backend** (via fungsi `SubscribeTopic`) menerima report dan mengupdate status di database menjadi `SUCCESS`.

Silakan coba implementasikan langkah 1 sampai 3 dulu. Jika koneksi backend ke HiveMQ berhasil, Anda akan melihat log `[MQTT] Connected to HiveMQ Cloud!` di terminal backend Anda.