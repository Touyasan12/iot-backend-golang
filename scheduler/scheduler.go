package scheduler

import (
	"fmt"
	"log"
	"time"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"
	"iot-backend-cursor/mqtt"
	"iot-backend-cursor/utils"

	"github.com/robfig/cron/v3"
)

var Cron *cron.Cron

func InitScheduler() {
	Cron = cron.New(cron.WithSeconds())

	// Run every minute
	Cron.AddFunc("0 * * * * *", checkSchedules)

	// Check manual UV expiration every 10 seconds
	Cron.AddFunc("*/10 * * * * *", checkManualUVExpiration)

	Cron.Start()
	log.Println("Scheduler started")
}

func checkSchedules() {
	now := time.Now()
	currentDay := now.Weekday().String()[:3] // Mon, Tue, etc.
	currentTime := now.Format("15:04")
	currentHour := now.Hour()
	currentMinute := now.Minute()

	log.Printf("Checking schedules at %s %s", currentDay, currentTime)

	// Check feeder schedules
	checkFeederSchedules(currentDay, currentTime)

	// Check UV schedules
	checkUVSchedules(currentDay, currentHour, currentMinute)
}

func checkFeederSchedules(dayName, timeStr string) {
	var schedules []models.PakanSchedule
	if err := database.DB.Where("day_name = ? AND time = ? AND is_active = ?", dayName, timeStr, true).Find(&schedules).Error; err != nil {
		log.Printf("Error checking feeder schedules: %v", err)
		return
	}

	for _, schedule := range schedules {
		// Check if we already processed this schedule today (within the last minute)
		now := time.Now()
		oneMinuteAgo := now.Add(-1 * time.Minute)

		var existingAction models.ActionHistory
		err := database.DB.Where("device_type = ? AND trigger_source = ? AND status IN ?",
			"FEEDER", "SCHEDULE", []string{"PENDING", "RUNNING", "SUCCESS"}).
			Where("start_time >= ? AND start_time <= ?", oneMinuteAgo, now).
			First(&existingAction).Error

		if err == nil {
			log.Printf("Feeder schedule already processed at %s", timeStr)
			continue
		}

		// Create action history
		action := models.ActionHistory{
			DeviceType:    "FEEDER",
			TriggerSource: "SCHEDULE",
			StartTime:     time.Now(),
			Status:        "PENDING",
			Value:         schedule.AmountGram,
		}

		if err := database.DB.Create(&action).Error; err != nil {
			log.Printf("Error creating feeder action: %v", err)
			continue
		}

		doses := utils.CalculateFeedDoses(schedule.AmountGram)

		// Publish MQTT command
		if err := mqtt.PublishFeederCommand(doses); err != nil {
			log.Printf("Error publishing feeder command: %v", err)
			action.Status = "FAILED"
			database.DB.Save(&action)
			continue
		}

		action.Status = "RUNNING"
		database.DB.Save(&action)
		log.Printf("Triggered feeder schedule: Day=%s, Time=%s, Amount=%dg", schedule.DayName, schedule.Time, schedule.AmountGram)
	}
}

func checkUVSchedules(dayName string, currentHour, currentMinute int) {
	var schedules []models.UVSchedule
	if err := database.DB.Where("day_name = ? AND is_active = ?", dayName, true).Find(&schedules).Error; err != nil {
		log.Printf("Error checking UV schedules: %v", err)
		return
	}

	// Check if there's a running manual UV (override)
	var manualUV models.ActionHistory
	hasManualUV := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "MANUAL", "RUNNING").
		Order("start_time DESC").
		First(&manualUV).Error == nil

	if hasManualUV {
		// Check if manual UV has ended
		if manualUV.EndTime != nil && time.Now().After(*manualUV.EndTime) {
			// Manual UV ended, mark as success
			manualUV.Status = "SUCCESS"
			database.DB.Save(&manualUV)
			hasManualUV = false
		} else {
			log.Printf("Manual UV is active, skipping schedule check")
			return
		}
	}

	for _, schedule := range schedules {
		startHour, startMin, err := parseTime(schedule.StartTime)
		if err != nil {
			log.Printf("Error parsing start time: %v", err)
			continue
		}

		endHour, endMin, err := parseTime(schedule.EndTime)
		if err != nil {
			log.Printf("Error parsing end time: %v", err)
			continue
		}

		// Check if current time is within schedule range
		isWithinRange := false
		currentTimeMinutes := currentHour*60 + currentMinute
		startTimeMinutes := startHour*60 + startMin
		endTimeMinutes := endHour*60 + endMin

		if startTimeMinutes <= endTimeMinutes {
			// Normal case: start <= end (e.g., 18:00 - 19:00 same day)
			isWithinRange = currentTimeMinutes >= startTimeMinutes && currentTimeMinutes < endTimeMinutes
		} else {
			// Overnight case: start > end (e.g., 20:00 - 04:00 crosses midnight)
			isWithinRange = currentTimeMinutes >= startTimeMinutes || currentTimeMinutes < endTimeMinutes
		}

		if isWithinRange {
			// Check if there is already a running schedule action
			var runningSchedule models.ActionHistory
			if err := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "SCHEDULE", "RUNNING").
				Order("start_time DESC").
				First(&runningSchedule).Error; err == nil {
				if runningSchedule.EndTime != nil && time.Now().Before(*runningSchedule.EndTime) {
					log.Printf("UV schedule already running (action ID %d)", runningSchedule.ID)
					continue
				}
			}

			// Calculate end time and duration FIRST (before DB transaction)
			var endTime time.Time
			var durationMinutes int

			if startTimeMinutes <= endTimeMinutes {
				// Normal same-day schedule (e.g., 18:00 - 19:00)
				if currentTimeMinutes >= startTimeMinutes && currentTimeMinutes < endTimeMinutes {
					durationMinutes = endTimeMinutes - currentTimeMinutes
				}
			} else {
				// Overnight schedule (e.g., 20:00 - 04:00)
				if currentTimeMinutes >= startTimeMinutes {
					// Before midnight (e.g., 23:00)
					minutesUntilMidnight := (24*60 - currentTimeMinutes)
					durationMinutes = minutesUntilMidnight + endTimeMinutes
				} else if currentTimeMinutes < endTimeMinutes {
					// After midnight (e.g., 02:00)
					durationMinutes = endTimeMinutes - currentTimeMinutes
				}
			}

			// Safety check: only proceed if duration is valid
			if durationMinutes <= 0 {
				log.Printf("Invalid duration calculated: %d minutes, skipping", durationMinutes)
				continue
			}

			endTime = time.Now().Add(time.Duration(durationMinutes) * time.Minute)

			// Step 1: Publish MQTT FIRST (outside DB transaction to avoid locking)
			if err := mqtt.PublishUVCommand("ON", 0); err != nil {
				log.Printf("Error publishing UV command: %v", err)
				continue
			}

			// Step 2: FAST DB transaction (only insert/update, no external calls)
			action := models.ActionHistory{
				DeviceType:    "UV",
				TriggerSource: "SCHEDULE",
				StartTime:     time.Now(),
				EndTime:       &endTime,
				Status:        "RUNNING",
				Value:         durationMinutes * 60, // Convert to seconds
			}

			if err := database.DB.Create(&action).Error; err != nil {
				log.Printf("Error creating UV action: %v", err)
				continue
			}

			// Update device status (fast, no external dependency)
			database.DB.Model(&models.DeviceStatus{}).Where("device_type = ?", "UV").Updates(map[string]interface{}{
				"status":       "ON",
				"remaining":    0,
				"last_updated": time.Now(),
			})

			log.Printf("Triggered UV schedule: Day=%s, Start=%s, End=%s, Duration=%dm", schedule.DayName, schedule.StartTime, schedule.EndTime, durationMinutes)
		} else {
			// Outside schedule range, ensure UV is turned off if a schedule action is running
			var runningSchedule models.ActionHistory
			if database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "SCHEDULE", "RUNNING").
				Order("start_time DESC").
				First(&runningSchedule).Error == nil {

				// Step 1: Send MQTT command FIRST (outside DB transaction)
				if err := mqtt.PublishUVCommand("OFF", 0); err != nil {
					log.Printf("Error turning OFF UV: %v", err)
				} else {
					// Step 2: FAST DB update (only after MQTT succeeds)
					runningSchedule.Status = "SUCCESS"
					now := time.Now()
					runningSchedule.EndTime = &now
					database.DB.Save(&runningSchedule)

					database.DB.Model(&models.DeviceStatus{}).Where("device_type = ?", "UV").Updates(map[string]interface{}{
						"status":       "OFF",
						"remaining":    0,
						"last_updated": now,
					})
					log.Printf("Turned OFF UV (outside schedule)")
				}
			}
		}
	}
}

func checkManualUVExpiration() {
	// Find running manual UV actions
	var manualUV models.ActionHistory
	err := database.DB.Where("device_type = ? AND trigger_source = ? AND status = ?", "UV", "MANUAL", "RUNNING").
		Order("start_time DESC").
		First(&manualUV).Error

	if err != nil {
		// No running manual UV found
		return
	}

	// Check if manual UV has ended
	if manualUV.EndTime != nil && time.Now().After(*manualUV.EndTime) {
		log.Printf("Manual UV expired (Action ID: %d), sending OFF command", manualUV.ID)

		// Send OFF command to ESP
		if err := mqtt.PublishUVCommand("OFF", 0); err != nil {
			log.Printf("Error sending UV OFF command: %v", err)
			return
		}

		// Update action history
		now := time.Now()
		manualUV.Status = "SUCCESS"
		manualUV.EndTime = &now
		database.DB.Save(&manualUV)

		// Update device status
		database.DB.Model(&models.DeviceStatus{}).Where("device_type = ?", "UV").Updates(map[string]interface{}{
			"status":       "OFF",
			"remaining":    0,
			"last_updated": now,
		})

		log.Printf("Manual UV turned OFF successfully")
	}
}

func parseTime(timeStr string) (hour, minute int, err error) {
	_, err = fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	return
}
