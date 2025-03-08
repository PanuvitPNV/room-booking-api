package utils

import (
	"os"
	"strconv"
	"sync"
	"time"
)

// DeadlockTesting controls deadlock testing functionality
var DeadlockTesting = struct {
	// Enabled indicates if deadlock testing mode is on
	Enabled bool

	// Lock sequence tracking for deadlock detection
	acquiredLocks    map[string]bool
	lockRequestOrder map[string][]string
	lockHistoryMutex sync.Mutex
}{
	Enabled:          false,
	acquiredLocks:    make(map[string]bool),
	lockRequestOrder: make(map[string][]string),
}

// Initialize deadlock testing settings from environment
func init() {
	// Check if deadlock test mode is enabled via environment variable
	if val, exists := os.LookupEnv("DEADLOCK_TEST_MODE"); exists {
		if enabled, err := strconv.ParseBool(val); err == nil {
			DeadlockTesting.Enabled = enabled
		}
	}
}

// DelayIfTesting introduces an artificial delay when in testing mode
func DelayIfTesting(duration time.Duration) {
	if DeadlockTesting.Enabled {
		time.Sleep(duration)
	}
}
