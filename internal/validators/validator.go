package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	validate := validator.New()

	// Register custom validations
	registerCustomValidations(validate)

	return &CustomValidator{
		validator: validate,
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}
	return nil
}

// Register custom validation rules
func registerCustomValidations(v *validator.Validate) {
	// Custom date validation
	v.RegisterValidation("futuredate", validateFutureDate)
	v.RegisterValidation("daterange", validateDateRange)

	// Custom room validations
	v.RegisterValidation("roomstatus", validateRoomStatus)

	// Custom booking validations
	v.RegisterValidation("bookingdates", validateBookingDates)
}
