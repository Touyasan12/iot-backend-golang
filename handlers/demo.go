package handlers

import (
	"net/http"
	"time"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"

	"github.com/gin-gonic/gin"
)

// SeedDemoData populates database with demo data
func SeedDemoData(c *gin.Context) {
	// Clear existing data (optional - comment out if you want to keep existing data)
	// database.DB.Exec("DELETE FROM pakan_schedules")
	// database.DB.Exec("DELETE FROM uv_schedules")
	// database.DB.Exec("DELETE FROM action_history")

	// Set initial stock
	var stock models.Stock
	if err := database.DB.First(&stock).Error; err != nil {
		stock = models.Stock{AmountGram: 1000} // 1kg stock
		database.DB.Create(&stock)
	} else {
		stock.AmountGram = 1000
		database.DB.Save(&stock)
	}

	// Create sample feeder schedules
	feederSchedules := []models.PakanSchedule{
		{DayName: "Mon", Time: "08:00", AmountGram: 10, IsActive: true},
		{DayName: "Mon", Time: "12:00", AmountGram: 10, IsActive: true},
		{DayName: "Mon", Time: "18:00", AmountGram: 10, IsActive: true},
		{DayName: "Tue", Time: "08:00", AmountGram: 10, IsActive: true},
		{DayName: "Tue", Time: "12:00", AmountGram: 10, IsActive: true},
		{DayName: "Wed", Time: "08:00", AmountGram: 10, IsActive: true},
		{DayName: "Wed", Time: "18:00", AmountGram: 10, IsActive: true},
		{DayName: "Thu", Time: "08:00", AmountGram: 10, IsActive: true},
		{DayName: "Fri", Time: "08:00", AmountGram: 10, IsActive: true},
		{DayName: "Sat", Time: "09:00", AmountGram: 10, IsActive: true},
		{DayName: "Sun", Time: "09:00", AmountGram: 10, IsActive: true},
	}

	for _, schedule := range feederSchedules {
		var existing models.PakanSchedule
		if err := database.DB.Where("day_name = ? AND time = ?", schedule.DayName, schedule.Time).First(&existing).Error; err != nil {
			database.DB.Create(&schedule)
		}
	}

	// Create sample UV schedules
	uvSchedules := []models.UVSchedule{
		{DayName: "Mon", StartTime: "20:00", EndTime: "04:00", IsActive: true},
		{DayName: "Tue", StartTime: "20:00", EndTime: "04:00", IsActive: true},
		{DayName: "Wed", StartTime: "20:00", EndTime: "04:00", IsActive: true},
		{DayName: "Thu", StartTime: "20:00", EndTime: "04:00", IsActive: true},
		{DayName: "Fri", StartTime: "20:00", EndTime: "04:00", IsActive: true},
		{DayName: "Sat", StartTime: "20:00", EndTime: "04:00", IsActive: true},
		{DayName: "Sun", StartTime: "20:00", EndTime: "04:00", IsActive: true},
	}

	for _, schedule := range uvSchedules {
		var existing models.UVSchedule
		if err := database.DB.Where("day_name = ? AND start_time = ? AND end_time = ?", schedule.DayName, schedule.StartTime, schedule.EndTime).First(&existing).Error; err != nil {
			database.DB.Create(&schedule)
		}
	}

	// Create sample history (last 7 days)
	now := time.Now()
	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i)
		
		// Feeder history
		for j := 0; j < 3; j++ {
			feedTime := time.Date(date.Year(), date.Month(), date.Day(), 8+j*4, 0, 0, 0, time.Local)
			if feedTime.Before(now) {
				endTime := feedTime.Add(5 * time.Second)
				action := models.ActionHistory{
					DeviceType:    "FEEDER",
					TriggerSource: "SCHEDULE",
					StartTime:     feedTime,
					EndTime:       &endTime,
					Status:        "SUCCESS",
					Value:         10,
				}
				database.DB.Create(&action)
			}
		}

		// UV history
		uvStart := time.Date(date.Year(), date.Month(), date.Day(), 20, 0, 0, 0, time.Local)
		uvEnd := time.Date(date.Year(), date.Month(), date.Day(), 4, 0, 0, 0, time.Local).AddDate(0, 0, 1)
		if uvStart.Before(now) {
			if uvEnd.After(now) {
				uvEnd = now
			}
			action := models.ActionHistory{
				DeviceType:    "UV",
				TriggerSource: "SCHEDULE",
				StartTime:     uvStart,
				EndTime:       &uvEnd,
				Status:        "SUCCESS",
				Value:         0,
			}
			database.DB.Create(&action)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Demo data seeded successfully",
		"stock":   stock.AmountGram,
		"schedules": gin.H{
			"feeder": len(feederSchedules),
			"uv":     len(uvSchedules),
		},
	})
}

// ClearDemoData clears all demo data
func ClearDemoData(c *gin.Context) {
	database.DB.Exec("DELETE FROM action_history")
	database.DB.Exec("DELETE FROM pakan_schedules")
	database.DB.Exec("DELETE FROM uv_schedules")
	
	var stock models.Stock
	if err := database.DB.First(&stock).Error; err == nil {
		stock.AmountGram = 0
		database.DB.Save(&stock)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Demo data cleared"})
}


