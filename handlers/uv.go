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

// GetUVSchedules returns all UV schedules with pagination
func GetUVSchedules(c *gin.Context) {
	var schedules []models.UVSchedule
	var total int64

	// Get pagination params (default: page 1, page_size 20, max 100)
	pagination := utils.GetPaginationParams(c, 20, 100)

	// Count total records
	if err := database.DB.Model(&models.UVSchedule{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated results ordered by day and start_time
	if err := database.DB.Order("day_name, start_time").
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

// CreateUVSchedule creates a new UV schedule
func CreateUVSchedule(c *gin.Context) {
	var schedule models.UVSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// UpdateUVSchedule updates a UV schedule
func UpdateUVSchedule(c *gin.Context) {
	id := c.Param("id")
	var schedule models.UVSchedule

	if err := database.DB.First(&schedule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Save(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DeleteUVSchedule deletes a UV schedule
func DeleteUVSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.UVSchedule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted"})
}

// ManualUVRequest represents the request body for manual UV control
type ManualUVRequest struct {
	DurationMinutes int `json:"duration_minutes" binding:"required"`
}

// ManualUV triggers manual UV control
func ManualUV(c *gin.Context) {
	var req ManualUVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration_minutes is required"})
		return
	}

	if req.DurationMinutes <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration_minutes must be greater than 0"})
		return
	}

	durationSec := req.DurationMinutes * 60
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationSec) * time.Second)

	// Create action history entry
	action := models.ActionHistory{
		DeviceType:    "UV",
		TriggerSource: "MANUAL",
		StartTime:     startTime,
		EndTime:       &endTime,
		Status:        "RUNNING",
		Value:         durationSec,
	}

	if err := database.DB.Create(&action).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish MQTT command
	if err := mqtt.PublishUVCommand("ON", durationSec); err != nil {
		action.Status = "FAILED"
		database.DB.Save(&action)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send command to device"})
		return
	}

	// Update device status immediately
	var deviceStatus models.DeviceStatus
	if err := database.DB.Where("device_type = ?", "UV").First(&deviceStatus).Error; err == nil {
		deviceStatus.Status = "ON"
		deviceStatus.Remaining = 0 // Backend doesn't track remaining, scheduler will turn off
		deviceStatus.LastUpdated = time.Now()
		database.DB.Save(&deviceStatus)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "UV command sent",
		"action_id":    action.ID,
		"duration_sec": durationSec,
		"end_time":     endTime.Format(time.RFC3339),
	})
}

// StopManualUV stops any currently running UV (manual or schedule)
func StopManualUV(c *gin.Context) {
	// Find any running UV action (manual or schedule)
	var action models.ActionHistory
	err := database.DB.Where("device_type = ? AND status = ?", "UV", "RUNNING").
		Order("start_time DESC").
		First(&action).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No running UV found"})
		return
	}

	// Publish MQTT command to turn off UV
	if err := mqtt.PublishUVCommand("OFF", 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send stop command to device"})
		return
	}

	now := time.Now()
	action.EndTime = &now
	action.Status = "STOPPED"
	database.DB.Save(&action)

	// Update device status
	var deviceStatus models.DeviceStatus
	if err := database.DB.Where("device_type = ?", "UV").First(&deviceStatus).Error; err == nil {
		deviceStatus.Status = "OFF"
		deviceStatus.Remaining = 0
		deviceStatus.LastUpdated = now
		database.DB.Save(&deviceStatus)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "UV stopped successfully",
		"action_id":      action.ID,
		"trigger_source": action.TriggerSource,
		"stopped_at":     now.Format(time.RFC3339),
	})
}

// GetUVStatus returns current UV status
func GetUVStatus(c *gin.Context) {
	var status models.DeviceStatus
	if err := database.DB.Where("device_type = ?", "UV").First(&status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UV status not found"})
		return
	}

	// Check if there's a running manual UV
	var manualUV models.ActionHistory
	hasManual := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "MANUAL", "RUNNING").
		Order("start_time DESC").
		First(&manualUV).Error == nil

	response := gin.H{
		"state":         status.Status,
		"remaining":     status.Remaining,
		"last_updated":  status.LastUpdated,
		"manual_active": hasManual,
	}

	if hasManual {
		response["manual_end_time"] = manualUV.EndTime
	}

	c.JSON(http.StatusOK, response)
}
