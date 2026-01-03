package database

import (
	"log"
)

// CreateIndexes creates database indexes for better performance
func CreateIndexes() {
	// Index for action_histories lookup (scheduler checks this frequently)
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_action_histories_lookup 
		ON action_histories(device_type, trigger_source, status, start_time DESC)
	`).Error; err != nil {
		log.Printf("Warning: Could not create index idx_action_histories_lookup: %v", err)
	} else {
		log.Println("Index created: idx_action_histories_lookup")
	}

	// Index for action_histories time-based queries
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_action_histories_time 
		ON action_histories(created_at DESC)
	`).Error; err != nil {
		log.Printf("Warning: Could not create index idx_action_histories_time: %v", err)
	} else {
		log.Println("Index created: idx_action_histories_time")
	}

	// Index for sensor_logs time-based queries
	if err := DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sensor_logs_time 
		ON sensor_logs(recorded_at DESC)
	`).Error; err != nil {
		log.Printf("Warning: Could not create index idx_sensor_logs_time: %v", err)
	} else {
		log.Println("Index created: idx_sensor_logs_time")
	}
}
