#include <Wire.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include <RTClib.h>
#include <ESP32Servo.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>
#include <DHT.h>

// ==========================================
// 1. KONFIGURASI KONEKSI (EDIT DISINI)
// ==========================================
const char* ssid = "dell";
const char* password = "deli1234";

// Konfigurasi HiveMQ (Sesuai Gambar Anda)
const char* mqtt_server = "5aed062215ab4b2f950052625f538fef.s1.eu.hivemq.cloud";
const int mqtt_port = 8883; // Port SSL/TLS
const char* mqtt_user = "ghalytsarhivemq"; // Username dari gambar 1
const char* mqtt_pass = "Ember1233";       // Password dari gambar 1

// Topik MQTT (Sesuai PRD)
const char* topic_feeder_cmd = "aquarium/feeder/command";
const char* topic_uv_cmd = "aquarium/uv/command";
const char* topic_report = "aquarium/device/report";
const char* topic_sensor = "aquarium/sensor/dht";

// ==========================================
// 2. KONFIGURASI PIN & HARDWARE
// ==========================================
#define RELAY_PIN 5
#define SERVO1_PIN 12 // Gate Atas (P2)
#define SERVO2_PIN 14 // Gate Bawah (P4)
#define DHTPIN 15
#define DHTTYPE DHT22

// OLED
#define SCREEN_WIDTH 128
#define SCREEN_HEIGHT 64
Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &Wire, -1);

// Objects
Servo servo1;
Servo servo2;
RTC_DS3231 rtc;
DHT dht(DHTPIN, DHTTYPE);
WiFiClientSecure espClient; // Gunakan Secure Client untuk HiveMQ
PubSubClient client(espClient);

// Variabel Global
float suhu = 0, hum = 0;
unsigned long lastMsg = 0;
unsigned long lastSensor = 0;

// ==========================================
// 3. LOGIKA FEEDER (STATE MACHINE)
// ==========================================
bool isFeeding = false;
int feedStep = 0;
unsigned long feedTimer = 0;
int targetDose = 1; // Default 1 kali takaran

void setup() {
  Serial.begin(115200);
  
  // Init Hardware
  pinMode(RELAY_PIN, OUTPUT);
  digitalWrite(RELAY_PIN, HIGH); // Default OFF
  
  servo1.attach(SERVO1_PIN);
  servo2.attach(SERVO2_PIN);
  servo1.write(90); // Posisi Tutup
  servo2.write(90); // Posisi Tutup
  
  dht.begin();
  
  if (!rtc.begin()) {
    Serial.println("RTC Error!");
    while (1);
  }

  // SET WAKTU SEKALI SAJA (UPLOAD SEKALI)
  rtc.adjust(DateTime(__DATE__, __TIME__));
  Serial.println("RTC time set from compile time");
  // ===================

  if (!display.begin(SSD1306_SWITCHCAPVCC, 0x3C)) {
    Serial.println("OLED Error!");
    for(;;);
  }
  display.clearDisplay();
  display.setTextColor(SSD1306_WHITE);
  
  // Init WiFi & MQTT
  setup_wifi();
  
  // PENTING: Untuk HiveMQ Cloud, kita set Insecure agar tidak ribet dengan sertifikat CA
  // Untuk production grade, sebaiknya upload sertifikat Root CA
  espClient.setInsecure(); 
  
  client.setServer(mqtt_server, mqtt_port);
  client.setCallback(callback);
}

void setup_wifi() {
  delay(10);
  Serial.println();
  Serial.print("Connecting to ");
  Serial.println(ssid);
  
  WiFi.begin(ssid, password);
  
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\nWiFi connected");
}

// Callback saat ada pesan masuk dari MQTT
void callback(char* topic, byte* payload, unsigned int length) {
  String message;
  for (int i = 0; i < length; i++) {
    message += (char)payload[i];
  }
  Serial.print("Pesan masuk ["); Serial.print(topic); Serial.print("]: ");
  Serial.println(message);

  // Parse JSON
  StaticJsonDocument<200> doc;
  DeserializationError error = deserializeJson(doc, message);

  if (error) {
    Serial.print("JSON Error: ");
    Serial.println(error.c_str());
    return;
  }

  // Cek Topik: FEEDER
  if (String(topic) == topic_feeder_cmd) {
    // Payload contoh: {"action": "FEED", "dose": 1}
    const char* action = doc["action"];
    if (strcmp(action, "FEED") == 0) {
      if (!isFeeding) {
        Serial.println(">>> START FEEDING SEQUENCE");
        isFeeding = true;
        feedStep = 0; // Mulai dari langkah awal
        targetDose = doc["dose"] | 1; // Default 1 jika tidak ada
      } else {
        Serial.println(">>> BUSY: Masih proses feeding sebelumnya");
      }
    }
  }

  // Cek Topik: UV
  if (String(topic) == topic_uv_cmd) {
    // Payload contoh: {"state": "ON"}
    const char* state = doc["state"];
    if (strcmp(state, "ON") == 0) {
      digitalWrite(RELAY_PIN, LOW);
      Serial.println(">>> UV MANUAL ON");
    } else {
      digitalWrite(RELAY_PIN, HIGH);
      Serial.println(">>> UV MANUAL OFF");
    }
    // Kirim status update segera
    publishStatusUV();
  }
}

void reconnect() {
  while (!client.connected()) {
    Serial.print("Mencoba koneksi MQTT HiveMQ...");
    String clientId = "ESP32Client-";
    clientId += String(random(0xffff), HEX);
    
    if (client.connect(clientId.c_str(), mqtt_user, mqtt_pass)) {
      Serial.println("Terhubung!");
      // Subscribe ke topik perintah
      client.subscribe(topic_feeder_cmd);
      client.subscribe(topic_uv_cmd);
    } else {
      Serial.print("Gagal, rc=");
      Serial.print(client.state());
      Serial.println(" coba lagi dalam 5 detik");
      delay(5000);
    }
  }
}

// Logika Double Gate (Non-Blocking)
void handleFeeder() {
  if (!isFeeding) return;

  unsigned long currentMillis = millis();

  switch (feedStep) {
    case 0: // BUKA GATE ATAS (P2)
      servo1.write(0); // Buka
      servo2.write(90); // Pastikan Bawah Tutup
      Serial.println("Step 1: Isi Takaran (P2 Buka)");
      feedTimer = currentMillis;
      feedStep = 1;
      break;

    case 1: // TUNGGU ISI (4 Detik)
      if (currentMillis - feedTimer >= 4000) {
        servo1.write(90); // Tutup P2
        Serial.println("Step 2: Kunci Takaran (P2 Tutup)");
        feedTimer = currentMillis;
        feedStep = 2;
      }
      break;

    case 2: // JEDA STABILISASI (1 Detik)
      if (currentMillis - feedTimer >= 1000) {
        servo2.write(0); // Buka P4 (Jatuhkan ke kolam)
        Serial.println("Step 3: Jatuhkan Pakan (P4 Buka)");
        feedTimer = currentMillis;
        feedStep = 3;
      }
      break;

    case 3: // TUNGGU JATUH (3 Detik)
      if (currentMillis - feedTimer >= 3000) {
        servo2.write(90); // Tutup P4
        Serial.println("Step 4: Selesai (P4 Tutup)");
        
        // Kirim Laporan Sukses ke Backend
        StaticJsonDocument<200> doc;
        doc["result"] = "SUCCESS";
        doc["type"] = "FEED";
        doc["feed_gram"] = 10; // Estimasi
        char buffer[256];
        serializeJson(doc, buffer);
        client.publish(topic_report, buffer);
        
        isFeeding = false; // Reset state
        feedStep = 0;
      }
      break;
  }
}

void publishSensor() {
  StaticJsonDocument<200> doc;
  doc["temp"] = suhu;
  doc["hum"] = hum;
  
  // Ambil waktu RTC lokal
  DateTime now = rtc.now();
  char timeBuffer[10];
  sprintf(timeBuffer, "%02d:%02d", now.hour(), now.minute());
  doc["rtc_time"] = timeBuffer;

  char buffer[256];
  serializeJson(doc, buffer);
  client.publish(topic_sensor, buffer);
}

void publishStatusUV() {
  StaticJsonDocument<200> doc;
  // Active LOW relay: LOW = ON, HIGH = OFF
  doc["state"] = (digitalRead(RELAY_PIN) == LOW) ? "ON" : "OFF";
  doc["remaining"] = 0; // ESP tidak track remaining time
  char buffer[256];
  serializeJson(doc, buffer);
  client.publish("aquarium/uv/status", buffer); // Topik status UV
  Serial.print("UV Status Published: ");
  Serial.println(buffer);
}

void updateOLED() {
   display.clearDisplay();
   DateTime now = rtc.now();
   
   // Header
   display.setCursor(0, 0);
   display.printf("%02d/%02d %02d:%02d", now.day(), now.month(), now.hour(), now.minute());
   display.drawLine(0, 10, 128, 10, SSD1306_WHITE);
   
   // Status
   display.setCursor(0, 15);
   display.printf("Temp: %.1f C", suhu);
   display.setCursor(0, 25);
   display.printf("Hum : %.1f %%", hum);
   
   display.setCursor(0, 45);
   if (isFeeding) {
     display.print("STATUS: FEEDING...");
   } else {
     display.print("STATUS: READY");
   }
   
   display.setCursor(0, 55);
   display.print("UV: ");
   // Active LOW relay: LOW = ON, HIGH = OFF
   if(digitalRead(RELAY_PIN) == LOW) display.print("ON"); else display.print("OFF");
   
   display.display();
}

void loop() {
  if (!client.connected()) {
    reconnect();
  }
  client.loop(); // Wajib dipanggil agar MQTT tetap hidup

  unsigned long now = millis();

  // 1. Baca Sensor & Kirim Data (Setiap 10 Detik)
  if (now - lastSensor > 10000) {
    lastSensor = now;
    suhu = dht.readTemperature();
    hum = dht.readHumidity();
    
    // Kirim ke MQTT
    if (!isnan(suhu)) {
      publishSensor();
    }
    // Update Layar
    updateOLED(); 
  }

  // 2. Jalankan Logika Feeder (Jika aktif)
  handleFeeder();
}