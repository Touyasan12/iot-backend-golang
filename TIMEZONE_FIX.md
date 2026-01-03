# 🕐 Timezone Issue - Backend Fix Required

## 🔴 Problem:
Ketika user membuat jadwal **08:00 WIB** di frontend, jadwal berjalan pada **15:00 WIB** (7 jam lebih lambat).

### Root Cause:
1. **Frontend** mengirim waktu: `"08:00"` (tanpa timezone info)
2. **Backend** menyimpan di database: `"08:00"` (OK)
3. **Scheduler** di backend berjalan dengan **UTC timezone** (server time)
4. Scheduler membandingkan `08:00` dengan waktu UTC server
5. Karena WIB = UTC+7, scheduler baru match saat UTC = 08:00, yang berarti WIB = 15:00

### Example:
```
User Input (Frontend): 08:00 WIB
Database: time = "08:00" (no timezone)
Scheduler Check: 
  - Server UTC time = 01:00 → Not match (08:00)
  - Server UTC time = 08:00 → MATCH! (tapi ini = 15:00 WIB ❌)
  
Expected: Scheduler should check at 01:00 UTC (= 08:00 WIB ✅)
```

## ✅ Solution for Backend (Go)

### Option 1: Set Application Timezone to WIB (Recommended)

**File: `main.go` or `scheduler/scheduler.go`**

```go
import (
    "time"
)

func init() {
    // Set application timezone to WIB (Asia/Jakarta)
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        log.Fatal("Failed to load timezone:", err)
    }
    time.Local = loc
}

func main() {
    // Your existing code...
}
```

### Option 2: Convert Time When Checking Schedule

**File: `scheduler/scheduler.go`**

```go
func checkSchedules() {
    // Get current time in WIB timezone
    wibLocation, _ := time.LoadLocation("Asia/Jakarta")
    nowWIB := time.Now().In(wibLocation)
    
    currentHour := nowWIB.Hour()
    currentMinute := nowWIB.Minute()
    dayName := nowWIB.Weekday().String()[:3] // Mon, Tue, etc.
    
    // Rest of your scheduler logic...
}
```

### Option 3: Store Timezone in Database

**Migration:**

```sql
-- Add timezone column to schedules
ALTER TABLE pakan_schedules ADD COLUMN timezone VARCHAR(50) DEFAULT 'Asia/Jakarta';
ALTER TABLE uv_schedules ADD COLUMN timezone VARCHAR(50) DEFAULT 'Asia/Jakarta';
```

**Code:**

```go
func checkSchedules() {
    var schedules []PakanSchedule
    db.Find(&schedules, "day_name = ? AND is_active = true", dayName)
    
    for _, schedule := range schedules {
        // Load schedule timezone (default to WIB)
        loc, _ := time.LoadLocation(schedule.Timezone)
        nowInScheduleTimezone := time.Now().In(loc)
        
        // Parse schedule time in its timezone
        scheduleTime, _ := time.ParseInLocation("15:04", schedule.Time, loc)
        
        if nowInScheduleTimezone.Hour() == scheduleTime.Hour() && 
           nowInScheduleTimezone.Minute() == scheduleTime.Minute() {
            // Execute schedule
        }
    }
}
```

## 🎯 Recommended Solution

**Use Option 1** - Set timezone di `init()` function:

```go
// File: main.go
package main

import (
    "log"
    "time"
    // ... other imports
)

func init() {
    // Set server timezone to WIB
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        log.Printf("Warning: Failed to load Asia/Jakarta timezone: %v", err)
        log.Println("Falling back to system timezone")
    } else {
        time.Local = loc
        log.Println("✅ Server timezone set to WIB (Asia/Jakarta)")
    }
}

func main() {
    // Your existing code...
}
```

## 🧪 Testing After Fix

After implementing the fix:

1. Create schedule for 08:00 WIB
2. Wait until 08:00 WIB (not 15:00)
3. Check if schedule executes correctly
4. Check history - should show 08:00 WIB in both API and UI

## 📝 Alternative: Backend Returns Timezone-Aware Timestamps

If you want to keep UTC in backend but show correct time in frontend:

**Backend (Go) - Add timezone to response:**

```go
type ActionHistoryResponse struct {
    ID            uint      `json:"id"`
    DeviceType    string    `json:"device_type"`
    TriggerSource string    `json:"trigger_source"`
    StartTime     time.Time `json:"start_time"`     // Already in UTC
    StartTimeWIB  string    `json:"start_time_wib"` // Add this
    // ... other fields
}

// In handler
wibLoc, _ := time.LoadLocation("Asia/Jakarta")
startTimeWIB := history.StartTime.In(wibLoc).Format("2006-01-02T15:04:05")

response.StartTimeWIB = startTimeWIB
```

But this is **NOT RECOMMENDED** because it doesn't fix the scheduler issue, only the display.

## ✅ Final Check

After implementing Option 1:

```bash
# SSH to Railway container
$ railway run bash

# Check current timezone
$ date
# Should show: Wed Nov 27 08:00:00 WIB 2025

# Check Go app timezone
$ curl http://localhost:8080/api/v1/debug/time
# Should return: {"server_time": "2025-11-27T08:00:00+07:00", "timezone": "WIB"}
```

---

**Priority:** HIGH 🔴  
**Complexity:** LOW ⚡  
**Estimated Time:** 5-10 minutes  
**Files to Change:** 1 file (`main.go` or `scheduler/scheduler.go`)
