package concurrency

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RetryableTransaction executes a function within a transaction with retry capabilities
// This is useful for handling transient database errors or deadlocks
func RetryableTransaction(
	db *gorm.DB,
	maxRetries int,
	backoff time.Duration,
	fn func(*gorm.DB) error,
) error {
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = db.Transaction(func(tx *gorm.DB) error {
			return fn(tx)
		})

		// If successful or not a retryable error, return immediately
		if err == nil || !isRetryableError(err) {
			return err
		}

		// Wait before retrying with exponential backoff
		sleepTime := backoff * time.Duration(1<<uint(attempt))
		time.Sleep(sleepTime)
	}

	// Return the last error if all retries failed
	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, err)
}

// isRetryableError determines if a database error can be retried
// For PostgreSQL, this typically includes deadlocks and serialization failures
func isRetryableError(err error) bool {
	// Common PostgreSQL error codes that indicate retryable conditions:
	// 40001 - serialization_failure
	// 40P01 - deadlock_detected
	errStr := err.Error()
	return strings.Contains(errStr, "deadlock") ||
		strings.Contains(errStr, "serialization") ||
		strings.Contains(errStr, "could not serialize access") ||
		strings.Contains(errStr, "40001") ||
		strings.Contains(errStr, "40P01")
}

// WithSelectForUpdate adds FOR UPDATE clause to a query for pessimistic locking
func WithSelectForUpdate(tx *gorm.DB) *gorm.DB {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"})
}

// WithSelectForShare adds FOR SHARE clause to a query for shared locking
// This prevents other transactions from updating records but allows reading
func WithSelectForShare(tx *gorm.DB) *gorm.DB {
	return tx.Clauses(clause.Locking{Strength: "SHARE"})
}

// WithSkipLocked adds SKIP LOCKED to a query to skip locked rows
// Useful for queue-like processing where we want to fetch unlocked rows only
func WithSkipLocked(tx *gorm.DB) *gorm.DB {
	return tx.Clauses(clause.Locking{
		Strength: "UPDATE",
		Options:  "SKIP LOCKED",
	})
}

// OptimisticLockingDemo demonstrates optimistic locking
// In GORM, this typically uses a version field to detect concurrent updates
func OptimisticLockingDemo(db *gorm.DB, roomType *struct {
	ID      int `gorm:"primaryKey"`
	Name    string
	Version int `gorm:"version"` // Version field for optimistic locking
}) error {
	// The version field will automatically be incremented by GORM when saving
	// If another transaction has modified this record, the version check will fail
	result := db.Save(roomType)

	if result.Error != nil {
		return fmt.Errorf("optimistic lock error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed, record was updated by another transaction")
	}

	return nil
}
