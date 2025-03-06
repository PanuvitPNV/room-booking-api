package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TransactionLogger provides specialized logging for transaction operations
type TransactionLogger struct {
	mu           sync.Mutex
	file         *os.File
	humanFile    *os.File
	detailsLevel int // 1=basic, 2=detailed, 3=verbose
}

// NewTransactionLogger creates a new transaction logger
func NewTransactionLogger(logDir string, detailsLevel int) (*TransactionLogger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create machine-readable log file
	currentTime := time.Now()
	fileName := fmt.Sprintf("transactions_%s.log", currentTime.Format("2006-01-02"))
	filePath := filepath.Join(logDir, fileName)

	// Create human-readable log file
	humanFileName := fmt.Sprintf("transactions_readable_%s.log", currentTime.Format("2006-01-02"))
	humanFilePath := filepath.Join(logDir, humanFileName)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	humanFile, err := os.OpenFile(humanFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to open human log file: %w", err)
	}

	return &TransactionLogger{
		file:         file,
		humanFile:    humanFile,
		detailsLevel: detailsLevel,
	}, nil
}

// Close closes the log files
func (l *TransactionLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.file.Close(); err != nil {
		return err
	}
	return l.humanFile.Close()
}

// LogTransactionStart logs the start of a transaction
func (l *TransactionLogger) LogTransactionStart(txID, clientID, operationType string, details map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Machine-readable format
	logLine := fmt.Sprintf("%s [TX-START] tx_id=%s client=%s operation=%s",
		timestamp, txID, clientID, operationType)

	for k, v := range details {
		logLine += fmt.Sprintf(" %s=%v", k, v)
	}

	l.file.WriteString(logLine + "\n")

	// Human-readable format
	humanLog := fmt.Sprintf("[%s] Transaction STARTED:\n", timestamp)
	humanLog += fmt.Sprintf("  Transaction ID: %s\n", txID)
	humanLog += fmt.Sprintf("  Client ID: %s\n", clientID)
	humanLog += fmt.Sprintf("  Operation: %s\n", operationType)

	if l.detailsLevel >= 2 && len(details) > 0 {
		humanLog += "  Details:\n"
		for k, v := range details {
			humanLog += fmt.Sprintf("    - %s: %v\n", k, v)
		}
	}

	l.humanFile.WriteString(humanLog + "\n")
}

// LogTransactionEnd logs the end of a transaction
func (l *TransactionLogger) LogTransactionEnd(txID string, success bool, duration time.Duration, result map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	// Machine-readable format
	logLine := fmt.Sprintf("%s [TX-END] tx_id=%s status=%s duration=%s",
		timestamp, txID, status, duration)

	for k, v := range result {
		logLine += fmt.Sprintf(" %s=%v", k, v)
	}

	l.file.WriteString(logLine + "\n")

	// Human-readable format
	humanLog := fmt.Sprintf("[%s] Transaction COMPLETED:\n", timestamp)
	humanLog += fmt.Sprintf("  Transaction ID: %s\n", txID)
	humanLog += fmt.Sprintf("  Status: %s\n", status)
	humanLog += fmt.Sprintf("  Duration: %s\n", duration)

	if l.detailsLevel >= 2 && len(result) > 0 {
		humanLog += "  Result:\n"
		for k, v := range result {
			humanLog += fmt.Sprintf("    - %s: %v\n", k, v)
		}
	}

	l.humanFile.WriteString(humanLog + "\n")
}

// LogConcurrencyEvent logs a concurrency-related event
func (l *TransactionLogger) LogConcurrencyEvent(txID, eventType, resourceType string, resourceID interface{}, details string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Machine-readable format
	logLine := fmt.Sprintf("%s [CONCURRENCY] tx_id=%s event=%s resource_type=%s resource_id=%v details=%s",
		timestamp, txID, eventType, resourceType, resourceID, details)

	l.file.WriteString(logLine + "\n")

	// Human-readable format
	humanLog := fmt.Sprintf("[%s] Concurrency Event:\n", timestamp)
	humanLog += fmt.Sprintf("  Transaction ID: %s\n", txID)
	humanLog += fmt.Sprintf("  Event: %s\n", eventType)
	humanLog += fmt.Sprintf("  Resource: %s (ID: %v)\n", resourceType, resourceID)

	if l.detailsLevel >= 2 {
		humanLog += fmt.Sprintf("  Details: %s\n", details)
	}

	l.humanFile.WriteString(humanLog + "\n")
}

// LogLockAcquired logs when a lock is acquired
func (l *TransactionLogger) LogLockAcquired(txID, resourceType string, resourceID interface{}, lockType string, waitTime time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Machine-readable format
	logLine := fmt.Sprintf("%s [LOCK-ACQUIRED] tx_id=%s resource_type=%s resource_id=%v lock_type=%s wait_time=%s",
		timestamp, txID, resourceType, resourceID, lockType, waitTime)

	l.file.WriteString(logLine + "\n")

	// Human-readable format
	humanLog := fmt.Sprintf("[%s] Lock Acquired:\n", timestamp)
	humanLog += fmt.Sprintf("  Transaction ID: %s\n", txID)
	humanLog += fmt.Sprintf("  Resource: %s (ID: %v)\n", resourceType, resourceID)
	humanLog += fmt.Sprintf("  Lock Type: %s\n", lockType)
	humanLog += fmt.Sprintf("  Wait Time: %s\n", waitTime)

	l.humanFile.WriteString(humanLog + "\n")
}

// LogLockReleased logs when a lock is released
func (l *TransactionLogger) LogLockReleased(txID, resourceType string, resourceID interface{}, lockType string, heldTime time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Machine-readable format
	logLine := fmt.Sprintf("%s [LOCK-RELEASED] tx_id=%s resource_type=%s resource_id=%v lock_type=%s held_time=%s",
		timestamp, txID, resourceType, resourceID, lockType, heldTime)

	l.file.WriteString(logLine + "\n")

	// Human-readable format if verbose mode
	if l.detailsLevel >= 3 {
		humanLog := fmt.Sprintf("[%s] Lock Released:\n", timestamp)
		humanLog += fmt.Sprintf("  Transaction ID: %s\n", txID)
		humanLog += fmt.Sprintf("  Resource: %s (ID: %v)\n", resourceType, resourceID)
		humanLog += fmt.Sprintf("  Lock Type: %s\n", lockType)
		humanLog += fmt.Sprintf("  Held Time: %s\n", heldTime)

		l.humanFile.WriteString(humanLog + "\n")
	}
}

// LogConflict logs when a conflict is detected between transactions
func (l *TransactionLogger) LogConflict(txID1, txID2, resourceType string, resourceID interface{}, resolution string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Machine-readable format
	logLine := fmt.Sprintf("%s [CONFLICT] tx_id1=%s tx_id2=%s resource_type=%s resource_id=%v resolution=%s",
		timestamp, txID1, txID2, resourceType, resourceID, resolution)

	l.file.WriteString(logLine + "\n")

	// Human-readable format - conflicts are always important, so always log them
	humanLog := fmt.Sprintf("[%s] ⚠️ CONFLICT DETECTED:\n", timestamp)
	humanLog += fmt.Sprintf("  Between Transaction: %s\n", txID1)
	humanLog += fmt.Sprintf("  And Transaction: %s\n", txID2)
	humanLog += fmt.Sprintf("  Resource: %s (ID: %v)\n", resourceType, resourceID)
	humanLog += fmt.Sprintf("  Resolution: %s\n", resolution)

	l.humanFile.WriteString(humanLog + "\n")
}

// LogBookingConflict logs a specific booking conflict for easy demonstration
func (l *TransactionLogger) LogBookingConflict(clientID1, clientID2 string, room int, checkInDate, checkOutDate time.Time, resolution string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Human-readable format designed for demonstration
	humanLog := fmt.Sprintf("\n==================================================\n")
	humanLog += fmt.Sprintf("🔴 BOOKING CONFLICT DEMONSTRATION [%s]\n", timestamp)
	humanLog += fmt.Sprintf("==================================================\n")
	humanLog += fmt.Sprintf("Two clients attempted to book the same room for overlapping dates:\n\n")
	humanLog += fmt.Sprintf("Client 1: %s\n", clientID1)
	humanLog += fmt.Sprintf("Client 2: %s\n", clientID2)
	humanLog += fmt.Sprintf("Room: %d\n", room)
	humanLog += fmt.Sprintf("Dates: %s to %s\n\n", checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02"))
	humanLog += fmt.Sprintf("TRANSACTION MANAGEMENT RESOLUTION: %s\n", resolution)
	humanLog += fmt.Sprintf("==================================================\n\n")

	l.humanFile.WriteString(humanLog)

	// Also log to machine-readable log
	l.file.WriteString(fmt.Sprintf("%s [BOOKING-CONFLICT] client1=%s client2=%s room=%d check_in=%s check_out=%s resolution=%s\n",
		timestamp, clientID1, clientID2, room, checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02"), resolution))
}
