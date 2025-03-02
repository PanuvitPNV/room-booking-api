package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/panuvitpnv/room-booking-api/internal/api/handlers"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	e *echo.Echo,
	bookingHandler *handlers.BookingHandler,
	receiptHandler *handlers.ReceiptHandler,
	roomHandler *handlers.RoomHandler,
) {
	// Create a versioned API group
	api := e.Group("/api/v1")

	// Add transaction tracking middleware to specific routes
	txMiddleware := middleware.TransactionTracker()

	// Monitoring/health check routes
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "UP",
		})
	})

	// Booking routes
	bookings := api.Group("/bookings")
	{
		bookings.POST("", bookingHandler.CreateBooking, txMiddleware)
		bookings.GET("/:id", bookingHandler.GetBooking)
		bookings.DELETE("/:id", bookingHandler.CancelBooking, txMiddleware)
		bookings.PUT("/:id", bookingHandler.UpdateBooking, txMiddleware)
		bookings.POST("/by-date", bookingHandler.GetBookingsByDateRange)
	}

	// Room routes
	rooms := api.Group("/rooms")
	{
		rooms.GET("", roomHandler.GetAllRooms)
		rooms.GET("/:roomNum", roomHandler.GetRoomByNumber)
		rooms.POST("/:roomNum/status", roomHandler.GetRoomStatus)
		rooms.GET("/types", roomHandler.GetRoomTypes)
		rooms.GET("/type/:typeId", roomHandler.GetRoomsByType)
		rooms.POST("/availability", roomHandler.GetRoomAvailabilitySummary)
		rooms.POST("/available", bookingHandler.GetAvailableRooms) // For booking availability
	}

	// Receipt routes
	receipts := api.Group("/receipts")
	{
		receipts.POST("", receiptHandler.CreateReceipt, txMiddleware)
		receipts.GET("", receiptHandler.GetAllReceipts)
		receipts.GET("/:id", receiptHandler.GetReceipt)
		receipts.GET("/booking/:bookingId", receiptHandler.GetReceiptByBooking)
		receipts.POST("/refund", receiptHandler.ProcessRefund, txMiddleware)
		receipts.POST("/by-date", receiptHandler.GetReceiptsByDateRange)
	}

	// Add docs route if we implement Swagger later
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
