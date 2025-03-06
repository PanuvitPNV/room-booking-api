package routes

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/api/handlers"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// TemplateRenderer is a custom HTML template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global template functions if needed
	return t.templates.ExecuteTemplate(w, name, data)
}

// SetupRoutes configures all application routes
func SetupRoutes(
	e *echo.Echo,
	bookingHandler *handlers.BookingHandler,
	receiptHandler *handlers.ReceiptHandler,
	roomHandler *handlers.RoomHandler,
) {
	// Create the templates directory structure if it doesn't exist
	// You can add code here to ensure directories exist

	// Initialize the template renderer
	templateFiles := "./web/templates/*.html"
	templates, err := template.ParseGlob(templateFiles)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		panic(fmt.Sprintf("Failed to parse templates: %v", err))
	}

	e.Renderer = &TemplateRenderer{
		templates: templates,
	}

	// Serve static files for assets
	e.Static("/static", "./web/static")

	// Home route for the booking system
	e.GET("/", func(c echo.Context) error {
		err := c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Title": "Hotel Booking System",
		})

		if err != nil {
			log.Printf("Error rendering template: %v", err)
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Error rendering template: %v", err))
		}

		return nil
	})

	// Swagger documentation route
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Create a versioned API group
	api := e.Group("/api/v1")

	// Add transaction tracking middleware to specific routes
	txMiddleware := middleware.TransactionTracker()

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
		rooms.POST("/available", bookingHandler.GetAvailableRooms)
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
}
