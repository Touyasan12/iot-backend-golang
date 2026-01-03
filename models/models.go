package models

import (
	"time"

	"gorm.io/gorm"
)

// PakanSchedule represents the feeding schedule
type PakanSchedule struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DayName    string    `json:"day_name" gorm:"not null"` // Mon, Tue, Wed, Thu, Fri, Sat, Sun
	Time       string    `json:"time" gorm:"not null"`     // HH:MM format
	AmountGram int       `json:"amount_gram" gorm:"default:10"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UVSchedule represents the UV sterilizer schedule
type UVSchedule struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DayName   string    `json:"day_name" gorm:"not null"`   // Mon, Tue, Wed, Thu, Fri, Sat, Sun
	StartTime string    `json:"start_time" gorm:"not null"` // HH:MM format
	EndTime   string    `json:"end_time" gorm:"not null"`   // HH:MM format
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ActionHistory represents the log of all actions
type ActionHistory struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	DeviceType    string         `json:"device_type" gorm:"not null"`    // FEEDER, UV
	TriggerSource string         `json:"trigger_source" gorm:"not null"` // SCHEDULE, MANUAL
	StartTime     time.Time      `json:"start_time" gorm:"not null"`
	EndTime       *time.Time     `json:"end_time"`                               // nullable
	Status        string         `json:"status" gorm:"not null;default:PENDING"` // PENDING, RUNNING, SUCCESS, FAILED, OVERRIDDEN
	Value         int            `json:"value"`                                  // grams for feeder, seconds for UV
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Stock represents the food stock
type Stock struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	AmountGram int       `json:"amount_gram" gorm:"default:0"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DeviceStatus represents current device status
type DeviceStatus struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	DeviceType  string    `json:"device_type" gorm:"uniqueIndex;not null"` // FEEDER, UV
	Status      string    `json:"status"`                                  // IDLE, DISPENSING, ON, OFF
	Remaining   int       `json:"remaining"`                               // remaining seconds for UV
	LastUpdated time.Time `json:"last_updated"`
}

// SensorLog represents temperature and humidity readings
type SensorLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Temperature float64   `json:"temperature" gorm:"not null"` // Temperature in Celsius
	Humidity    float64   `json:"humidity"`                    // Humidity in percentage (optional)
	RecordedAt  time.Time `json:"recorded_at" gorm:"index;not null"`
}
