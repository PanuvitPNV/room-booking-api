package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/handlers"
)

type Router struct {
	bookingHandler *handlers.BookingHandler
	roomHandler    *handlers.RoomHandler
	guestHandler   *handlers.GuestHandler
}

func NewRouter(
	bookingHandler *handlers.BookingHandler,
	roomHandler *handlers.RoomHandler,
	guestHandler *handlers.GuestHandler,
) *Router {
	return &Router{
		bookingHandler: bookingHandler,
		roomHandler:    roomHandler,
		guestHandler:   guestHandler,
	}
}

func (r *Router) SetupRoutes(e *echo.Echo) {
	v1 := e.Group("/v1")

	// Booking routes
	bookings := v1.Group("/bookings")
	{
		bookings.POST("", r.bookingHandler.CreateBooking)
		bookings.GET("", r.bookingHandler.ListBookings)
		bookings.GET("/:id", r.bookingHandler.GetBooking)
		bookings.POST("/check-availability", r.bookingHandler.CheckAvailability)
		bookings.POST("/:id/cancel", r.bookingHandler.CancelBooking)
	}

	// Room routes
	rooms := v1.Group("/rooms")
	{
		rooms.POST("", r.roomHandler.CreateRoom)
		rooms.GET("", r.roomHandler.ListRooms)
		rooms.GET("/:room_num", r.roomHandler.GetRoom)
		rooms.POST("/available", r.roomHandler.GetAvailableRooms)

		// Add room type routes
		rooms.POST("/types", r.roomHandler.CreateRoomType)    // Add this
		rooms.GET("/types", r.roomHandler.ListRoomTypes)      // Add this
		rooms.GET("/types/:id", r.roomHandler.GetRoomType)    // Add this
		rooms.PUT("/types/:id", r.roomHandler.UpdateRoomType) // Add this
	}

	// Guest routes
	guests := v1.Group("/guests")
	{
		guests.POST("", r.guestHandler.CreateGuest)
		guests.GET("", r.guestHandler.ListGuests)
		guests.GET("/:id", r.guestHandler.GetGuest)
		guests.PUT("/:id", r.guestHandler.UpdateGuest)
		guests.GET("/:id/bookings", r.guestHandler.GetGuestBookingHistory)
	}
}
