package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
)

// Transaction Manager interface
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

// Base repository with transaction support
type BaseRepository struct {
	db *gorm.DB
}

type contextKey string

const txKey contextKey = "tx"

func (r *BaseRepository) WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

func (r *BaseRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

// RoomType Repository interface
type RoomTypeRepository interface {
	TxManager
	Create(ctx context.Context, roomType *models.RoomType) error
	GetByID(ctx context.Context, typeID int) (*models.RoomType, error)
	Update(ctx context.Context, roomType *models.RoomType) error
	Delete(ctx context.Context, typeID int) error
	List(ctx context.Context) ([]models.RoomType, error)
}

// Room Repository interface
type RoomRepository interface {
	TxManager
	Create(ctx context.Context, room *models.Room) error
	GetByNum(ctx context.Context, roomNum int) (*models.Room, error)
	Update(ctx context.Context, room *models.Room) error
	Delete(ctx context.Context, roomNum int) error
	List(ctx context.Context) ([]models.Room, error)
	GetByType(ctx context.Context, typeID int) ([]models.Room, error)
	GetAvailableRooms(ctx context.Context, checkIn, checkOut time.Time) ([]models.Room, error)
	GetAvailableRoomsByType(ctx context.Context, typeID int, checkIn, checkOut time.Time) ([]models.Room, error)
}

// Guest Repository interface
type GuestRepository interface {
	TxManager
	Create(ctx context.Context, guest *models.Guest) error
	GetByID(ctx context.Context, guestID int) (*models.Guest, error)
	Update(ctx context.Context, guest *models.Guest) error
	Delete(ctx context.Context, guestID int) error
	List(ctx context.Context) ([]models.Guest, error)
}

// Booking Repository interface
type BookingRepository interface {
	TxManager
	Create(ctx context.Context, booking *models.Booking) error
	GetByID(ctx context.Context, bookingID int) (*models.Booking, error)
	Update(ctx context.Context, booking *models.Booking) error
	Delete(ctx context.Context, bookingID int) error
	List(ctx context.Context) ([]models.Booking, error)
	GetGuestBookings(ctx context.Context, guestID int) ([]models.Booking, error)
	GetRoomBookings(ctx context.Context, roomNum int) ([]models.Booking, error)
	CheckRoomAvailability(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, error)
}

// RoomStatus Repository interface
type RoomStatusRepository interface {
	TxManager
	Create(ctx context.Context, status *models.RoomStatus) error
	GetByRoomAndDate(ctx context.Context, roomNum int, date time.Time) (*models.RoomStatus, error)
	Update(ctx context.Context, status *models.RoomStatus) error
	List(ctx context.Context, date time.Time) ([]models.RoomStatus, error)
	GetRoomStatusRange(ctx context.Context, roomNum int, startDate, endDate time.Time) ([]models.RoomStatus, error)
}

// Error definitions
var (
	ErrRecordNotFound    = errors.New("record not found")
	ErrDuplicateEntry    = errors.New("duplicate entry")
	ErrTransactionFailed = errors.New("transaction failed")
	ErrInvalidData       = errors.New("invalid data")
)

// Transaction options type
type TxOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

// Extended transaction manager with options
type TxManagerWithOptions interface {
	TxManager
	WithinTransactionWithOptions(ctx context.Context, opts *TxOptions, fn func(txCtx context.Context) error) error
}

// Add options to base repository
func (r *BaseRepository) WithinTransactionWithOptions(ctx context.Context, opts *TxOptions, fn func(txCtx context.Context) error) error {
	txOpts := &sql.TxOptions{
		Isolation: opts.Isolation,
		ReadOnly:  opts.ReadOnly,
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	}, txOpts)
}
