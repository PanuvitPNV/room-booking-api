package repository

import (
	"context"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
)

type RoomTypeRepository interface {
	Create(ctx context.Context, roomType *models.RoomType) error
	GetByID(ctx context.Context, typeID int) (*models.RoomType, error)
	Update(ctx context.Context, roomType *models.RoomType) error
	Delete(ctx context.Context, typeID int) error
	List(ctx context.Context) ([]models.RoomType, error)
}

type RoomRepository interface {
	Create(ctx context.Context, room *models.Room) error
	GetByNum(ctx context.Context, roomNum int) (*models.Room, error)
	Update(ctx context.Context, room *models.Room) error
	Delete(ctx context.Context, roomNum int) error
	List(ctx context.Context) ([]models.Room, error)
	GetAvailableRooms(ctx context.Context, checkIn, checkOut time.Time) ([]models.Room, error)
}

type GuestRepository interface {
	Create(ctx context.Context, guest *models.Guest) error
	GetByID(ctx context.Context, guestID int) (*models.Guest, error)
	Update(ctx context.Context, guest *models.Guest) error
	Delete(ctx context.Context, guestID int) error
	List(ctx context.Context) ([]models.Guest, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) error
	GetByID(ctx context.Context, bookingID int) (*models.Booking, error)
	Update(ctx context.Context, booking *models.Booking) error
	Delete(ctx context.Context, bookingID int) error
	List(ctx context.Context) ([]models.Booking, error)
	GetGuestBookings(ctx context.Context, guestID int) ([]models.Booking, error)
}

type RoomStatusRepository interface {
	Create(ctx context.Context, status *models.RoomStatus) error
	GetByRoomAndDate(ctx context.Context, roomNum int, date time.Time) (*models.RoomStatus, error)
	Update(ctx context.Context, status *models.RoomStatus) error
	List(ctx context.Context, date time.Time) ([]models.RoomStatus, error)
}
