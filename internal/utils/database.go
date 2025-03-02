package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DB holds the global database connection
var DB *gorm.DB

// SetDB sets the global DB instance
func SetDB(db *gorm.DB) {
	DB = db
}

// TransactionFn defines a function that executes within a transaction
type TransactionFn func(tx *gorm.DB) error

// WithTransaction creates a new transaction and handles commit/rollback
func WithTransaction(ctx context.Context, fn TransactionFn) error {
	tx := DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Transaction panic: %v", r)
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// RetryFn defines a function that can be retried
type RetryFn func() error

// RunWithRetry executes the given function with retries in case of concurrent transaction errors
func RunWithRetry(attempts int, fn RetryFn) error {
	var err error

	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil // Success, no need to retry
		}

		// If this is a transaction conflict error, wait and retry
		if isTransactionConflictError(err) {
			backoff := time.Duration(50*(i+1)) * time.Millisecond
			time.Sleep(backoff)
			continue
		}

		// If it's not a transaction conflict error, return immediately
		return err
	}

	return fmt.Errorf("operation failed after %d attempts: %w", attempts, err)
}

// isTransactionConflictError determines if an error is related to transaction conflicts
func isTransactionConflictError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common PostgreSQL transaction error messages
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "deadlock detected") ||
		strings.Contains(errorMsg, "could not serialize access") ||
		strings.Contains(errorMsg, "serialization failure") ||
		strings.Contains(errorMsg, "concurrent update") ||
		errors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueViolation checks if the error is a database unique constraint violation
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "duplicate key")
}

// ErrOptimisticLock is returned when an optimistic lock version check fails
var ErrOptimisticLock = errors.New("record was updated by another transaction")

// BeginTransaction starts a new database transaction
func BeginTransaction() (*gorm.DB, error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

// CommitTransaction commits a database transaction
func CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTransaction rolls back a database transaction
func RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}
