package handlers

import (
	"net/http"
	"time"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"
	"iot-backend-cursor/mqtt"
	"iot-backend-cursor/utils"

	"github.com/gin-gonic/gin"
)

// GetFeederSchedules returns all feeding schedules with pagination
func GetFeederSchedules(c *gin.Context) {
	var schedules []models.PakanSchedule
	var total int64

	// Get pagination params (default: page 1, page_size 20, max 100)
	pagination := utils.GetPaginationParams(c, 20, 100)

	// Count total records
	if err := database.DB.Model(&models.PakanSchedule{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated results ordered by day and time
	if err := database.DB.Order("day_name, time").
		Offset(pagination.Offset).
		Limit(pagination.PageSize).
		Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response with pagination metadata
	paginationMeta := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, total)

	c.JSON(http.StatusOK, gin.H{
		"data":       schedules,
		"pagination": paginationMeta,
	})
}

// CreateFeederSchedule creates a new feeding schedule
func CreateFeederSchedule(c *gin.Context) {
	var schedule models.PakanSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate: max 5 schedules per day
	var count int64
	database.DB.Model(&models.PakanSchedule{}).
		Where("day_name = ? AND is_active = ?", schedule.DayName, true).
		Count(&count)

	if count >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 5 feeding schedules per day"})
		return
	}

	// Set default amount if not provided
	if schedule.AmountGram == 0 {
		schedule.AmountGram = 10
	}

	if err := database.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// UpdateFeederSchedule updates a feeding schedule
func UpdateFeederSchedule(c *gin.Context) {
	id := c.Param("id")
	var schedule models.PakanSchedule

	if err := database.DB.First(&schedule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate: max 5 schedules per day (excluding current one)
	var count int64
	database.DB.Model(&models.PakanSchedule{}).
		Where("day_name = ? AND is_active = ? AND id != ?", schedule.DayName, true, id).
		Count(&count)

	if count >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 5 feeding schedules per day"})
		return
	}

	if err := database.DB.Save(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DeleteFeederSchedule deletes a feeding schedule
func DeleteFeederSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.PakanSchedule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted"})
}

type ManualFeedRequest struct {
	AmountGram int `json:"amount_gram"`
}

// ManualFeed triggers manual feeding
func ManualFeed(c *gin.Context) {
	var req ManualFeedRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	amountGram := utils.NormalizeFeedAmount(req.AmountGram)
	doses := utils.CalculateFeedDoses(amountGram)

	// Get last successful feed
	var lastFeed models.ActionHistory
	err := database.DB.Where("device_type = ? AND status = ?", "FEEDER", "SUCCESS").
		Order("start_time DESC").
		First(&lastFeed).Error

	// Create action history entry
	action := models.ActionHistory{
		DeviceType:    "FEEDER",
		TriggerSource: "MANUAL",
		StartTime:     time.Now(),
		Status:        "PENDING",
		Value:         amountGram,
	}

	if err := database.DB.Create(&action).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish MQTT command
	if err := mqtt.PublishFeederCommand(doses); err != nil {
		action.Status = "FAILED"
		database.DB.Save(&action)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send command to device"})
		return
	}

	action.Status = "RUNNING"
	database.DB.Save(&action)

	// Prepare response with last feed info
	response := gin.H{
		"message":     "Feeding command sent",
		"action_id":   action.ID,
		"amount_gram": amountGram,
	}

	if err == nil {
		response["last_feed"] = gin.H{
			"day":  lastFeed.StartTime.Format("Monday"),
			"time": lastFeed.StartTime.Format("15:04"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetLastFeedInfo returns information about the last successful feed
func GetLastFeedInfo(c *gin.Context) {
	var lastFeed models.ActionHistory
	err := database.DB.Where("device_type = ? AND status = ?", "FEEDER", "SUCCESS").
		Order("start_time DESC").
		First(&lastFeed).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"exists":  false,
			"message": "No previous feed found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exists": true,
		"day":    lastFeed.StartTime.Format("Monday"),
		"time":   lastFeed.StartTime.Format("15:04"),
		"date":   lastFeed.StartTime.Format("2006-01-02"),
	})
}