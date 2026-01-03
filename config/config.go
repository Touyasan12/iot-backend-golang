package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBType       string // "postgres" or "sqlite"
	MQTTBroker   string
	MQTTPort     string
	MQTTUser     string
	MQTTPass     string
	MQTTClientID string
	DemoMode     bool // Enable demo mode (no real MQTT connection)
}

func LoadConfig() *Config {
	// Load .env file if exists
	godotenv.Load()

	demoMode := getEnv("DEMO_MODE", "false") == "true"

	config := &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "aquarium_db"),
		DBType:       getEnv("DB_TYPE", "sqlite"), // Default to sqlite for demo
		MQTTBroker:   getEnv("MQTT_BROKER", "localhost"),
		MQTTPort:     getEnv("MQTT_PORT", "1883"),
		MQTTUser:     getEnv("MQTT_USER", ""),
		MQTTPass:     getEnv("MQTT_PASS", ""),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "aquarium-backend"),
		DemoMode:     demoMode,
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Config) GetDSN() string {
	if c.DBType == "sqlite" {
		return c.DBName + ".db"
	}
	return "host=" + c.DBHost + " user=" + c.DBUser + " password=" + c.DBPassword + " dbname=" + c.DBName + " port=" + c.DBPort + " sslmode=disable"
}

func (c *Config) GetMQTTBrokerURL() string {
	// Use TLS for port 8883, TCP for others
	if c.MQTTPort == "8883" {
		return "tls://" + c.MQTTBroker + ":" + c.MQTTPort
	}
	return "tcp://" + c.MQTTBroker + ":" + c.MQTTPort
}
