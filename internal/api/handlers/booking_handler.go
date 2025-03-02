package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// BookingHandler handles HTTP requests related to bookings
type BookingHandler struct {
	bookingService *services.BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBookingRequest represents a request to create a booking
type CreateBookingRequest struct {
	BookingName  string    `json:"booking_name" validate:"required"`
	RoomNum      int       `json:"room_num" validate:"required"`
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// CreateBooking handles the creation of a new booking
// @Summary Create a new booking
// @Description Create a new booking for a room
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Booking details"
// @Success 201 {object} models.Booking
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	booking := models.Booking{
		BookingName:  req.BookingName,
		RoomNum:      req.RoomNum,
		CheckInDate:  req.CheckInDate,
		CheckOutDate: req.CheckOutDate,
	}

	if err := h.bookingService.CreateBooking(ctx, &booking); err != nil {
		log.Errorf("Failed to create booking: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create booking: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, booking)
}

// GetBooking retrieves a booking by ID
// @Summary Get a booking
// @Description Get a booking by ID
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} models.Booking
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid booking ID",
		})
	}

	booking, err := h.bookingService.GetBookingByID(ctx, id)
	if err != nil {
		log.Errorf("Failed to get booking: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Booking not found",
		})
	}

	return c.JSON(http.StatusOK, booking)
}

// CancelBooking cancels a booking
// @Summary Cancel a booking
// @Description Cancel a booking by ID
// @Tags bookings
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bookings/{id} [delete]
func (h *BookingHandler) CancelBooking(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid booking ID",
		})
	}

	if err := h.bookingService.CancelBooking(ctx, id); err != nil {
		log.Errorf("Failed to cancel booking: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to cancel booking: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Booking cancelled successfully",
	})
}

// GetAvailableRoomsRequest represents a request to find available rooms
type GetAvailableRoomsRequest struct {
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// GetAvailableRooms retrieves available rooms for a date range
// @Summary Get available rooms
// @Description Get available rooms for a date range
// @Tags rooms
// @Accept json
// @Produce json
// @Param dates body GetAvailableRoomsRequest true "Date range"
// @Success 200 {array} models.Room
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rooms/available [post]
func (h *BookingHandler) GetAvailableRooms(c echo.Context) error {
	ctx := c.Request().Context()

	var req GetAvailableRoomsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	rooms, err := h.bookingService.GetAvailableRooms(ctx, req.CheckInDate, req.CheckOutDate)
	if err != nil {
		log.Errorf("Failed to get available rooms: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get available rooms: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, rooms)
}

// UpdateBookingRequest represents a request to update a booking
type UpdateBookingRequest struct {
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// UpdateBooking updates a booking
// @Summary Update a booking
// @Description Update a booking's dates
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param booking body UpdateBookingRequest true "New booking dates"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bookings/{id} [put]
func (h *BookingHandler) UpdateBooking(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid booking ID",
		})
	}

	var req UpdateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	err = h.bookingService.UpdateBooking(ctx, id, req.CheckInDate, req.CheckOutDate)
	if err != nil {
		log.Errorf("Failed to update booking: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update booking: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Booking updated successfully",
	})
}

// GetBookingsByDateRangeRequest represents a request to find bookings in a date range
type GetBookingsByDateRangeRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// GetBookingsByDateRange retrieves bookings for a date range
// @Summary Get bookings by date range
// @Description Get all bookings within a date range
// @Tags bookings
// @Accept json
// @Produce json
// @Param dates body GetBookingsByDateRangeRequest true "Date range"
// @Success 200 {array} models.Booking
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bookings/by-date [post]
func (h *BookingHandler) GetBookingsByDateRange(c echo.Context) error {
	ctx := c.Request().Context()

	var req GetBookingsByDateRangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	bookings, err := h.bookingService.GetBookingsByDateRange(ctx, req.StartDate, req.EndDate)
	if err != nil {
		log.Errorf("Failed to get bookings: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get bookings: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, bookings)
}
