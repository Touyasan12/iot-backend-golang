package handlers

import (
	"net/http"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"
	"iot-backend-cursor/utils"

	"github.com/gin-gonic/gin"
)

// GetHistory returns action history with optional filters and pagination
func GetHistory(c *gin.Context) {
	var history []models.ActionHistory
	var total int64

	// Get pagination params (default: page 1, page_size 50, max 200)
	pagination := utils.GetPaginationParams(c, 50, 200)

	// Build query with filters
	query := database.DB.Model(&models.ActionHistory{})

	// Filter by device type
	if deviceType := c.Query("device_type"); deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}

	// Filter by trigger source
	if triggerSource := c.Query("trigger_source"); triggerSource != "" {
		query = query.Where("trigger_source = ?", triggerSource)
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get paginated results
	if err := query.Order("created_at DESC").
		Offset(pagination.Offset).
		Limit(pagination.PageSize).
		Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response with pagination metadata
	paginationMeta := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, total)

	c.JSON(http.StatusOK, gin.H{
		"data":       history,
		"pagination": paginationMeta,
	})
}
