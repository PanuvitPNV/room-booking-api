package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/dto/response"
	"github.com/panuvitpnv/room-booking-api/internal/service"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBooking handles new booking creation
// @Summary Create a new booking
// @Description Create a new room booking with concurrent handling
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body request.CreateBookingRequest true "Booking details"
// @Success 201 {object} response.BookingResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c echo.Context) error {
	var req request.CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request format",
			Code:  http.StatusBadRequest,
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Validation failed",
			Code:  http.StatusBadRequest,
		})
	}

	booking, err := h.bookingService.CreateBooking(c.Request().Context(), &req)
	if err != nil {
		switch err {
		case service.ErrRoomNotAvailable:
			return c.JSON(http.StatusConflict, response.ErrorResponse{
				Error: "Room not available for selected dates",
				Code:  http.StatusConflict,
			})
		case service.ErrGuestNotFound:
			return c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "Guest not found",
				Code:  http.StatusBadRequest,
			})
		default:
			return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to create booking",
				Code:  http.StatusInternalServerError,
			})
		}
	}

	return c.JSON(http.StatusCreated, response.BookingResponse{
		BookingID:    booking.BookingID,
		RoomNum:      booking.RoomNum,
		GuestID:      booking.GuestID,
		CheckInDate:  booking.CheckInDate,
		CheckOutDate: booking.CheckOutDate,
		TotalPrice:   booking.TotalPrice,
		Room:         convertToRoomResponse(booking.Room),
		Guest:        convertToGuestResponse(booking.Guest),
	})
}

// GetBooking retrieves booking details
// @Summary Get booking details
// @Description Get details of a specific booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} response.BookingDetailResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid booking ID",
			Code:  http.StatusBadRequest,
		})
	}

	booking, err := h.bookingService.GetBookingByID(c.Request().Context(), id)
	if err != nil {
		if err == service.ErrBookingNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Booking not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve booking",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.BookingDetailResponse{
		BookingID:    booking.BookingID,
		RoomNum:      booking.RoomNum,
		GuestID:      booking.GuestID,
		CheckInDate:  booking.CheckInDate,
		CheckOutDate: booking.CheckOutDate,
		TotalPrice:   booking.TotalPrice,
		Room:         convertToRoomResponse(booking.Room),
		Guest:        convertToGuestResponse(booking.Guest),
	})
}

// ListBookings retrieves a list of bookings with filters
// @Summary List bookings
// @Description Get a list of bookings with optional filters
// @Tags bookings
// @Accept json
// @Produce json
// @Param guest_id query int false "Filter by guest ID"
// @Param room_num query int false "Filter by room number"
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} response.BookingListResponse
// @Router /bookings [get]
func (h *BookingHandler) ListBookings(c echo.Context) error {
	// Use the separate function for parsing request params
	req, err := request.ParseGetBookingsRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request parameters",
			Code:  http.StatusBadRequest,
		})
	}

	// Call service
	bookings, total, err := h.bookingService.ListBookings(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve bookings",
			Code:  http.StatusInternalServerError,
		})
	}

	// Convert bookings to response format
	var bookingResponses []response.BookingResponse
	for _, booking := range bookings {
		bookingResponses = append(bookingResponses, response.BookingResponse{
			BookingID:    booking.BookingID,
			RoomNum:      booking.RoomNum,
			GuestID:      booking.GuestID,
			CheckInDate:  booking.CheckInDate,
			CheckOutDate: booking.CheckOutDate,
			TotalPrice:   booking.TotalPrice,
			Room:         convertToRoomResponse(booking.Room),
			Guest:        convertToGuestResponse(booking.Guest),
		})
	}

	// Return paginated response
	return c.JSON(http.StatusOK, response.BookingListResponse{
		Bookings: bookingResponses,
		Total:    total,
		Page:     req.Page,
		LastPage: (total + req.PageSize - 1) / req.PageSize,
	})
}

// CancelBooking handles booking cancellation
// @Summary Cancel booking
// @Description Cancel an existing booking
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid booking ID",
			Code:  http.StatusBadRequest,
		})
	}

	if err := h.bookingService.CancelBooking(c.Request().Context(), id); err != nil {
		if err == service.ErrBookingNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Booking not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to cancel booking",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Booking cancelled successfully",
	})
}

// CheckAvailability checks room availability for given dates
// @Summary Check room availability
// @Description Check if a room is available for specific dates
// @Tags bookings
// @Accept json
// @Produce json
// @Param request body request.CheckAvailabilityRequest true "Availability check details"
// @Success 200 {object} response.BookingAvailabilityResponse
// @Router /bookings/check-availability [post]
func (h *BookingHandler) CheckAvailability(c echo.Context) error {
	var req request.CheckAvailabilityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request format",
			Code:  http.StatusBadRequest,
		})
	}

	available, room, err := h.bookingService.CheckRoomAvailability(
		c.Request().Context(),
		req.RoomNum,
		req.CheckInDate,
		req.CheckOutDate,
	)
	if err != nil {
		if err == service.ErrRoomNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Room not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to check availability",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.BookingAvailabilityResponse{
		RoomNum:       room.RoomNum,
		RoomType:      room.RoomType.Name,
		Status:        getAvailabilityStatus(available),
		PricePerNight: room.RoomType.PricePerNight,
	})
}
