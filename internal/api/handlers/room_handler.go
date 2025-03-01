package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	responseModels "github.com/panuvitpnv/room-booking-api/internal/api/models"
	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// RoomHandler handles HTTP requests related to rooms
type RoomHandler struct {
	roomService *services.RoomService
}

// NewRoomHandler creates a new RoomHandler
func NewRoomHandler(roomService *services.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// RegisterRoutes registers all room routes
func (h *RoomHandler) RegisterRoutes(e *echo.Echo) {
	rooms := e.Group("/api/rooms")

	// Simplified endpoints for UI
	rooms.GET("", h.GetAllRoomsWithDetails)
	rooms.GET("/:id", h.GetRoomWithDetails)
	rooms.GET("/type/:typeId", h.GetRoomsByTypeWithDetails)
	rooms.GET("/:id/calendar", h.GetRoomCalendar)

	roomTypes := e.Group("/api/room-types")
	roomTypes.GET("", h.GetAllRoomTypes)
}

// GetAllRoomsWithDetails godoc
// @Summary      Get all rooms with details
// @Description  Retrieve all hotel rooms with their types and facilities
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.RoomResponse
// @Failure      500  {object}  map[string]string
// @Router       /rooms [get]
func (h *RoomHandler) GetAllRoomsWithDetails(c echo.Context) error {
	tx := middleware.GetTransaction(c)

	rooms, err := h.roomService.GetAllRoomsWithDetails(tx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve rooms: " + err.Error(),
		})
	}

	// Convert to response model
	response := responseModels.ConvertRoomsToResponse(rooms)

	return c.JSON(http.StatusOK, response)
}

// GetRoomWithDetails godoc
// @Summary      Get a room with details
// @Description  Retrieve a specific room with its type and facilities
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Room Number"
// @Success      200  {object}  models.RoomResponse
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /rooms/{id} [get]
func (h *RoomHandler) GetRoomWithDetails(c echo.Context) error {
	roomNumStr := c.Param("id")
	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	tx := middleware.GetTransaction(c)

	room, err := h.roomService.GetRoomWithDetails(tx, roomNum)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Room not found: " + err.Error(),
		})
	}

	// Convert to response model
	response := responseModels.ConvertToRoomResponse(*room)

	return c.JSON(http.StatusOK, response)
}

// GetRoomsByTypeWithDetails godoc
// @Summary      Get rooms by type with details
// @Description  Retrieve all rooms of a specific room type with facilities
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        typeId   path      integer  true  "Room Type ID"
// @Success      200  {array}   models.RoomResponse
// @Failure      500  {object}  map[string]string
// @Router       /rooms/type/{typeId} [get]
func (h *RoomHandler) GetRoomsByTypeWithDetails(c echo.Context) error {
	typeIDStr := c.Param("typeId")
	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room type ID",
		})
	}

	tx := middleware.GetTransaction(c)

	rooms, err := h.roomService.GetRoomsByTypeWithDetails(tx, typeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve rooms: " + err.Error(),
		})
	}

	// Convert to response model
	response := responseModels.ConvertRoomsToResponse(rooms)

	return c.JSON(http.StatusOK, response)
}

// GetRoomCalendar godoc
// @Summary      Get room calendar
// @Description  Retrieve the availability calendar for a specific room
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        id         path      integer  true  "Room Number"
// @Param        startDate  query     string   true  "Start Date (YYYY-MM-DD)"
// @Param        endDate    query     string   true  "End Date (YYYY-MM-DD)"
// @Success      200  {array}   models.RoomStatus
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /rooms/{id}/calendar [get]
func (h *RoomHandler) GetRoomCalendar(c echo.Context) error {
	roomNumStr := c.Param("id")
	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	startDate := c.QueryParam("startDate")
	endDate := c.QueryParam("endDate")

	if startDate == "" || endDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "startDate and endDate query parameters are required",
		})
	}

	tx := middleware.GetTransaction(c)

	statuses, err := h.roomService.GetRoomCalendar(tx, roomNum, startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to retrieve room calendar: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, statuses)
}

// GetAllRoomTypes godoc
// @Summary      Get all room types
// @Description  Retrieve all room types with their facilities
// @Tags         room-types
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.RoomTypeResponse
// @Failure      500  {object}  map[string]string
// @Router       /room-types [get]
func (h *RoomHandler) GetAllRoomTypes(c echo.Context) error {
	tx := middleware.GetTransaction(c)

	roomTypes, err := h.roomService.GetAllRoomTypes(tx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve room types: " + err.Error(),
		})
	}

	// Convert to response model
	response := responseModels.ConvertRoomTypesToResponse(roomTypes)

	return c.JSON(http.StatusOK, response)
}
