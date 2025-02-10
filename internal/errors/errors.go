package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Op      string `json:"op,omitempty"` // Operation that failed (e.g., "db.CreateBooking")
	Err     error  `json:"-"`            // Original error (if any)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Common application errors
var (
	// Not Found errors
	ErrRoomNotFound     = &AppError{Code: http.StatusNotFound, Message: "room not found"}
	ErrGuestNotFound    = &AppError{Code: http.StatusNotFound, Message: "guest not found"}
	ErrBookingNotFound  = &AppError{Code: http.StatusNotFound, Message: "booking not found"}
	ErrRoomTypeNotFound = &AppError{Code: http.StatusNotFound, Message: "room type not found"}

	// Conflict errors
	ErrRoomNotAvailable = &AppError{Code: http.StatusConflict, Message: "room not available for selected dates"}
	ErrDuplicateRoom    = &AppError{Code: http.StatusConflict, Message: "room already exists"}
	ErrDuplicateGuest   = &AppError{Code: http.StatusConflict, Message: "guest with this email or phone already exists"}

	// Bad Request errors
	ErrInvalidDateRange = &AppError{Code: http.StatusBadRequest, Message: "invalid date range"}
	ErrInvalidData      = &AppError{Code: http.StatusBadRequest, Message: "invalid data provided"}

	// Internal Server errors
	ErrDatabase = &AppError{Code: http.StatusInternalServerError, Message: "database error occurred"}
	ErrInternal = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
)

// Error constructors
func NewError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func WrapError(err error, message string) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: message,
	}
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func NewOperationError(op string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "operation failed",
		Op:      op,
		Err:     err,
	}
}
