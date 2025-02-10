package errors

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func ErrorHandler(err error, c echo.Context) error {
	var response ErrorResponse

	switch e := err.(type) {
	case *AppError:
		// Handle our custom AppError
		response = ErrorResponse{
			Code:    e.Code,
			Message: e.Message,
			Details: map[string]string{
				"detail": e.Detail,
				"op":     e.Op,
			},
		}
		return c.JSON(e.Code, response)

	case *echo.HTTPError:
		// Handle Echo's built-in errors
		response = ErrorResponse{
			Code:    e.Code,
			Message: fmt.Sprintf("%v", e.Message),
		}
		return c.JSON(e.Code, response)

	case validator.ValidationErrors:
		// Handle validation errors
		details := make([]map[string]string, 0)
		for _, err := range e {
			details = append(details, map[string]string{
				"field":   err.Field(),
				"message": fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", err.Field(), err.Tag()),
			})
		}
		response = ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Validation failed",
			Details: details,
		}
		return c.JSON(http.StatusBadRequest, response)

	default:
		// Handle unknown errors
		response = ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
}

// Middleware for handling panics
func RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					var err error
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("%v", r)
					}

					// Log the error
					LogError(c, err)

					// Return 500 error to client
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Code:    http.StatusInternalServerError,
						Message: "Internal server error",
						Details: err.Error(),
					})
				}
			}()
			return next(c)
		}
	}
}
