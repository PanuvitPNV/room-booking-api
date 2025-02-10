package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// CustomErrorHandler handles validation errors
func CustomErrorHandler(err error, c echo.Context) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(ValidationErrors, 0)

		for _, err := range validationErrors {
			errors = append(errors, ValidationError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: err.Param(),
			})
		}

		return c.JSON(400, map[string]interface{}{
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	return err
}

// ValidateCreateBooking validates booking creation request
func ValidateCreateBooking(v *validator.Validate) validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		booking := sl.Current().Interface().(struct {
			CheckInDate  time.Time `json:"check_in_date"`
			CheckOutDate time.Time `json:"check_out_date"`
		})

		if !booking.CheckOutDate.After(booking.CheckInDate) {
			sl.ReportError(booking.CheckOutDate, "check_out_date", "CheckOutDate", "aftercheckin", "")
		}

		now := time.Now()
		if booking.CheckInDate.Before(now) {
			sl.ReportError(booking.CheckInDate, "check_in_date", "CheckInDate", "future", "")
		}
	}
}

// ValidateUpdateBooking validates booking update request
func ValidateUpdateBooking(v *validator.Validate) validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		booking := sl.Current().Interface().(struct {
			CheckInDate  *time.Time `json:"check_in_date"`
			CheckOutDate *time.Time `json:"check_out_date"`
		})

		if booking.CheckInDate != nil && booking.CheckOutDate != nil {
			if !booking.CheckOutDate.After(*booking.CheckInDate) {
				sl.ReportError(*booking.CheckOutDate, "check_out_date", "CheckOutDate", "aftercheckin", "")
			}
		}
	}
}

// ValidateRoomAvailability validates room availability request
func ValidateRoomAvailability(v *validator.Validate) validator.StructLevelFunc {
	return func(sl validator.StructLevel) {
		req := sl.Current().Interface().(struct {
			CheckInDate  time.Time `json:"check_in_date"`
			CheckOutDate time.Time `json:"check_out_date"`
		})

		if !req.CheckOutDate.After(req.CheckInDate) {
			sl.ReportError(req.CheckOutDate, "check_out_date", "CheckOutDate", "aftercheckin", "")
		}

		now := time.Now()
		if req.CheckInDate.Before(now) {
			sl.ReportError(req.CheckInDate, "check_in_date", "CheckInDate", "future", "")
		}
	}
}
