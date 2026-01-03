package handlers

import (
	"net/http"
	"time"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"
	"iot-backend-cursor/utils"

	"github.com/gin-gonic/gin"
)

// GetCurrentSensor returns the latest sensor reading
func GetCurrentSensor(c *gin.Context) {
	var sensor models.SensorLog
	err := database.DB.Order("recorded_at DESC").First(&sensor).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"exists":  false,
			"message": "No sensor data available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"temperature":  sensor.Temperature,
		"humidity":     sensor.Humidity,
		"last_updated": sensor.RecordedAt,
	})
}

// GetSensorHistory returns sensor readings history with pagination and time filter
func GetSensorHistory(c *gin.Context) {
	var sensors []models.SensorLog
	var total int64

	// Get pagination params (default: page 1, page_size 100, max 500)
	pagination := utils.GetPaginationParams(c, 100, 500)

	// Time period filter
	period := c.Query("period")
	query := database.DB.Model(&models.SensorLog{})

	// Apply time filter
	now := time.Now()
	switch period {
	case "24h":
		query = query.Where("recorded_at >= ?", now.Add(-24*time.Hour))
	case "7d":
		query = query.Where("recorded_at >= ?", now.Add(-7*24*time.Hour))
	case "30d":
		query = query.Where("recorded_at >= ?", now.Add(-30*24*time.Hour))
	default:
		// No filter, return all (with pagination)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated results
	if err := query.Order("recorded_at DESC").
		Offset(pagination.Offset).
		Limit(pagination.PageSize).
		Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response with pagination metadata
	paginationMeta := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, total)

	c.JSON(http.StatusOK, gin.H{
		"data":       sensors,
		"pagination": paginationMeta,
	})
}

// InjectSensorData manually injects sensor data (for testing/demo purposes)
func InjectSensorData(c *gin.Context) {
	var input struct {
		Temperature float64 `json:"temperature" binding:"required"`
		Humidity    float64 `json:"humidity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate ranges
	if input.Temperature < -40 || input.Temperature > 80 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Temperature must be between -40 and 80 Celsius"})
		return
	}
	if input.Humidity < 0 || input.Humidity > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Humidity must be between 0 and 100 percent"})
		return
	}

	// Create sensor log
	sensor := models.SensorLog{
		Temperature: input.Temperature,
		Humidity:    input.Humidity,
		RecordedAt:  time.Now(),
	}

	if err := database.DB.Create(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Sensor data injected successfully",
		"id":          sensor.ID,
		"temperature": sensor.Temperature,
		"humidity":    sensor.Humidity,
		"recorded_at": sensor.RecordedAt,
	})
}
