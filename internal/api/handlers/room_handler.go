package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// RoomHandler handles HTTP requests related to rooms
type RoomHandler struct {
	roomService *services.RoomService
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomService *services.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// GetAllRooms retrieves all rooms
// @Summary Get all rooms
// @Description Get all rooms with their types
// @Tags rooms
// @Produce json
// @Success 200 {array} models.Room
// @Failure 500 {object} map[string]string
// @Router /rooms [get]
func (h *RoomHandler) GetAllRooms(c echo.Context) error {
	ctx := c.Request().Context()

	rooms, err := h.roomService.GetAllRooms(ctx)
	if err != nil {
		log.Errorf("Failed to get rooms: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get rooms: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, rooms)
}

// GetRoomByNumber retrieves a room by number
// @Summary Get a room by number
// @Description Get a room by its room number
// @Tags rooms
// @Produce json
// @Param roomNum path int true "Room Number"
// @Success 200 {object} models.Room
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rooms/{roomNum} [get]
func (h *RoomHandler) GetRoomByNumber(c echo.Context) error {
	ctx := c.Request().Context()

	roomNum, err := strconv.Atoi(c.Param("roomNum"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	room, err := h.roomService.GetRoomByNumber(ctx, roomNum)
	if err != nil {
		log.Errorf("Failed to get room: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Room not found",
		})
	}

	return c.JSON(http.StatusOK, room)
}

// GetRoomStatusRequest represents a request to get room status for a date range
type GetRoomStatusRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// GetRoomStatus retrieves room status for a date range
// @Summary Get room status
// @Description Get room status for a specific room and date range
// @Tags rooms
// @Accept json
// @Produce json
// @Param roomNum path int true "Room Number"
// @Param dates body GetRoomStatusRequest true "Date range"
// @Success 200 {array} models.RoomStatus
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rooms/{roomNum}/status [post]
func (h *RoomHandler) GetRoomStatus(c echo.Context) error {
	ctx := c.Request().Context()

	roomNum, err := strconv.Atoi(c.Param("roomNum"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	var req GetRoomStatusRequest
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

	statuses, err := h.roomService.GetRoomStatusForDateRange(ctx, roomNum, req.StartDate, req.EndDate)
	if err != nil {
		log.Errorf("Failed to get room status: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get room status: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, statuses)
}

// GetRoomTypes retrieves all room types
// @Summary Get all room types
// @Description Get all available room types
// @Tags rooms
// @Produce json
// @Success 200 {array} models.RoomType
// @Failure 500 {object} map[string]string
// @Router /rooms/types [get]
func (h *RoomHandler) GetRoomTypes(c echo.Context) error {
	ctx := c.Request().Context()

	roomTypes, err := h.roomService.GetRoomTypes(ctx)
	if err != nil {
		log.Errorf("Failed to get room types: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get room types: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, roomTypes)
}

// GetRoomsByType retrieves rooms by type
// @Summary Get rooms by type
// @Description Get all rooms of a specific type
// @Tags rooms
// @Produce json
// @Param typeId path int true "Room Type ID"
// @Success 200 {array} models.Room
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rooms/type/{typeId} [get]
func (h *RoomHandler) GetRoomsByType(c echo.Context) error {
	ctx := c.Request().Context()

	typeID, err := strconv.Atoi(c.Param("typeId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room type ID",
		})
	}

	rooms, err := h.roomService.GetRoomsByType(ctx, typeID)
	if err != nil {
		log.Errorf("Failed to get rooms: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get rooms: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, rooms)
}

// GetRoomAvailabilityRequest represents a request to get room availability summary
type GetRoomAvailabilityRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// GetRoomAvailabilitySummary retrieves room availability summary
// @Summary Get room availability summary
// @Description Get availability summary for all rooms in a date range
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body GetRoomAvailabilityRequest true "Date range"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rooms/availability [post]
func (h *RoomHandler) GetRoomAvailabilitySummary(c echo.Context) error {
	ctx := c.Request().Context()

	var req GetRoomAvailabilityRequest
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

	summary, err := h.roomService.GetRoomAvailabilitySummary(ctx, req.StartDate, req.EndDate)
	if err != nil {
		log.Errorf("Failed to get room availability: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get room availability: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"summary": summary,
		"period": map[string]string{
			"start_date": req.StartDate.Format("2006-01-02"),
			"end_date":   req.EndDate.Format("2006-01-02"),
		},
	})
}
