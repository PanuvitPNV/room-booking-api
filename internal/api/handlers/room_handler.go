package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
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

	// Query-only endpoints for rooms
	rooms.GET("", h.GetAllRooms)
	rooms.GET("/:id", h.GetRoomByID)
	rooms.GET("/:id/facilities", h.GetRoomWithFacilities)
	rooms.GET("/type/:typeId", h.GetRoomsByType)
	rooms.GET("/:id/calendar", h.GetRoomCalendar)

	// Query-only endpoints for room types
	roomTypes := e.Group("/api/room-types")
	roomTypes.GET("", h.GetAllRoomTypes)
	roomTypes.GET("/:id", h.GetRoomTypeByID)

	// Query-only endpoints for facilities
	facilities := e.Group("/api/facilities")
	facilities.GET("", h.GetAllFacilities)
	facilities.GET("/:id", h.GetFacilityByID)
}

// GetAllRooms godoc
// @Summary      Get all rooms
// @Description  Retrieve all hotel rooms with their types
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Room
// @Failure      500  {object}  map[string]string
// @Router       /rooms [get]
func (h *RoomHandler) GetAllRooms(c echo.Context) error {
	tx := middleware.GetTransaction(c)

	rooms, err := h.roomService.GetAllRooms(tx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve rooms: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, rooms)
}

// GetRoomByID godoc
// @Summary      Get a room by ID
// @Description  Retrieve a specific room by its room number
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Room Number"
// @Success      200  {object}  models.Room
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /rooms/{id} [get]
func (h *RoomHandler) GetRoomByID(c echo.Context) error {
	roomNumStr := c.Param("id")
	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	tx := middleware.GetTransaction(c)

	room, err := h.roomService.GetRoomByID(tx, roomNum)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Room not found: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, room)
}

// GetRoomWithFacilities godoc
// @Summary      Get a room with its facilities
// @Description  Retrieve a specific room with all its facilities
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Room Number"
// @Success      200  {object}  models.Room
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /rooms/{id}/facilities [get]
func (h *RoomHandler) GetRoomWithFacilities(c echo.Context) error {
	roomNumStr := c.Param("id")
	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	tx := middleware.GetTransaction(c)

	room, err := h.roomService.GetRoomWithFacilities(tx, roomNum)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Room not found: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, room)
}

// GetRoomsByType godoc
// @Summary      Get rooms by type
// @Description  Retrieve all rooms of a specific room type
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        typeId   path      integer  true  "Room Type ID"
// @Success      200  {array}   models.Room
// @Failure      500  {object}  map[string]string
// @Router       /rooms/type/{typeId} [get]
func (h *RoomHandler) GetRoomsByType(c echo.Context) error {
	typeIDStr := c.Param("typeId")
	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room type ID",
		})
	}

	tx := middleware.GetTransaction(c)

	rooms, err := h.roomService.GetRoomsByType(tx, typeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve rooms: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, rooms)
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
// @Success      200  {array}   models.RoomType
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

	return c.JSON(http.StatusOK, roomTypes)
}

// GetRoomTypeByID godoc
// @Summary      Get a room type by ID
// @Description  Retrieve a specific room type with its facilities
// @Tags         room-types
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Room Type ID"
// @Success      200  {object}  models.RoomType
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /room-types/{id} [get]
func (h *RoomHandler) GetRoomTypeByID(c echo.Context) error {
	typeIDStr := c.Param("id")
	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room type ID",
		})
	}

	tx := middleware.GetTransaction(c)

	roomType, err := h.roomService.GetRoomTypeByID(tx, typeID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Room type not found: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, roomType)
}

// GetAllFacilities godoc
// @Summary      Get all facilities
// @Description  Retrieve all room facilities
// @Tags         facilities
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Facility
// @Failure      500  {object}  map[string]string
// @Router       /facilities [get]
func (h *RoomHandler) GetAllFacilities(c echo.Context) error {
	tx := middleware.GetTransaction(c)

	facilities, err := h.roomService.GetAllFacilities(tx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve facilities: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, facilities)
}

// GetFacilityByID godoc
// @Summary      Get a facility by ID
// @Description  Retrieve a specific facility
// @Tags         facilities
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Facility ID"
// @Success      200  {object}  models.Facility
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /facilities/{id} [get]
func (h *RoomHandler) GetFacilityByID(c echo.Context) error {
	facIDStr := c.Param("id")
	facID, err := strconv.Atoi(facIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid facility ID",
		})
	}

	tx := middleware.GetTransaction(c)

	facility, err := h.roomService.GetFacilityByID(tx, facID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Facility not found: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, facility)
}
