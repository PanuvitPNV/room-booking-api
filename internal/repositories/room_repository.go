package repositories

import (
	"errors"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/utils/concurrency"
	"gorm.io/gorm"
)

// RoomRepository handles database operations for room data
type RoomRepository struct {
	db *gorm.DB
}

// NewRoomRepository creates a new RoomRepository
func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

// GetAllRooms retrieves all rooms with their room types
func (r *RoomRepository) GetAllRooms(tx *gorm.DB) ([]models.Room, error) {
	var rooms []models.Room
	err := tx.Preload("RoomType").Order("room_num").Find(&rooms).Error
	return rooms, err
}

// GetRoomByID retrieves a room by its ID with its room type
func (r *RoomRepository) GetRoomByID(tx *gorm.DB, roomNum int) (*models.Room, error) {
	var room models.Room
	err := tx.Preload("RoomType").First(&room, "room_num = ?", roomNum).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}
	return &room, nil
}

// GetRoomWithFacilities retrieves a room by ID with its room type and facilities
func (r *RoomRepository) GetRoomWithFacilities(tx *gorm.DB, roomNum int) (*models.Room, error) {
	var room models.Room
	err := tx.Preload("RoomType.RoomFacilities.Facility").
		First(&room, "room_num = ?", roomNum).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}

	return &room, nil
}

// GetRoomsByType retrieves all rooms of a specific room type
func (r *RoomRepository) GetRoomsByType(tx *gorm.DB, typeID int) ([]models.Room, error) {
	var rooms []models.Room
	err := tx.Preload("RoomType").
		Where("type_id = ?", typeID).
		Order("room_num").
		Find(&rooms).Error

	return rooms, err
}

// GetRoomCalendar retrieves the room status calendar for a specific room and date range
func (r *RoomRepository) GetRoomCalendar(tx *gorm.DB, roomNum int, startDate, endDate string) ([]models.RoomStatus, error) {
	var statuses []models.RoomStatus

	err := tx.Where("room_num = ? AND calendar BETWEEN ? AND ?",
		roomNum, startDate, endDate).
		Order("calendar").
		Find(&statuses).Error

	return statuses, err
}

// GetAllRoomTypes retrieves all room types with their facilities
func (r *RoomRepository) GetAllRoomTypes(tx *gorm.DB) ([]models.RoomType, error) {
	var roomTypes []models.RoomType

	err := concurrency.WithSelectForShare(tx).
		Preload("RoomFacilities.Facility").
		Order("type_id").
		Find(&roomTypes).Error

	return roomTypes, err
}

// GetRoomTypeByID retrieves a room type by ID with its facilities
func (r *RoomRepository) GetRoomTypeByID(tx *gorm.DB, typeID int) (*models.RoomType, error) {
	var roomType models.RoomType

	err := tx.Preload("RoomFacilities.Facility").
		First(&roomType, "type_id = ?", typeID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room type not found")
		}
		return nil, err
	}

	return &roomType, nil
}

// GetAllFacilities retrieves all facilities
func (r *RoomRepository) GetAllFacilities(tx *gorm.DB) ([]models.Facility, error) {
	var facilities []models.Facility
	err := tx.Order("fac_id").Find(&facilities).Error
	return facilities, err
}

// GetFacilityByID retrieves a facility by ID
func (r *RoomRepository) GetFacilityByID(tx *gorm.DB, facilityID int) (*models.Facility, error) {
	var facility models.Facility

	err := tx.First(&facility, "fac_id = ?", facilityID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("facility not found")
		}
		return nil, err
	}

	return &facility, nil
}
