package services

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"github.com/panuvitpnv/room-booking-api/internal/utils"
)

// RoomService handles room-related business logic
type RoomService struct {
	roomRepo    *repositories.RoomRepository
	lockManager *utils.LockManager
}

// NewRoomService creates a new room service
func NewRoomService(
	roomRepo *repositories.RoomRepository,
	lockManager *utils.LockManager,
) *RoomService {
	return &RoomService{
		roomRepo:    roomRepo,
		lockManager: lockManager,
	}
}

// GetAllRooms retrieves all rooms
func (s *RoomService) GetAllRooms(ctx context.Context) ([]models.Room, error) {
	var rooms []models.Room
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		rooms, err = s.roomRepo.GetAllRooms(tx)
		return err
	})

	return rooms, err
}

// GetRoomByNumber retrieves a room by its number
func (s *RoomService) GetRoomByNumber(ctx context.Context, roomNum int) (*models.Room, error) {
	var room *models.Room
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		room, err = s.roomRepo.GetRoomByNumber(tx, roomNum)
		return err
	})

	return room, err
}

// GetRoomStatusForDateRange gets the status of a room for each day in a date range
func (s *RoomService) GetRoomStatusForDateRange(ctx context.Context, roomNum int, startDate, endDate time.Time) ([]models.RoomStatus, error) {
	var statuses []models.RoomStatus
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		statuses, err = s.roomRepo.GetRoomStatusForDateRange(tx, roomNum, startDate, endDate)
		return err
	})

	return statuses, err
}

// GetRoomTypes retrieves all room types
func (s *RoomService) GetRoomTypes(ctx context.Context) ([]models.RoomType, error) {
	var roomTypes []models.RoomType
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		roomTypes, err = s.roomRepo.GetRoomTypes(tx)
		return err
	})

	return roomTypes, err
}

// GetRoomsByType retrieves all rooms of a specific type
func (s *RoomService) GetRoomsByType(ctx context.Context, typeID int) ([]models.Room, error) {
	var rooms []models.Room
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		rooms, err = s.roomRepo.GetRoomsByType(tx, typeID)
		return err
	})

	return rooms, err
}

// GetRoomAvailabilitySummary gets availability summary for all rooms in a date range
func (s *RoomService) GetRoomAvailabilitySummary(ctx context.Context, startDate, endDate time.Time) (map[int]map[string]int, error) {
	var summary map[int]map[string]int
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		summary, err = s.roomRepo.GetRoomAvailabilitySummary(tx, startDate, endDate)
		return err
	})

	return summary, err
}
