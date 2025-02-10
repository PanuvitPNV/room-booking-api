package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// validateFutureDate checks if a date is in the future
func validateFutureDate(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return date.After(time.Now())
}

// validateDateRange checks if check-out is after check-in
func validateDateRange(fl validator.FieldLevel) bool {
	checkIn, ok := fl.Parent().FieldByName("CheckInDate").Interface().(time.Time)
	if !ok {
		return false
	}

	checkOut, ok := fl.Parent().FieldByName("CheckOutDate").Interface().(time.Time)
	if !ok {
		return false
	}

	return checkOut.After(checkIn)
}

// validateRoomStatus checks if room status is valid
func validateRoomStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := map[string]bool{
		"Available": true,
		"Occupied":  true,
	}
	return validStatuses[status]
}

// validateBookingDates checks if booking dates are valid
func validateBookingDates(fl validator.FieldLevel) bool {
	checkIn, ok := fl.Parent().FieldByName("CheckInDate").Interface().(time.Time)
	if !ok {
		return false
	}

	checkOut, ok := fl.Parent().FieldByName("CheckOutDate").Interface().(time.Time)
	if !ok {
		return false
	}

	// Check if dates are in the future
	now := time.Now()
	if checkIn.Before(now) || checkOut.Before(now) {
		return false
	}

	// Check if check-out is after check-in
	if !checkOut.After(checkIn) {
		return false
	}

	// Check if booking duration is within limits (e.g., max 30 days)
	maxDuration := 30 * 24 * time.Hour
	return checkOut.Sub(checkIn) <= maxDuration
}
