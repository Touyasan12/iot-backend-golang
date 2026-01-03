package handlers

import (
	"net/http"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"

	"github.com/gin-gonic/gin"
)

// GetDashboard returns dashboard data
func GetDashboard(c *gin.Context) {
	// Get stock
	var stock models.Stock
	database.DB.First(&stock)

	// Get UV status
	var uvStatus models.DeviceStatus
	database.DB.Where("device_type = ?", "UV").First(&uvStatus)

	// Get feeder status
	var feederStatus models.DeviceStatus
	database.DB.Where("device_type = ?", "FEEDER").First(&feederStatus)

	// Check for running manual UV
	var manualUV models.ActionHistory
	hasManualUV := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "MANUAL", "RUNNING").
		Order("start_time DESC").
		First(&manualUV).Error == nil

	// Get latest sensor reading
	var sensor models.SensorLog
	var environment *gin.H
	if err := database.DB.Order("recorded_at DESC").First(&sensor).Error; err == nil {
		environment = &gin.H{
			"temperature":  sensor.Temperature,
			"humidity":     sensor.Humidity,
			"last_updated": sensor.RecordedAt,
		}
	}

	response := gin.H{
		"stock": gin.H{
			"amount_gram": stock.AmountGram,
		},
		"uv": gin.H{
			"state":         uvStatus.Status,
			"remaining":     uvStatus.Remaining,
			"last_updated":  uvStatus.LastUpdated,
			"manual_active": hasManualUV,
		},
		"feeder": gin.H{
			"status":       feederStatus.Status,
			"last_updated": feederStatus.LastUpdated,
		},
		"environment": environment,
	}

	c.JSON(http.StatusOK, response)
}
