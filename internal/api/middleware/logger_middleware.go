package middleware

import (
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/utils"

	"github.com/labstack/echo/v4"
)

// RequestLoggerMiddleware logs all HTTP requests
func RequestLoggerMiddleware(logger *utils.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Log request details
			logger.LogRequest(c, latency)

			return err
		}
	}
}

// TransactionLoggerMiddleware tracks transaction operations
func TransactionLoggerMiddleware(logger *utils.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add transaction ID to context
			txID := utils.GenerateTransactionID()
			c.Set("transaction_id", txID)
			c.Response().Header().Set("X-Transaction-ID", txID)

			// Start timing
			start := time.Now()

			// Process request
			err := next(c)

			// Get response status
			status := c.Response().Status
			success := status >= 200 && status < 400

			// Calculate duration
			duration := time.Since(start)

			// Log transaction details for write operations
			method := c.Request().Method
			if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
				path := c.Request().URL.Path

				var txType string
				if path == "/api/v1/bookings" && method == "POST" {
					txType = "CREATE_BOOKING"
				} else if path == "/api/v1/receipts" && method == "POST" {
					txType = "PROCESS_PAYMENT"
				} else if path == "/api/v1/receipts/refund" && method == "POST" {
					txType = "PROCESS_REFUND"
				} else if path == "/api/v1/bookings" && method == "PUT" {
					txType = "UPDATE_BOOKING"
				} else if path == "/api/v1/bookings" && method == "DELETE" {
					txType = "CANCEL_BOOKING"
				} else {
					txType = "OTHER"
				}

				details := c.Request().URL.String()
				if err != nil {
					details += " | Error: " + err.Error()
				}

				logger.LogTransaction(txType, method, txID, details, duration, success)
			}

			return err
		}
	}
}
