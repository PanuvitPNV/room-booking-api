package middleware

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// TransactionMiddleware wraps HTTP requests in a database transaction
func TransactionMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip transaction for GET requests since they're read-only
			if c.Request().Method == "GET" {
				return next(c)
			}

			// Begin transaction
			tx := db.Begin()
			if tx.Error != nil {
				return echo.NewHTTPError(500, "Failed to begin transaction: "+tx.Error.Error())
			}

			// Store transaction in context
			c.Set("tx", tx)

			// Handle panics and rollback
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					panic(r) // Re-throw panic after rollback
				}
			}()

			// Process the request
			err := next(c)

			// Handle errors and rollback
			if err != nil {
				tx.Rollback()
				return err
			}

			// Commit the transaction
			if err := tx.Commit().Error; err != nil {
				return echo.NewHTTPError(500, "Failed to commit transaction: "+err.Error())
			}

			return nil
		}
	}
}

// GetTransaction retrieves the current transaction from the context
// Falls back to the main database connection if no transaction exists
func GetTransaction(c echo.Context) *gorm.DB {
	tx, ok := c.Get("tx").(*gorm.DB)
	if !ok {
		// Return the main DB connection if no transaction is found
		db, _ := c.Get("db").(*gorm.DB)
		return db
	}
	return tx
}
