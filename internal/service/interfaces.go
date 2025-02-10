package service

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/models"
)

// RoomTypeService handles business logic for room types
type RoomTypeService interface {
	CreateRoomType(ctx context.Context, req *request.CreateRoomTypeRequest) (*models.RoomType, error)
	GetRoomTypeByID(ctx context.Context, typeID int) (*models.RoomType, error)
	UpdateRoomType(ctx context.Context, typeID int, req *request.UpdateRoomTypeRequest) (*models.RoomType, error)
	DeleteRoomType(ctx context.Context, typeID int) error
	ListRoomTypes(ctx context.Context, page, pageSize int) ([]models.RoomType, int, error)
}

// RoomService handles business logic for rooms
type RoomService interface {
	CreateRoom(ctx context.Context, req *request.CreateRoomRequest) (*models.Room, error)
	GetRoomByNum(ctx context.Context, roomNum int) (*models.Room, error)
	UpdateRoom(ctx context.Context, roomNum int, req *request.UpdateRoomRequest) (*models.Room, error)
	DeleteRoom(ctx context.Context, roomNum int) error
	ListRooms(ctx context.Context, page, pageSize int) ([]models.Room, int, error)
	GetRoomsByType(ctx context.Context, typeID int) ([]models.Room, error)
	GetAvailableRooms(ctx context.Context, checkIn, checkOut time.Time, typeID *int) ([]models.Room, error)
}

// GuestService handles business logic for guests
type GuestService interface {
	CreateGuest(ctx context.Context, req *request.CreateGuestRequest) (*models.Guest, error)
	GetGuestByID(ctx context.Context, guestID int) (*models.Guest, error)
	UpdateGuest(ctx context.Context, guestID int, req *request.UpdateGuestRequest) (*models.Guest, error)
	DeleteGuest(ctx context.Context, guestID int) error
	ListGuests(ctx context.Context, page, pageSize int) ([]models.Guest, int, error)
	GetGuestBookingHistory(ctx context.Context, guestID int) ([]models.Booking, error)
}

// BookingService handles business logic for bookings with transaction support
type BookingService interface {
	CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*models.Booking, error)
	GetBookingByID(ctx context.Context, bookingID int) (*models.Booking, error)
	UpdateBooking(ctx context.Context, bookingID int, req *request.UpdateBookingRequest) (*models.Booking, error)
	CancelBooking(ctx context.Context, bookingID int) error
	ListBookings(ctx context.Context, req *request.GetBookingsRequest) ([]models.Booking, int, error)
	CheckRoomAvailability(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, *models.Room, error)
	GetBookingsByGuest(ctx context.Context, guestID int) ([]models.Booking, error)
	GetBookingsByRoom(ctx context.Context, roomNum int) ([]models.Booking, error)
}

// RoomStatusService handles business logic for room status
type RoomStatusService interface {
	GetRoomStatus(ctx context.Context, roomNum int, date time.Time) (*models.RoomStatus, error)
	UpdateRoomStatus(ctx context.Context, status *models.RoomStatus) error
	GetRoomStatusRange(ctx context.Context, roomNum int, startDate, endDate time.Time) ([]models.RoomStatus, error)
	GetAllRoomStatus(ctx context.Context, date time.Time) ([]models.RoomStatus, error)
}

// TransactionOptions defines options for transaction management
type TransactionOptions struct {
	IsolationLevel string
	ReadOnly       bool
	Timeout        time.Duration
}

// TransactionManager handles transaction operations
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	WithTransactionOptions(ctx context.Context, opts *TransactionOptions, fn func(ctx context.Context) error) error
}

// Common errors
var (
	ErrRoomNotAvailable  = errors.New("room not available for the selected dates")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrGuestNotFound     = errors.New("guest not found")
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomTypeNotFound  = errors.New("room type not found")
	ErrDuplicateRoom     = errors.New("room already exists")
	ErrDuplicateGuest    = errors.New("guest already exists")
	ErrInvalidData       = errors.New("invalid data")
	ErrTransactionFailed = errors.New("transaction failed")
)

// Status constants
const (
	StatusAvailable = "Available"
	StatusOccupied  = "Occupied"
)
