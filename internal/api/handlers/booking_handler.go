package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// BookingHandler handles HTTP requests related to bookings
type BookingHandler struct {
	bookingService *services.BookingService
}

// NewBookingHandler creates a new BookingHandler
func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// RegisterRoutes registers all booking routes
func (h *BookingHandler) RegisterRoutes(e *echo.Echo) {
	bookings := e.Group("/api/bookings")

	// Regular booking operations
	bookings.GET("", h.SearchAvailableRooms)
	bookings.GET("/:id", h.GetBooking)
	bookings.PUT("/:id", h.UpdateBooking)
	bookings.DELETE("/:id", h.CancelBooking)

	// Payment-first booking creation
	bookings.POST("/with-payment", h.CreateBookingWithPayment)

	// Availability check endpoint
	bookings.GET("/check-availability", h.CheckRoomAvailability)
}

// SearchAvailableRoomsResponse represents the response for SearchAvailableRooms
type SearchAvailableRoomsResponse struct {
	RoomNum  int    `json:"room_num"`
	TypeName string `json:"type_name"`
	Area     int    `json:"area"`
	Price    int    `json:"price_per_night"`
	Capacity int    `json:"capacity"`
}

// SearchAvailableRooms godoc
// @Summary      Search for available rooms
// @Description  Find rooms available for booking in a specific date range and guest count
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        checkIn    query     string  true  "Check-in date (YYYY-MM-DD)"
// @Param        checkOut   query     string  true  "Check-out date (YYYY-MM-DD)"
// @Param        guests     query     integer  true  "Number of guests"
// @Success      200  {array}   models.Room
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /bookings [get]
func (h *BookingHandler) SearchAvailableRooms(c echo.Context) error {
	// Parse query parameters
	checkInStr := c.QueryParam("checkIn")
	checkOutStr := c.QueryParam("checkOut")
	guestsStr := c.QueryParam("guests")

	// Validate and parse the dates
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check-in date format"})
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check-out date format"})
	}

	// Parse guests count
	guests, err := strconv.Atoi(guestsStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid guest count"})
	}

	// Get database transaction from context (or the main DB if no transaction exists)
	tx := middleware.GetTransaction(c)

	// Call service to search for available rooms
	rooms, err := h.bookingService.SearchAvailableRooms(tx, checkIn, checkOut, guests)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to search for available rooms: " + err.Error()})
	}

	return c.JSON(http.StatusOK, rooms)
}

// CheckRoomAvailability godoc
// @Summary      Check room availability
// @Description  Check if a specific room is available for a date range
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        roomNum    query     integer  true  "Room number"
// @Param        checkIn    query     string   true  "Check-in date (YYYY-MM-DD)"
// @Param        checkOut   query     string   true  "Check-out date (YYYY-MM-DD)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bookings/check-availability [get]
func (h *BookingHandler) CheckRoomAvailability(c echo.Context) error {
	// Parse query parameters
	roomNumStr := c.QueryParam("roomNum")
	checkInStr := c.QueryParam("checkIn")
	checkOutStr := c.QueryParam("checkOut")

	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room number"})
	}

	// Validate and parse the dates
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check-in date format"})
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid check-out date format"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Check availability without locking (just a read operation)
	available, err := h.bookingService.CheckRoomAvailability(c.Request().Context(), tx, roomNum, checkIn, checkOut)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check room availability: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"room_num":  roomNum,
		"check_in":  checkInStr,
		"check_out": checkOutStr,
		"available": available,
	})
}

// GetBooking godoc
// @Summary      Get booking details
// @Description  Retrieve details of a specific booking by ID
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Booking ID"
// @Success      200  {object}  models.Booking
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Booking not found"
// @Router       /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to get booking
	booking, err := h.bookingService.GetBooking(tx, bookingID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found: " + err.Error()})
	}

	return c.JSON(http.StatusOK, booking)
}

// UpdateBooking godoc
// @Summary      Update a booking
// @Description  Update an existing booking with transaction and concurrency control
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id       path      integer              true  "Booking ID"
// @Param        booking  body      services.BookingRequest true  "Updated booking details"
// @Success      200      {object}  models.Booking
// @Failure      400      {object}  map[string]string  "Bad request"
// @Router       /bookings/{id} [put]
func (h *BookingHandler) UpdateBooking(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Parse request body
	var req services.BookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to update booking
	booking, err := h.bookingService.UpdateBooking(c.Request().Context(), tx, bookingID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to update booking: " + err.Error()})
	}

	return c.JSON(http.StatusOK, booking)
}

// CancelBooking godoc
// @Summary      Cancel a booking
// @Description  Cancel an existing booking and free up the room
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Booking ID"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Router       /bookings/{id} [delete]
func (h *BookingHandler) CancelBooking(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to cancel booking
	err = h.bookingService.CancelBooking(c.Request().Context(), tx, bookingID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to cancel booking: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Booking cancelled successfully"})
}

// BookingWithPaymentRequest represents a request to create a booking with payment
type BookingWithPaymentRequest struct {
	Booking services.BookingRequest `json:"booking"`
	Payment services.PaymentRequest `json:"payment"`
}

// BookingWithPaymentResponse represents the response for CreateBookingWithPayment
type BookingWithPaymentResponse struct {
	Booking *models.Booking `json:"booking"`
	Receipt *models.Receipt `json:"receipt"`
}

// CreateBookingWithPayment godoc
// @Summary      Create booking with payment
// @Description  Create a new booking with payment in a single atomic transaction. First to pay gets the room.
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        request  body      BookingWithPaymentRequest  true  "Booking and payment details"
// @Success      201      {object}  BookingWithPaymentResponse
// @Failure      400      {object}  map[string]string  "Bad request"
// @Router       /bookings/with-payment [post]
func (h *BookingHandler) CreateBookingWithPayment(c echo.Context) error {
	// Parse request body
	var req BookingWithPaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Basic validation
	if req.Booking.BookingName == "" || req.Booking.RoomNum == 0 ||
		req.Booking.CheckInDate.IsZero() || req.Booking.CheckOutDate.IsZero() ||
		req.Payment.PaymentMethod == "" || req.Payment.PaymentDate.IsZero() {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
	}

	// Call service to create booking with payment
	booking, receipt, err := h.bookingService.CreateBookingWithPayment(
		c.Request().Context(),
		req.Booking,
		req.Payment,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to create booking with payment: " + err.Error()})
	}

	// Return both booking and receipt
	return c.JSON(http.StatusCreated, BookingWithPaymentResponse{
		Booking: booking,
		Receipt: receipt,
	})
}
