package database

import (
	"log"

	"iot-backend-cursor/config"
	"iot-backend-cursor/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	var err error
	var dialector gorm.Dialector

	if cfg.DBType == "sqlite" {
		dialector = sqlite.Open(cfg.GetDSN())
	} else {
		dialector = postgres.Open(cfg.GetDSN())
	}

	// Set logger based on DB type (less verbose for remote DB)
	logLevel := logger.Info
	if cfg.DBType == "postgres" {
		logLevel = logger.Warn // Only log warnings/errors for remote DB
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	// Auto migrate
	err = DB.AutoMigrate(
		&models.PakanSchedule{},
		&models.UVSchedule{},
		&models.ActionHistory{},
		&models.Stock{},
		&models.DeviceStatus{},
		&models.SensorLog{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migrated successfully")

	// Create indexes for better query performance
	CreateIndexes()

	// Initialize stock if not exists
	var stock models.Stock
	if err := DB.First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			DB.Create(&models.Stock{AmountGram: 0})
			log.Println("Initialized stock with 0 grams")
		}
	}

	// Initialize device statuses
	devices := []string{"FEEDER", "UV"}
	for _, deviceType := range devices {
		var status models.DeviceStatus
		if err := DB.Where("device_type = ?", deviceType).First(&status).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				DB.Create(&models.DeviceStatus{
					DeviceType:  deviceType,
					Status:      "IDLE",
					Remaining:   0,
					LastUpdated: models.DeviceStatus{}.LastUpdated,
				})
			}
		}
	}
}
