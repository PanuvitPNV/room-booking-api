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

// GetAllRooms retrieves all rooms
func (s *RoomService) GetAllRooms(tx *gorm.DB) ([]models.Room, error) {
	return s.roomRepo.GetAllRooms(tx)
}

// GetRoomByID retrieves a room by ID
func (s *RoomService) GetRoomByID(tx *gorm.DB, roomNum int) (*models.Room, error) {
	return s.roomRepo.GetRoomByID(tx, roomNum)
}

// GetRoomWithFacilities retrieves a room with all its facilities
func (s *RoomService) GetRoomWithFacilities(tx *gorm.DB, roomNum int) (*models.Room, error) {
	return s.roomRepo.GetRoomWithFacilities(tx, roomNum)
}

// GetRoomsByType retrieves all rooms of a specific type
func (s *RoomService) GetRoomsByType(tx *gorm.DB, typeID int) ([]models.Room, error) {
	return s.roomRepo.GetRoomsByType(tx, typeID)
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

// GetAllRoomTypes retrieves all room types
func (s *RoomService) GetAllRoomTypes(tx *gorm.DB) ([]models.RoomType, error) {
	return s.roomRepo.GetAllRoomTypes(tx)
}

// GetRoomTypeByID retrieves a room type by ID
func (s *RoomService) GetRoomTypeByID(tx *gorm.DB, typeID int) (*models.RoomType, error) {
	return s.roomRepo.GetRoomTypeByID(tx, typeID)
}

// GetAllFacilities retrieves all facilities
func (s *RoomService) GetAllFacilities(tx *gorm.DB) ([]models.Facility, error) {
	return s.roomRepo.GetAllFacilities(tx)
}

// GetFacilityByID retrieves a facility by ID
func (s *RoomService) GetFacilityByID(tx *gorm.DB, facilityID int) (*models.Facility, error) {
	return s.roomRepo.GetFacilityByID(tx, facilityID)
}
