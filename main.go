package main

import (
	"log"
	"time"

	"iot-backend-cursor/config"
	"iot-backend-cursor/database"
	"iot-backend-cursor/mqtt"
	"iot-backend-cursor/routes"
	"iot-backend-cursor/scheduler"
)

// init sets the application timezone to Asia/Jakarta (WIB)
func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Warning: Failed to load Asia/Jakarta timezone: %v, using UTC", err)
		return
	}
	time.Local = loc
	log.Println("✅ Timezone set to Asia/Jakarta (WIB)")
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	database.InitDB(cfg)

	// Initialize MQTT client (or mock if demo mode)
	if cfg.DemoMode {
		mqtt.InitMockMQTT()
		log.Println("Running in DEMO MODE - No MQTT broker required")
	} else {
		mqtt.InitMQTT(cfg)
	}

	// Initialize scheduler
	scheduler.InitScheduler()

	// Setup routes
	r := routes.SetupRoutes()

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
