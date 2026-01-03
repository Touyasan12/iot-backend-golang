package mqtt

import (
	"log"
	"time"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"
)

var MockMode bool = false

// InitMockMQTT initializes mock MQTT client for demo mode
func InitMockMQTT() {
	MockMode = true
	log.Println("MQTT Mock Mode: Enabled - Simulating device responses")

	// Initialize device statuses
	var feederStatus models.DeviceStatus
	if err := database.DB.Where("device_type = ?", "FEEDER").First(&feederStatus).Error; err != nil {
		feederStatus = models.DeviceStatus{
			DeviceType:  "FEEDER",
			Status:      "IDLE",
			Remaining:   0,
			LastUpdated: time.Now(),
		}
		database.DB.Create(&feederStatus)
	}

	var uvStatus models.DeviceStatus
	if err := database.DB.Where("device_type = ?", "UV").First(&uvStatus).Error; err != nil {
		uvStatus = models.DeviceStatus{
			DeviceType:  "UV",
			Status:      "OFF",
			Remaining:   0,
			LastUpdated: time.Now(),
		}
		database.DB.Create(&uvStatus)
	}

	// Start mock sensor data generator
	go mockSensorDataGenerator()
}

// mockSensorDataGenerator simulates DHT sensor readings every 5 minutes
func mockSensorDataGenerator() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Send initial data immediately
	sendMockSensorData()

	for range ticker.C {
		sendMockSensorData()
	}
}

// sendMockSensorData generates and saves mock sensor readings
func sendMockSensorData() {
	// Generate realistic aquarium temperature (25-30°C) and humidity (60-80%)
	temperature := 25.0 + float64(time.Now().Unix()%5) + float64(time.Now().Nanosecond()%100)/100.0
	humidity := 60.0 + float64(time.Now().Unix()%20) + float64(time.Now().Nanosecond()%100)/100.0

	sensorLog := models.SensorLog{
		Temperature: temperature,
		Humidity:    humidity,
		RecordedAt:  time.Now(),
	}

	if err := database.DB.Create(&sensorLog).Error; err == nil {
		log.Printf("[MOCK] Sensor data: Temp=%.2f°C, Humidity=%.2f%%", temperature, humidity)
	}
}

// MockPublishFeederCommand simulates publishing feeder command
func MockPublishFeederCommand(dose int) error {
	log.Printf("[MOCK] Published feeder command: dose=%d", dose)

	// Simulate device processing (instant - no delay)
	go func() {
		// Update feeder status to DISPENSING
		var deviceStatus models.DeviceStatus
		if err := database.DB.Where("device_type = ?", "FEEDER").First(&deviceStatus).Error; err == nil {
			deviceStatus.Status = "DISPENSING"
			deviceStatus.LastUpdated = time.Now()
			database.DB.Save(&deviceStatus)
		}

		// Update feeder status back to IDLE (instant)
		if err := database.DB.Where("device_type = ?", "FEEDER").First(&deviceStatus).Error; err == nil {
			deviceStatus.Status = "IDLE"
			deviceStatus.LastUpdated = time.Now()
			database.DB.Save(&deviceStatus)
		}

		// Find the latest pending/running action
		var action models.ActionHistory
		if err := database.DB.Where("device_type = ? AND status IN ?", "FEEDER", []string{"PENDING", "RUNNING"}).
			Order("created_at DESC").
			First(&action).Error; err == nil {
			// Simulate successful feed
			now := time.Now()
			action.EndTime = &now
			action.Status = "SUCCESS"

			if err := database.DB.Save(&action).Error; err == nil {
				// Update stock
				var stock models.Stock
				if err := database.DB.First(&stock).Error; err == nil {
					stock.AmountGram -= action.Value
					if stock.AmountGram < 0 {
						stock.AmountGram = 0
					}
					database.DB.Save(&stock)
					log.Printf("[MOCK] Feed completed successfully: %dg dispensed, stock: %dg", action.Value, stock.AmountGram)
				}
			}
		}
	}()

	return nil
}

// MockPublishUVCommand simulates publishing UV command
func MockPublishUVCommand(state string, durationSec int) error {
	log.Printf("[MOCK] Published UV command: state=%s, duration=%d", state, durationSec)

	if state == "ON" {
		// Update UV status
		var deviceStatus models.DeviceStatus
		if err := database.DB.Where("device_type = ?", "UV").First(&deviceStatus).Error; err == nil {
			deviceStatus.Status = "ON"
			deviceStatus.Remaining = durationSec
			deviceStatus.LastUpdated = time.Now()
			database.DB.Save(&deviceStatus)
		}

		// If duration is 0 (schedule mode), we'll let the scheduler handle it
		if durationSec > 0 {
			// For manual UV, simulate countdown
			go func() {
				remaining := durationSec
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for remaining > 0 {
					<-ticker.C
					remaining--

					var deviceStatus models.DeviceStatus
					if err := database.DB.Where("device_type = ?", "UV").First(&deviceStatus).Error; err == nil {
						deviceStatus.Remaining = remaining
						deviceStatus.LastUpdated = time.Now()
						database.DB.Save(&deviceStatus)
					}

					// Check if UV was turned off manually
					if deviceStatus.Status == "OFF" {
						break
					}
				}

				// Turn off UV when duration ends
				if err := database.DB.Where("device_type = ?", "UV").First(&deviceStatus).Error; err == nil {
					deviceStatus.Status = "OFF"
					deviceStatus.Remaining = 0
					deviceStatus.LastUpdated = time.Now()
					database.DB.Save(&deviceStatus)
				}

				// Update action history
				var action models.ActionHistory
				if err := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "MANUAL", "RUNNING").
					Order("start_time DESC").
					First(&action).Error; err == nil {
					now := time.Now()
					action.EndTime = &now
					action.Status = "SUCCESS"
					database.DB.Save(&action)
					log.Printf("[MOCK] UV turned off after %d seconds", durationSec)
				}
			}()
		} else {
			// Schedule mode - update action history
			var action models.ActionHistory
			if err := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "SCHEDULE", "RUNNING").
				Order("start_time DESC").
				First(&action).Error; err == nil {
				log.Printf("[MOCK] UV turned ON (schedule mode)")
			}
		}
	} else if state == "OFF" {
		// Turn off UV
		var deviceStatus models.DeviceStatus
		if err := database.DB.Where("device_type = ?", "UV").First(&deviceStatus).Error; err == nil {
			deviceStatus.Status = "OFF"
			deviceStatus.Remaining = 0
			deviceStatus.LastUpdated = time.Now()
			database.DB.Save(&deviceStatus)
		}

		// Update any running schedule action
		var action models.ActionHistory
		if err := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "SCHEDULE", "RUNNING").
			Order("start_time DESC").
			First(&action).Error; err == nil {
			now := time.Now()
			action.EndTime = &now
			action.Status = "SUCCESS"
			database.DB.Save(&action)
		}

		log.Printf("[MOCK] UV turned OFF")
	}

	return nil
}
