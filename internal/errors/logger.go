package errors

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

type ErrorLogger struct {
	RequestID string    `json:"request_id"`
	Time      time.Time `json:"time"`
	Error     string    `json:"error"`
	Stack     string    `json:"stack,omitempty"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
}

func LogError(c echo.Context, err error) {
	logger := c.Logger()

	errLog := ErrorLogger{
		RequestID: c.Response().Header().Get(echo.HeaderXRequestID),
		Time:      time.Now(),
		Error:     err.Error(),
		Path:      c.Request().URL.Path,
		Method:    c.Request().Method,
	}

	if appErr, ok := err.(*AppError); ok {
		if appErr.Err != nil {
			errLog.Stack = fmt.Sprintf("%+v", appErr.Err)
		}
	}

	logger.Error(fmt.Sprintf("Error occurred: %+v", errLog))
}

// Middleware for logging errors
func ErrorLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				LogError(c, err)
			}
			return err
		}
	}
}
