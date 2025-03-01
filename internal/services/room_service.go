package services

import (
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"gorm.io/gorm"
)

// RoomService handles business logic for rooms
type RoomService struct {
	db       *gorm.DB
	roomRepo *repositories.RoomRepository
}

// NewRoomService creates a new RoomService
func NewRoomService(db *gorm.DB, roomRepo *repositories.RoomRepository) *RoomService {
	return &RoomService{
		db:       db,
		roomRepo: roomRepo,
	}
}

// GetAllRoomsWithDetails retrieves all rooms with their types and facilities
func (s *RoomService) GetAllRoomsWithDetails(tx *gorm.DB) ([]models.Room, error) {
	return s.roomRepo.GetAllRoomsWithDetails(tx)
}

// GetRoomWithDetails retrieves a room by ID with its type and facilities
func (s *RoomService) GetRoomWithDetails(tx *gorm.DB, roomNum int) (*models.Room, error) {
	return s.roomRepo.GetRoomWithDetails(tx, roomNum)
}

// GetRoomsByTypeWithDetails retrieves all rooms of a specific type with facilities
func (s *RoomService) GetRoomsByTypeWithDetails(tx *gorm.DB, typeID int) ([]models.Room, error) {
	return s.roomRepo.GetRoomsByTypeWithDetails(tx, typeID)
}

// GetAllRoomTypes retrieves all room types with their facilities
func (s *RoomService) GetAllRoomTypes(tx *gorm.DB) ([]models.RoomType, error) {
	return s.roomRepo.GetAllRoomTypes(tx)
}

// GetRoomCalendar retrieves the availability calendar for a room
func (s *RoomService) GetRoomCalendar(tx *gorm.DB, roomNum int, startDate, endDate string) ([]models.RoomStatus, error) {
	// Validate date format
	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	return s.roomRepo.GetRoomCalendar(tx, roomNum, startDate, endDate)
}
