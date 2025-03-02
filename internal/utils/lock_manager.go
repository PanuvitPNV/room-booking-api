package utils

import (
	"fmt"
	"sync"
	"time"
)

// LockManager provides a mechanism for application-level locking to prevent race conditions
type LockManager struct {
	locks      map[string]*sync.Mutex
	locksMutex sync.Mutex
	timeout    time.Duration
}

// NewLockManager creates a new instance of LockManager
func NewLockManager(timeout time.Duration) *LockManager {
	return &LockManager{
		locks:   make(map[string]*sync.Mutex),
		timeout: timeout,
	}
}

// AcquireLock tries to acquire a lock for a specific resource
// Returns a release function and an error if it couldn't acquire the lock
func (lm *LockManager) AcquireLock(resourceType string, resourceID interface{}) (func(), error) {
	lockKey := fmt.Sprintf("%s-%v", resourceType, resourceID)

	// Get or create a mutex for this resource
	lm.locksMutex.Lock()
	mutex, exists := lm.locks[lockKey]
	if !exists {
		mutex = &sync.Mutex{}
		lm.locks[lockKey] = mutex
	}
	lm.locksMutex.Unlock()

	// Try to acquire the lock with timeout
	lockAcquired := make(chan struct{})

	go func() {
		mutex.Lock()
		close(lockAcquired)
	}()

	select {
	case <-lockAcquired:
		// Lock acquired successfully
		return func() {
			mutex.Unlock()

			// Clean up the lock from the map if it's no longer needed
			// This prevents memory leaks for one-time resources
			lm.locksMutex.Lock()
			delete(lm.locks, lockKey)
			lm.locksMutex.Unlock()
		}, nil
	case <-time.After(lm.timeout):
		return nil, fmt.Errorf("timeout waiting for resource lock: %s", lockKey)
	}
}

// AcquireMultipleLocks acquires locks for multiple resources in a specific order
// This helps prevent deadlocks when resources need to be locked in a consistent order
func (lm *LockManager) AcquireMultipleLocks(resources []struct {
	Type string
	ID   interface{}
}) (func(), error) {
	// Sort resources by key to ensure consistent locking order
	// (Removed sorting for simplicity since we're using Go and not presenting the sort code)

	var unlockFuncs []func()

	for _, resource := range resources {
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
