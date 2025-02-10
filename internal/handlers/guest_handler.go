package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/dto/response"
	"github.com/panuvitpnv/room-booking-api/internal/service"
)

type GuestHandler struct {
	guestService service.GuestService
}

func NewGuestHandler(guestService service.GuestService) *GuestHandler {
	return &GuestHandler{
		guestService: guestService,
	}
}

// CreateGuest handles new guest registration
// @Summary Create a new guest
// @Description Register a new guest in the system
// @Tags guests
// @Accept json
// @Produce json
// @Param guest body request.CreateGuestRequest true "Guest details"
// @Success 201 {object} response.GuestResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /guests [post]
func (h *GuestHandler) CreateGuest(c echo.Context) error {
	var req request.CreateGuestRequest
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

	guest, err := h.guestService.CreateGuest(c.Request().Context(), &req)
	if err != nil {
		if err == service.ErrDuplicateGuest {
			return c.JSON(http.StatusConflict, response.ErrorResponse{
				Error: "Guest already exists",
				Code:  http.StatusConflict,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create guest",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusCreated, convertToGuestResponse(*guest))
}

// GetGuest retrieves guest details
// @Summary Get guest details
// @Description Get details of a specific guest
// @Tags guests
// @Accept json
// @Produce json
// @Param id path int true "Guest ID"
// @Success 200 {object} response.GuestDetailResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /guests/{id} [get]
func (h *GuestHandler) GetGuest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid guest ID",
			Code:  http.StatusBadRequest,
		})
	}

	guest, err := h.guestService.GetGuestByID(c.Request().Context(), id)
	if err != nil {
		if err == service.ErrGuestNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Guest not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve guest",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.GuestDetailResponse{
		GuestID:     guest.GuestID,
		FirstName:   guest.FirstName,
		LastName:    guest.LastName,
		DateOfBirth: guest.DateOfBirth,
		Email:       guest.Email,
		Phone:       guest.Phone,
	})
}

// UpdateGuest handles guest information updates
// @Summary Update guest information
// @Description Update an existing guest's information
// @Tags guests
// @Accept json
// @Produce json
// @Param id path int true "Guest ID"
// @Param guest body request.UpdateGuestRequest true "Guest details"
// @Success 200 {object} response.GuestResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /guests/{id} [put]
func (h *GuestHandler) UpdateGuest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid guest ID",
			Code:  http.StatusBadRequest,
		})
	}

	var req request.UpdateGuestRequest
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

	guest, err := h.guestService.UpdateGuest(c.Request().Context(), id, &req)
	if err != nil {
		if err == service.ErrGuestNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Guest not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to update guest",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, convertToGuestResponse(*guest))
}

// ListGuests retrieves a list of guests
// @Summary List guests
// @Description Get a paginated list of guests
// @Tags guests
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} response.GuestListResponse
// @Router /guests [get]
func (h *GuestHandler) ListGuests(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	guests, total, err := h.guestService.ListGuests(c.Request().Context(), page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve guests",
			Code:  http.StatusInternalServerError,
		})
	}

	var guestResponses []response.GuestResponse
	for _, guest := range guests {
		guestResponses = append(guestResponses, convertToGuestResponse(guest))
	}

	return c.JSON(http.StatusOK, response.GuestListResponse{
		Guests:   guestResponses,
		Total:    total,
		Page:     page,
		LastPage: (total + pageSize - 1) / pageSize,
	})
}

// GetGuestBookingHistory retrieves booking history for a guest
// @Summary Get guest booking history
// @Description Get the booking history for a specific guest
// @Tags guests
// @Accept json
// @Produce json
// @Param id path int true "Guest ID"
// @Success 200 {object} response.GuestBookingHistoryResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /guests/{id}/bookings [get]
func (h *GuestHandler) GetGuestBookingHistory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid guest ID",
			Code:  http.StatusBadRequest,
		})
	}

	bookings, err := h.guestService.GetGuestBookingHistory(c.Request().Context(), id)
	if err != nil {
		if err == service.ErrGuestNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Guest not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve booking history",
			Code:  http.StatusInternalServerError,
		})
	}

	var bookingResponses []response.BookingResponse
	for _, booking := range bookings {
		bookingResponses = append(bookingResponses, convertToBookingResponse(booking))
	}

	return c.JSON(http.StatusOK, response.GuestBookingHistoryResponse{
		GuestID:  id,
		Bookings: bookingResponses,
	})
}
