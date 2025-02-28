package concurrency

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// LockType defines the type of lock to acquire
type LockType int

const (
	// SharedLock allows multiple transactions to read a resource
	SharedLock LockType = iota
	// ExclusiveLock prevents other transactions from reading or writing to a resource
	ExclusiveLock
)

// RoomLock handles locking rooms for booking operations
type RoomLock struct {
	db *gorm.DB
}

// NewRoomLock creates a new RoomLock instance
func NewRoomLock(db *gorm.DB) *RoomLock {
	return &RoomLock{db: db}
}

// LockRoom acquires a lock on a room for a specified date range
// This prevents concurrent bookings of the same room for overlapping dates
func (rl *RoomLock) LockRoom(ctx context.Context, tx *gorm.DB, roomNum int, lockType LockType) error {
	// Execute advisory lock - uses room number as the lock key
	// This will block until the lock is acquired or context is cancelled
	query := fmt.Sprintf("SELECT pg_advisory_xact_lock(%d)", roomNum)
	if lockType == SharedLock {
		query = fmt.Sprintf("SELECT pg_advisory_xact_lock_shared(%d)", roomNum)
	}

	result := tx.Exec(query)
	return result.Error
}

// LockRoomDateRange locks a room for a specific date range
// This is more specific than LockRoom and prevents concurrent bookings for overlapping dates
func (rl *RoomLock) LockRoomDateRange(
	ctx context.Context,
	tx *gorm.DB,
	roomNum int,
	checkInDate, checkOutDate time.Time,
) error {
	// Add timeout to context to prevent indefinite waiting
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Lock all room status records for the given date range
	err := tx.Exec(`
		SELECT rs.room_num 
		FROM room_status rs 
		WHERE rs.room_num = ? AND rs.calendar BETWEEN ? AND ?
		FOR UPDATE
	`, roomNum, checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02")).Error

	if err != nil {
		return fmt.Errorf("failed to lock room date range: %w", err)
	}

	return nil
}

// TryLockWithTimeout attempts to acquire a lock with a timeout
// Returns true if lock was acquired, false if timeout occurred
func (rl *RoomLock) TryLockWithTimeout(
	ctx context.Context,
	tx *gorm.DB,
	roomNum int,
	checkInDate, checkOutDate time.Time,
	timeout time.Duration,
) (bool, error) {
	// Create a context with timeout
	lockCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Try to acquire lock in a non-blocking way
	lockChan := make(chan error, 1)

	go func() {
		err := rl.LockRoomDateRange(lockCtx, tx, roomNum, checkInDate, checkOutDate)
		lockChan <- err
	}()

	// Wait for either lock acquisition or timeout
	select {
	case err := <-lockChan:
		if err != nil {
			return false, err
		}
		return true, nil
	case <-lockCtx.Done():
		if lockCtx.Err() == context.DeadlineExceeded {
			return false, nil // Timeout occurred, lock not acquired
		}
		return false, lockCtx.Err()
	}
}
