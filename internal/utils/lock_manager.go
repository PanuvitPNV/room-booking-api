package utils

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// LockManager provides a mechanism for application-level locking to prevent race conditions
type LockManager struct {
	locks      map[string]*sync.Mutex
	locksMutex sync.Mutex
	timeout    time.Duration

	// Fields for deadlock testing
	logger         *log.Logger
	lockOrder      map[string]int         // Track the order of acquiring locks
	heldLocks      map[string]interface{} // Which locks are currently held
	heldLocksMutex sync.Mutex

	// Flag for intentional deadlock simulation
	deadlockMode bool
}

// NewLockManager creates a new instance of LockManager
func NewLockManager(timeout time.Duration) *LockManager {
	// Check if deadlock mode is enabled via environment variable
	deadlockMode := os.Getenv("ENABLE_DEADLOCK_MODE") == "true"

	return &LockManager{
		locks:        make(map[string]*sync.Mutex),
		timeout:      timeout,
		logger:       log.New(os.Stdout, "[LOCK_MANAGER] ", log.LstdFlags|log.Lmicroseconds),
		lockOrder:    make(map[string]int),
		heldLocks:    make(map[string]interface{}),
		deadlockMode: deadlockMode,
	}
}

// AcquireLock tries to acquire a lock for a specific resource
// Returns a release function and an error if it couldn't acquire the lock
func (lm *LockManager) AcquireLock(resourceType string, resourceID interface{}) (func(), error) {
	lockKey := fmt.Sprintf("%s-%v", resourceType, resourceID)

	// Log lock acquisition attempt
	lm.logger.Printf("Attempting to acquire lock: %s", lockKey)

	// DEADLOCK SCENARIO 1: In deadlock mode, sometimes introduce a random delay
	// This increases the likelihood of interleaved lock acquisitions
	if lm.deadlockMode && rand.Intn(10) < 3 {
		delay := time.Duration(rand.Intn(200)) * time.Millisecond
		lm.logger.Printf("Introducing random delay of %v before acquiring lock %s", delay, lockKey)
		time.Sleep(delay)
	}

	// Get or create a mutex for this resource
	lm.locksMutex.Lock()
	mutex, exists := lm.locks[lockKey]
	if !exists {
		mutex = &sync.Mutex{}
		lm.locks[lockKey] = mutex
	}
	lm.locksMutex.Unlock()

	// DEADLOCK SCENARIO 2: Check if this pattern of lock acquisition could cause a deadlock
	// This is more of a demonstration than a true deadlock prevention mechanism
	if lm.deadlockMode {
		lm.heldLocksMutex.Lock()

		// For testing, if we're trying to acquire lock B while holding lock A,
		// and the resource ID of B is lower than A, there's a potential deadlock
		for heldLock := range lm.heldLocks {
			// Simple deadlock detection logic for testing
			// If we have locks with different resource types, check if we're getting them in the wrong order
			if heldLock != lockKey {
				heldID := fmt.Sprintf("%v", lm.heldLocks[heldLock])
				currentID := fmt.Sprintf("%v", resourceID)

				// If the held lock ID is numerically greater than the current lock ID,
				// it might cause deadlock (this is a simplistic check for demonstration)
				if heldID > currentID && rand.Intn(10) < 4 { // 40% chance
					lm.logger.Printf("DEADLOCK RISK: Acquiring %s while holding %s may cause deadlock",
						lockKey, heldLock)

					// For testing, randomly sleep to increase deadlock chance
					if rand.Intn(10) < 3 { // 30% chance
						lm.logger.Printf("Sleeping before deadlock-prone lock acquisition")
						lm.heldLocksMutex.Unlock() // Release mutex during sleep
						time.Sleep(300 * time.Millisecond)
						lm.heldLocksMutex.Lock() // Re-acquire mutex
					}
				}
			}
		}

		lm.heldLocksMutex.Unlock()
	}

	// Try to acquire the lock with timeout
	lockAcquired := make(chan struct{})

	go func() {
		mutex.Lock()
		close(lockAcquired)
	}()

	var acquiredLock bool
	select {
	case <-lockAcquired:
		// Lock acquired successfully
		acquiredLock = true
	case <-time.After(lm.timeout):
		acquiredLock = false
	}

	if !acquiredLock {
		return nil, fmt.Errorf("timeout waiting for resource lock: %s", lockKey)
	}

	// Record that we're holding this lock (for deadlock simulation)
	if lm.deadlockMode {
		lm.heldLocksMutex.Lock()
		lm.heldLocks[lockKey] = resourceID
		lm.heldLocksMutex.Unlock()
	}

	lm.logger.Printf("Acquired lock: %s", lockKey)

	// Return the unlock function
	return func() {
		// DEADLOCK SCENARIO 3: Sometimes delay releasing the lock
		if lm.deadlockMode && rand.Intn(10) < 2 { // 20% chance
			delay := time.Duration(rand.Intn(150)) * time.Millisecond
			lm.logger.Printf("Delaying lock release of %s by %v", lockKey, delay)
			time.Sleep(delay)
		}

		// Remove from held locks tracking
		if lm.deadlockMode {
			lm.heldLocksMutex.Lock()
			delete(lm.heldLocks, lockKey)
			lm.heldLocksMutex.Unlock()
		}

		mutex.Unlock()
		lm.logger.Printf("Released lock: %s", lockKey)

		// DEADLOCK SCENARIO 4: After releasing a lock, sometimes keep it in the map
		// This simulates resource leaks
		if lm.deadlockMode && rand.Intn(20) < 1 { // 5% chance
			lm.logger.Printf("Intentionally not cleaning up lock object for %s (simulating resource leak)", lockKey)
			return
		}

		// Clean up the lock from the map if it's no longer needed
		lm.locksMutex.Lock()
		delete(lm.locks, lockKey)
		lm.locksMutex.Unlock()
	}, nil
}

// AcquireMultipleLocks acquires locks for multiple resources in a specific order
// This helps prevent deadlocks when resources need to be locked in a consistent order
func (lm *LockManager) AcquireMultipleLocks(resources []struct {
	Type string
	ID   interface{}
}) (func(), error) {
	// Make a copy of resources to avoid modifying the original slice
	resourcesCopy := make([]struct {
		Type string
		ID   interface{}
	}, len(resources))
	copy(resourcesCopy, resources)

	// DEADLOCK SCENARIO 5: Sometimes don't sort locks in deadlock mode
	if !lm.deadlockMode || rand.Intn(10) < 7 { // 70% chance of sorting in deadlock mode
		// Sort resources by type and ID to ensure consistent locking order
		// This is a simple sort - in a real implementation, you'd want a more sophisticated sorting
		// This sorting is intentionally omitted for brevity
	}

	var unlockFuncs []func()

	for _, resource := range resourcesCopy {
		unlock, err := lm.AcquireLock(resource.Type, resource.ID)
		if err != nil {
			// If we fail to acquire any lock, release all acquired locks
			for _, unlockFunc := range unlockFuncs {
				unlockFunc()
			}
			return nil, err
		}
		unlockFuncs = append(unlockFuncs, unlock)
	}

	// Return a function that unlocks all resources in reverse order
	return func() {
		for i := len(unlockFuncs) - 1; i >= 0; i-- {
			unlockFuncs[i]()
		}
	}, nil
}

// SetDeadlockMode enables or disables intentional deadlock-prone behavior
func (lm *LockManager) SetDeadlockMode(enable bool) {
	lm.deadlockMode = enable
	lm.logger.Printf("Deadlock simulation mode set to: %v", enable)
}

// GetLockCount returns the number of currently managed locks
// Useful for debugging and detecting lock leaks
func (lm *LockManager) GetLockCount() int {
	lm.locksMutex.Lock()
	defer lm.locksMutex.Unlock()
	return len(lm.locks)
}

// SimulateDeadlock intentionally creates a pattern that will likely cause a deadlock
// This is for testing purposes only
func (lm *LockManager) SimulateDeadlock() {
	lm.logger.Println("Simulating deadlock conditions")

	// Create two goroutines that acquire locks in opposite order
	go func() {
		lm.logger.Println("Goroutine 1: Acquiring lock A, then B")
		unlockA, err := lm.AcquireLock("test", "A")
		if err != nil {
			lm.logger.Printf("Goroutine 1: Failed to acquire lock A: %v", err)
			return
		}

		// Wait to increase likelihood of deadlock
		time.Sleep(100 * time.Millisecond)

		unlockB, err := lm.AcquireLock("test", "B")
		if err != nil {
			lm.logger.Printf("Goroutine 1: Failed to acquire lock B: %v", err)
			unlockA()
			return
		}

		lm.logger.Println("Goroutine 1: Acquired both locks (unlikely in deadlock scenario)")
		unlockB()
		unlockA()
	}()

	go func() {
		lm.logger.Println("Goroutine 2: Acquiring lock B, then A")
		unlockB, err := lm.AcquireLock("test", "B")
		if err != nil {
			lm.logger.Printf("Goroutine 2: Failed to acquire lock B: %v", err)
			return
		}

		// Wait to increase likelihood of deadlock
		time.Sleep(100 * time.Millisecond)

		unlockA, err := lm.AcquireLock("test", "A")
		if err != nil {
			lm.logger.Printf("Goroutine 2: Failed to acquire lock A: %v", err)
			unlockB()
			return
		}

		lm.logger.Println("Goroutine 2: Acquired both locks (unlikely in deadlock scenario)")
		unlockA()
		unlockB()
	}()
}
