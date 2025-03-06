package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// Logger is a custom logger that writes to both console and file
type Logger struct {
	*log.Logger
	file      *os.File
	mu        sync.Mutex
	filePath  string
	maxSize   int64 // Maximum size in bytes before rotation
	maxBackup int   // Maximum number of old log files to retain
}

// NewLogger creates a new logger that writes to both console and file
func NewLogger(logDir string) (*Logger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with current date
	currentTime := time.Now()
	fileName := fmt.Sprintf("hotel_booking_%s.log", currentTime.Format("2006-01-02"))
	filePath := filepath.Join(logDir, fileName)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Create new logger
	l := log.New("hotel-booking")
	l.SetOutput(multiWriter)
	l.SetLevel(log.INFO)
	l.SetHeader("${time_rfc3339} [${level}] ${short_file}:${line}: ")

	return &Logger{
		Logger:    l,
		file:      file,
		filePath:  filePath,
		maxSize:   10 * 1024 * 1024, // 10MB
		maxBackup: 7,                // Keep 7 days of logs
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// LogRequest logs HTTP request details
func (l *Logger) LogRequest(c echo.Context, latency time.Duration) {
	req := c.Request()
	res := c.Response()

	// Get client IP
	clientIP := c.RealIP()

	// Create log message
	msg := fmt.Sprintf("%s %s %s %d %s %s",
		req.Method,
		req.URL.Path,
		req.Proto,
		res.Status,
		latency.String(),
		clientIP,
	)

	// Log based on status code
	if res.Status >= 500 {
		l.Error(msg)
	} else if res.Status >= 400 {
		l.Warn(msg)
	} else {
		l.Info(msg)
	}
}

// LogTransaction logs transaction details
func (l *Logger) LogTransaction(txType, operation string, entityID interface{}, details string, duration time.Duration, success bool) {
	// Get file and line where the log is called from
	_, file, line, _ := runtime.Caller(1)
	shortFile := filepath.Base(file)

	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	// Create log message
	msg := fmt.Sprintf("[TX] type=%s operation=%s entity=%v duration=%s status=%s details=%s",
		txType, operation, entityID, duration.String(), status, details)

	// Log based on success/failure
	if !success {
		l.Error(msg, shortFile, line)
	} else {
		l.Info(msg, shortFile, line)
	}

	// Check if we need to rotate the log file
	l.checkRotate()
}

// LogConcurrency logs concurrency-related events
func (l *Logger) LogConcurrency(event, resourceType string, resourceID interface{}, details string) {
	// Get file and line where the log is called from
	_, file, line, _ := runtime.Caller(1)
	shortFile := filepath.Base(file)

	// Create log message
	msg := fmt.Sprintf("[CONCURRENCY] event=%s resource_type=%s resource_id=%v details=%s",
		event, resourceType, resourceID, details)

	l.Info(msg, shortFile, line)
}

// checkRotate checks if log file should be rotated based on size
func (l *Logger) checkRotate() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	// Get file info
	info, err := l.file.Stat()
	if err != nil {
		l.Error("Failed to get log file info:", err)
		return
	}

	// Check if file exceeds max size
	if info.Size() < l.maxSize {
		return
	}

	// Close current file
	l.file.Close()

	// Rotate file
	now := time.Now()
	backupName := fmt.Sprintf("%s.%s", l.filePath, now.Format("20060102-150405"))
	if err := os.Rename(l.filePath, backupName); err != nil {
		l.Error("Failed to rotate log file:", err)
		return
	}

	// Open new file
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.Error("Failed to open new log file:", err)
		return
	}

	// Update writer
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.SetOutput(multiWriter)
	l.file = file

	// Delete old backup files
	l.cleanOldLogs()
}

// cleanOldLogs removes log files older than maxBackup days
func (l *Logger) cleanOldLogs() {
	dir := filepath.Dir(l.filePath)
	pattern := filepath.Join(dir, "hotel_booking_*.log*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		l.Error("Failed to find old log files:", err)
		return
	}

	if len(matches) <= l.maxBackup {
		return
	}

	// Sort files by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	files := make([]fileInfo, 0, len(matches))
	for _, path := range matches {
		// Skip current log file
		if path == l.filePath {
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		files = append(files, fileInfo{path: path, modTime: info.ModTime()})
	}

	// Sort files by modification time (oldest first)
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files that exceed maxBackup
	for i := 0; i < len(files)-l.maxBackup; i++ {
		os.Remove(files[i].path)
	}
}

// FormatLog formats a log message with timestamp and optional context
func (l *Logger) FormatLog(level, msg string, ctx ...interface{}) string {
	timestamp := time.Now().Format(time.RFC3339)

	if len(ctx) > 0 {
		contextStr := make([]string, len(ctx))
		for i, c := range ctx {
			contextStr[i] = fmt.Sprintf("%v", c)
		}
		return fmt.Sprintf("%s [%s] %s | Context: %s",
			timestamp, strings.ToUpper(level), msg, strings.Join(contextStr, ", "))
	}

	return fmt.Sprintf("%s [%s] %s", timestamp, strings.ToUpper(level), msg)
}
