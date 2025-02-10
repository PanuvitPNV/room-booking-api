package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/dto/response"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/service"
)

type RoomHandler struct {
	roomService       service.RoomService
	roomTypeService   service.RoomTypeService
	roomStatusService service.RoomStatusService
}

func NewRoomHandler(
	roomService service.RoomService,
	roomTypeService service.RoomTypeService,
	roomStatusService service.RoomStatusService,
) *RoomHandler {
	return &RoomHandler{
		roomService:       roomService,
		roomTypeService:   roomTypeService,
		roomStatusService: roomStatusService,
	}
}

// CreateRoom handles room creation
// @Summary Create a new room
// @Description Create a new room with specified type
// @Tags rooms
// @Accept json
// @Produce json
// @Param room body request.CreateRoomRequest true "Room details"
// @Success 201 {object} response.RoomResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /rooms [post]
func (h *RoomHandler) CreateRoom(c echo.Context) error {
	var req request.CreateRoomRequest
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

	room, err := h.roomService.CreateRoom(c.Request().Context(), &req)
	if err != nil {
		switch err {
		case service.ErrRoomTypeNotFound:
			return c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "Room type not found",
				Code:  http.StatusBadRequest,
			})
		case service.ErrDuplicateRoom:
			return c.JSON(http.StatusConflict, response.ErrorResponse{
				Error: "Room number already exists",
				Code:  http.StatusConflict,
			})
		default:
			return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to create room",
				Code:  http.StatusInternalServerError,
			})
		}
	}

	return c.JSON(http.StatusCreated, convertToRoomResponse(*room))
}

// GetRoom retrieves room details
// @Summary Get room details
// @Description Get details of a specific room
// @Tags rooms
// @Accept json
// @Produce json
// @Param room_num path int true "Room Number"
// @Success 200 {object} response.RoomDetailResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /rooms/{room_num} [get]
func (h *RoomHandler) GetRoom(c echo.Context) error {
	roomNum, err := strconv.Atoi(c.Param("room_num"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid room number",
			Code:  http.StatusBadRequest,
		})
	}

	room, err := h.roomService.GetRoomByNum(c.Request().Context(), roomNum)
	if err != nil {
		if err == service.ErrRoomNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Room not found",
				Code:  http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve room",
			Code:  http.StatusInternalServerError,
		})
	}

	// Get current room status
	status, err := h.roomStatusService.GetRoomStatus(c.Request().Context(), roomNum, time.Now())
	if err != nil && err != service.ErrRoomStatusNotFound {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve room status",
			Code:  http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.RoomDetailResponse{
		RoomNum:  room.RoomNum,
		TypeID:   room.TypeID,
		RoomType: convertToRoomTypeResponse(room.RoomType),
		CurrentStatus: response.RoomStatusInfo{
			Status:    getStatusString(status),
			BookingID: status.BookingID,
		},
	})
}

// ListRooms retrieves a list of rooms with filters
// @Summary List rooms
// @Description Get a list of rooms with optional type filter
// @Tags rooms
// @Accept json
// @Produce json
// @Param type_id query int false "Filter by room type"
// @Param page query int true "Page number"
// @Param page_size query int true "Items per page"
// @Success 200 {object} response.RoomListResponse
// @Router /rooms [get]
func (h *RoomHandler) ListRooms(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	typeID, _ := strconv.Atoi(c.QueryParam("type_id"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var rooms []models.Room
	var total int
	var err error

	if typeID > 0 {
		rooms, err = h.roomService.GetRoomsByType(c.Request().Context(), typeID)
		total = len(rooms)
	} else {
		rooms, total, err = h.roomService.ListRooms(c.Request().Context(), page, pageSize)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to retrieve rooms",
			Code:  http.StatusInternalServerError,
		})
	}

	var roomResponses []response.RoomResponse
	for _, room := range rooms {
		roomResponses = append(roomResponses, convertToRoomResponse(room))
	}

	return c.JSON(http.StatusOK, response.RoomListResponse{
		Rooms:    roomResponses,
		Total:    total,
		Page:     page,
		LastPage: (total + pageSize - 1) / pageSize,
	})
}

// GetAvailableRooms retrieves available rooms for given dates
// @Summary List available rooms
// @Description Get a list of available rooms for specific dates
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body request.GetAvailableRoomsRequest true "Availability request"
// @Success 200 {array} response.RoomAvailabilityResponse
// @Router /rooms/available [post]
func (h *RoomHandler) GetAvailableRooms(c echo.Context) error {
	var req request.GetAvailableRoomsRequest
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

	rooms, err := h.roomService.GetAvailableRooms(
		c.Request().Context(),
		req.CheckInDate,
		req.CheckOutDate,
		req.TypeID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to check room availability",
			Code:  http.StatusInternalServerError,
		})
	}

	var availableRooms []response.RoomAvailabilityResponse
	for _, room := range rooms {
		availableRooms = append(availableRooms, response.RoomAvailabilityResponse{
			RoomNum:       room.RoomNum,
			RoomType:      room.RoomType.Name,
			Status:        "Available",
			PricePerNight: room.RoomType.PricePerNight,
		})
	}

	return c.JSON(http.StatusOK, availableRooms)
}
