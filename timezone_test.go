package main

import (
	"log"
	"testing"
	"time"
)

func TestTimezone(t *testing.T) {
	// This will trigger the init() function
	log.Println("🕐 Current time.Local:", time.Local)
	log.Println("🕐 Current time.Now():", time.Now())
	log.Println("🕐 time.Now().Location():", time.Now().Location())
	
	// Expected: Asia/Jakarta (WIB)
	if time.Local.String() != "Asia/Jakarta" {
		t.Errorf("Expected timezone Asia/Jakarta, got %s", time.Local.String())
	}
	
	// Test: Create time without explicit timezone - should be WIB
	testTime := time.Now()
	log.Printf("🕐 Test time: %s (zone: %s)", testTime.Format("15:04:05 MST"), testTime.Location())
	
	// Compare with explicit WIB time
	wibLoc, _ := time.LoadLocation("Asia/Jakarta")
	wibTime := time.Now().In(wibLoc)
	log.Printf("🕐 WIB time:  %s (zone: %s)", wibTime.Format("15:04:05 MST"), wibTime.Location())
	
	// They should be the same
	if testTime.Location().String() != wibTime.Location().String() {
		t.Errorf("Expected time.Now() to be in WIB, got %s", testTime.Location())
	}
}
