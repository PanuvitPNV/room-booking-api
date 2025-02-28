package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// ConcurrentHandler handles HTTP requests related to concurrency demonstrations
type ConcurrentHandler struct {
	concurrentService *services.ConcurrentService
}

// NewConcurrentHandler creates a new ConcurrentHandler
func NewConcurrentHandler(concurrentService *services.ConcurrentService) *ConcurrentHandler {
	return &ConcurrentHandler{
		concurrentService: concurrentService,
	}
}

// RegisterRoutes registers all concurrent scenario routes
func (h *ConcurrentHandler) RegisterRoutes(e *echo.Echo) {
	demos := e.Group("/api/demos")

	demos.GET("/lost-update/:id", h.DemoLostUpdate)
	demos.GET("/lost-update-with-locking/:id", h.DemoLostUpdateWithLocking)
	demos.GET("/dirty-read/:id", h.DemoDirtyRead)
	demos.GET("/phantom-read", h.DemoPhantomRead)
	demos.GET("/serialization-anomaly", h.DemoSerializationAnomaly)
	demos.GET("/concurrent-bookings", h.DemoConcurrentBookings)
}

// DemoResponse represents a standard response for demonstration endpoints
type DemoResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Result      string `json:"result"`
}

// DemoLostUpdate godoc
// @Summary      Lost Update Demo
// @Description  Demonstrates the lost update problem in concurrent transactions
// @Tags         demos
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Booking ID"
// @Success      200  {object}  DemoResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /demos/lost-update/{id} [get]
func (h *ConcurrentHandler) DemoLostUpdate(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Run the lost update demonstration
	result, err := h.concurrentService.DemoLostUpdate(c.Request().Context(), bookingID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to run demonstration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DemoResponse{
		Title:       "Lost Update Problem Demonstration",
		Description: "This demo shows how concurrent transactions can lead to lost updates when two transactions read and update the same data without proper concurrency control.",
		Result:      result,
	})
}

// DemoLostUpdateWithLocking godoc
// @Summary      Lost Update Prevention Demo
// @Description  Demonstrates how pessimistic locking prevents lost updates
// @Tags         demos
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Booking ID"
// @Success      200  {object}  DemoResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /demos/lost-update-with-locking/{id} [get]
func (h *ConcurrentHandler) DemoLostUpdateWithLocking(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Run the lost update with locking demonstration
	result, err := h.concurrentService.DemoLostUpdateWithLocking(c.Request().Context(), bookingID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to run demonstration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DemoResponse{
		Title:       "Lost Update Prevention with Pessimistic Locking",
		Description: "This demo shows how using row-level locks (SELECT FOR UPDATE) prevents lost updates by forcing the second transaction to wait until the first one completes.",
		Result:      result,
	})
}

// DemoDirtyRead godoc
// @Summary      Dirty Read Demo
// @Description  Demonstrates the dirty read problem in database transactions
// @Tags         demos
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Booking ID"
// @Success      200  {object}  DemoResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /demos/dirty-read/{id} [get]
func (h *ConcurrentHandler) DemoDirtyRead(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Run the dirty read demonstration
	result, err := h.concurrentService.DemoDirtyRead(c.Request().Context(), bookingID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to run demonstration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DemoResponse{
		Title:       "Dirty Read Problem Demonstration",
		Description: "This demo shows how one transaction can read uncommitted changes made by another transaction, which might later be rolled back, leading to inconsistent data views.",
		Result:      result,
	})
}

// DemoPhantomRead godoc
// @Summary      Phantom Read Demo
// @Description  Demonstrates the phantom read problem in database transactions
// @Tags         demos
// @Accept       json
// @Produce      json
// @Param        checkIn   query     string  true  "Check-in date (YYYY-MM-DD)"
// @Param        checkOut  query     string  true  "Check-out date (YYYY-MM-DD)"
// @Success      200  {object}  DemoResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /demos/phantom-read [get]
func (h *ConcurrentHandler) DemoPhantomRead(c echo.Context) error {
	// Parse query parameters
	checkInStr := c.QueryParam("checkIn")
	checkOutStr := c.QueryParam("checkOut")

	if checkInStr == "" || checkOutStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Both checkIn and checkOut query parameters are required (format: YYYY-MM-DD)",
		})
	}

	// Run the phantom read demonstration
	result, err := h.concurrentService.DemoPhantomRead(c.Request().Context(), checkInStr, checkOutStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to run demonstration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DemoResponse{
		Title:       "Phantom Read Problem Demonstration",
		Description: "This demo shows how one transaction can get different results from the same query when another transaction adds or removes rows that match the query conditions.",
		Result:      result,
	})
}

// DemoSerializationAnomaly godoc
// @Summary      Serialization Anomaly Demo
// @Description  Demonstrates serialization anomaly in database transactions
// @Tags         demos
// @Accept       json
// @Produce      json
// @Success      200  {object}  DemoResponse
// @Failure      500  {object}  map[string]string
// @Router       /demos/serialization-anomaly [get]
func (h *ConcurrentHandler) DemoSerializationAnomaly(c echo.Context) error {
	// Run the serialization anomaly demonstration
	result, err := h.concurrentService.DemoSerializationAnomaly(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to run demonstration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DemoResponse{
		Title:       "Serialization Anomaly Demonstration",
		Description: "This demo shows a situation where the results of a set of transactions executed concurrently might not be the same as if they were executed sequentially in some order.",
		Result:      result,
	})
}

// DemoConcurrentBookings godoc
// @Summary      Concurrent Booking Demo
// @Description  Demonstrates how pessimistic locking prevents double booking
// @Tags         demos
// @Accept       json
// @Produce      json
// @Param        roomNum    query     integer  true  "Room number"
// @Param        checkIn    query     string   true  "Check-in date (YYYY-MM-DD)"
// @Param        checkOut   query     string   true  "Check-out date (YYYY-MM-DD)"
// @Success      200  {object}  DemoResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /demos/concurrent-bookings [get]
func (h *ConcurrentHandler) DemoConcurrentBookings(c echo.Context) error {
	// Parse query parameters
	roomNumStr := c.QueryParam("roomNum")
	checkInStr := c.QueryParam("checkIn")
	checkOutStr := c.QueryParam("checkOut")

	if roomNumStr == "" || checkInStr == "" || checkOutStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "roomNum, checkIn, and checkOut query parameters are required",
		})
	}

	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room number"})
	}

	// Run the concurrent bookings demonstration
	result, err := h.concurrentService.DemoConcurrentBookings(c.Request().Context(), roomNum, checkInStr, checkOutStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to run demonstration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, DemoResponse{
		Title:       "Concurrent Booking Prevention Demonstration",
		Description: "This demo shows how pessimistic locking prevents double booking of the same room for overlapping dates when concurrent booking attempts are made.",
		Result:      result,
	})
}
