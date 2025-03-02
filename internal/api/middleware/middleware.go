package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/panuvitpnv/room-booking-api/internal/config"
)

// CustomValidator is a custom validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the request data
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// SetupMiddleware configures middleware for the Echo instance
func SetupMiddleware(e *echo.Echo, config *config.Config) {
	// Add request ID
	e.Use(middleware.RequestID())

	// Add logger middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	// Add recover middleware
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
		LogLevel:  log.ERROR,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			log.Errorf("PANIC: %v\n%s", err, stack)
			return nil
		},
	}))

	// Add CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.Server.AllowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))

	// Add body limit
	e.Use(middleware.BodyLimit(config.Server.BodyLimit))

	// Add request timeout
	e.Use(TimeoutMiddleware(config.Server.Timeout))

	// Set up validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Handle errors
	e.HTTPErrorHandler = CustomHTTPErrorHandler
}

// TimeoutMiddleware handles request timeouts
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout*time.Second)
			defer cancel()

			// Update the request with the new context
			c.SetRequest(c.Request().WithContext(ctx))

			// Channel to track completion
			done := make(chan error)

			// Execute the next handler in a goroutine
			go func() {
				done <- next(c)
			}()

			// Wait for completion or timeout
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					return echo.NewHTTPError(http.StatusRequestTimeout, "Request timeout")
				}
				return ctx.Err()
			case err := <-done:
				return err
			}
		}
	}
}

// LoggingMiddleware logs transaction information
func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			txID := req.Header.Get(echo.HeaderXRequestID)
			if txID == "" {
				txID = res.Header().Get(echo.HeaderXRequestID)
			}

			log.Infof("[%s] Started %s %s", txID, req.Method, req.URL.Path)

			err := next(c)

			stop := time.Now()
			latency := stop.Sub(start)

			log.Infof("[%s] Completed %d %s in %s", txID, res.Status, http.StatusText(res.Status), latency)

			return err
		}
	}
}

// TransactionTracker is middleware for tracking database transactions
func TransactionTracker() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// This would be used to add transaction tracking headers or context values
			// For now we'll just add a header to identify transaction operations
			c.Response().Header().Set("X-Transaction-Tracked", "true")
			return next(c)
		}
	}
}

// CustomHTTPErrorHandler is a custom error handler for Echo
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"
	var details interface{}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
		details = he.Internal
	} else {
		// Log unexpected errors with stack trace
		buf := make([]byte, 2048)
		n := runtime.Stack(buf, false)
		log.Errorf("Unexpected error: %v\n%s", err, buf[:n])
	}

	// Don't log 404s
	if code != http.StatusNotFound {
		log.Errorf("Error: %s", message)
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]interface{}{
				"error":   message,
				"details": details,
			})
		}
		if err != nil {
			log.Errorf("Failed to send error response: %v", err)
		}
	}
}
