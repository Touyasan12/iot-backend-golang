package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"iot-backend-cursor/config"
	"iot-backend-cursor/database"
	"iot-backend-cursor/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var Client mqtt.Client

type FeederCommand struct {
	Action string `json:"action"` // FEED
	Dose   int    `json:"dose"`   // number of doses
}

type UVCommand struct {
	State       string `json:"state"`        // ON, OFF
	DurationSec int    `json:"duration_sec"` // duration in seconds (0 for schedule)
}

type FeederStatus struct {
	Status string `json:"status"` // IDLE, DISPENSING
}

type UVStatus struct {
	State     string `json:"state"`     // ON, OFF
	Remaining int    `json:"remaining"` // remaining seconds
}

type DeviceReport struct {
	Result   string `json:"result"`    // SUCCESS, FAILED
	Type     string `json:"type"`      // FEED, UV
	FeedGram int    `json:"feed_gram"` // grams of food (for FEED type)
}

type SensorData struct {
	Temperature float64 `json:"temp"`     // Temperature in Celsius
	Humidity    float64 `json:"hum"`      // Humidity in percentage
	RTCTime     string  `json:"rtc_time"` // RTC time (for monitoring, optional)
}

func InitMQTT(cfg *config.Config) {
	brokerURL := cfg.GetMQTTBrokerURL()
	log.Printf("🔌 Connecting to MQTT broker: %s (Client ID: %s)", brokerURL, cfg.MQTTClientID)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(cfg.MQTTClientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	// Set username and password if provided
	if cfg.MQTTUser != "" {
		opts.SetUsername(cfg.MQTTUser)
		log.Printf("🔐 MQTT User: %s", cfg.MQTTUser)
		if cfg.MQTTPass != "" {
			opts.SetPassword(cfg.MQTTPass)
			log.Printf("🔑 MQTT Password: [SET]")
		}
	} else {
		log.Println("⚠️  MQTT User not set!")
	}

	// Configure TLS for secure connections (port 8883)
	if cfg.MQTTPort == "8883" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Skip certificate verification (like ESP32)
		}
		opts.SetTLSConfig(tlsConfig)
		log.Println("🔒 MQTT TLS enabled for secure connection")
	}

	Client = mqtt.NewClient(opts)

	log.Println("⏳ Connecting to MQTT broker...")
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("❌ Failed to connect to MQTT broker: %v", token.Error())
		log.Fatal("MQTT connection failed")
	}

	log.Println("✅ Connected to MQTT broker successfully!")

	// Subscribe to status topics
	subscribeToTopics()
}

func subscribeToTopics() {
	log.Println("📡 Subscribing to MQTT topics...")
	topics := []string{
		"aquarium/feeder/status",
		"aquarium/uv/status",
		"aquarium/device/report",
		"aquarium/sensor/dht",
	}

	for _, topic := range topics {
		if token := Client.Subscribe(topic, 0, messageHandler); token.Wait() && token.Error() != nil {
			log.Printf("❌ Failed to subscribe to %s: %v", topic, token.Error())
		} else {
			log.Printf("✅ Subscribed to topic: %s", topic)
		}
	}
	log.Println("📡 All MQTT subscriptions completed!")
}

func messageHandler(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	log.Printf("Received message on topic %s: %s", topic, string(payload))

	switch topic {
	case "aquarium/feeder/status":
		handleFeederStatus(payload)
	case "aquarium/uv/status":
		handleUVStatus(payload)
	case "aquarium/device/report":
		handleDeviceReport(payload)
	case "aquarium/sensor/dht":
		handleSensorData(payload)
	}
}

func handleFeederStatus(payload []byte) {
	var status FeederStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		log.Printf("Error parsing feeder status: %v", err)
		return
	}

	// Update device status in database
	var deviceStatus models.DeviceStatus
	if err := database.DB.Where("device_type = ?", "FEEDER").First(&deviceStatus).Error; err != nil {
		log.Printf("Error finding feeder status: %v", err)
		return
	}

	deviceStatus.Status = status.Status
	deviceStatus.LastUpdated = time.Now()
	database.DB.Save(&deviceStatus)
}

func handleUVStatus(payload []byte) {
	var status UVStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		log.Printf("Error parsing UV status: %v", err)
		return
	}

	log.Printf("[UV Status] Received from ESP: state=%s, remaining=%d", status.State, status.Remaining)

	// Note: We don't update database here anymore to avoid conflict with scheduler
	// Backend scheduler is the source of truth for UV state
	// This is just for monitoring/logging purposes
}

func handleDeviceReport(payload []byte) {
	var report DeviceReport
	if err := json.Unmarshal(payload, &report); err != nil {
		log.Printf("Error parsing device report: %v", err)
		return
	}

	// Find the latest pending/running action for this device type
	var action models.ActionHistory
	query := database.DB.Where("device_type = ? AND status IN ?", report.Type, []string{"PENDING", "RUNNING"}).
		Order("created_at DESC").
		First(&action)

	if query.Error != nil {
		log.Printf("Error finding action history: %v", query.Error)
		return
	}

	// Update action history
	now := time.Now()
	action.EndTime = &now

	if report.Result == "SUCCESS" {
		action.Status = "SUCCESS"
		if report.Type == "FEED" {
			// Update stock
			var stock models.Stock
			if err := database.DB.First(&stock).Error; err == nil {
				stock.AmountGram -= report.FeedGram
				if stock.AmountGram < 0 {
					stock.AmountGram = 0
				}
				database.DB.Save(&stock)
			}
		}
	} else {
		action.Status = "FAILED"
	}

	database.DB.Save(&action)
	log.Printf("Updated action history: ID=%d, Status=%s", action.ID, action.Status)
}

func handleSensorData(payload []byte) {
	var data SensorData
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("Error parsing sensor data: %v", err)
		return
	}

	// Save sensor data to database
	sensorLog := models.SensorLog{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		RecordedAt:  time.Now(),
	}

	if err := database.DB.Create(&sensorLog).Error; err != nil {
		log.Printf("Error saving sensor data: %v", err)
		return
	}

	log.Printf("Saved sensor data: Temp=%.2f°C, Humidity=%.2f%%", data.Temperature, data.Humidity)
}

func PublishFeederCommand(dose int) error {
	if MockMode {
		log.Println("⚠️  MOCK MODE: Simulating feeder command")
		return MockPublishFeederCommand(dose)
	}

	command := FeederCommand{
		Action: "FEED",
		Dose:   dose,
	}

	payload, err := json.Marshal(command)
	if err != nil {
		log.Printf("❌ Error marshaling feeder command: %v", err)
		return err
	}

	log.Printf("📤 Publishing to aquarium/feeder/command: %s", string(payload))

	if Client == nil || !Client.IsConnected() {
		log.Println("❌ MQTT Client not connected!")
		return fmt.Errorf("MQTT client not connected")
	}

	token := Client.Publish("aquarium/feeder/command", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("❌ Failed to publish feeder command: %v", token.Error())
		return token.Error()
	}

	log.Printf("✅ Published feeder command: dose=%d", dose)
	return nil
}

func PublishUVCommand(state string, durationSec int) error {
	if MockMode {
		log.Println("⚠️  MOCK MODE: Simulating UV command")
		return MockPublishUVCommand(state, durationSec)
	}

	command := UVCommand{
		State:       state,
		DurationSec: durationSec,
	}

	payload, err := json.Marshal(command)
	if err != nil {
		log.Printf("❌ Error marshaling UV command: %v", err)
		return err
	}

	log.Printf("📤 Publishing to aquarium/uv/command: %s", string(payload))

	if Client == nil || !Client.IsConnected() {
		log.Println("❌ MQTT Client not connected!")
		return fmt.Errorf("MQTT client not connected")
	}

	token := Client.Publish("aquarium/uv/command", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("❌ Failed to publish UV command: %v", token.Error())
		return token.Error()
	}

	log.Printf("✅ Published UV command: state=%s, duration=%d", state, durationSec)
	return nil
}
