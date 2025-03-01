package repositories

import (
	"errors"

	"github.com/panuvitpnv/room-booking-api/internal/models"
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

// GetAllRoomsWithDetails retrieves all rooms with their types and facilities
func (r *RoomRepository) GetAllRoomsWithDetails(tx *gorm.DB) ([]models.Room, error) {
	// First, get all rooms with their types
	var rooms []models.Room
	if err := tx.Preload("RoomType").Find(&rooms).Error; err != nil {
		return nil, err
	}

	// For each room type, get all facilities
	for i, room := range rooms {
		// Get all facility IDs for this room type
		var roomFacilities []models.RoomFacility
		if err := tx.Where("type_id = ?", room.TypeID).Find(&roomFacilities).Error; err != nil {
			return nil, err
		}

		// Initialize the RoomFacilities array
		rooms[i].RoomType.RoomFacilities = make([]models.RoomFacility, 0)

		// For each facility ID, get the complete facility and add to room type
		for _, rf := range roomFacilities {
			var facility models.Facility
			if err := tx.Where("fac_id = ?", rf.FacilityID).First(&facility).Error; err != nil {
				continue // Skip if facility not found
			}

			// Create complete RoomFacility with Facility
			completeRF := models.RoomFacility{
				TypeID:     rf.TypeID,
				FacilityID: rf.FacilityID,
				Facility:   facility,
			}

			// Add to room's RoomFacilities
			rooms[i].RoomType.RoomFacilities = append(rooms[i].RoomType.RoomFacilities, completeRF)
		}
	}

	return rooms, nil
}

// GetRoomWithDetails retrieves a room by ID with its type and facilities
func (r *RoomRepository) GetRoomWithDetails(tx *gorm.DB, roomNum int) (*models.Room, error) {
	// Get the room with its type
	var room models.Room
	if err := tx.Preload("RoomType").Where("room_num = ?", roomNum).First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}

	// Get all facility IDs for this room type
	var roomFacilities []models.RoomFacility
	if err := tx.Where("type_id = ?", room.TypeID).Find(&roomFacilities).Error; err != nil {
		return nil, err
	}

	// Initialize the RoomFacilities array
	room.RoomType.RoomFacilities = make([]models.RoomFacility, 0)

	// For each facility ID, get the complete facility and add to room type
	for _, rf := range roomFacilities {
		var facility models.Facility
		if err := tx.Where("fac_id = ?", rf.FacilityID).First(&facility).Error; err != nil {
			continue // Skip if facility not found
		}

		// Create complete RoomFacility with Facility
		completeRF := models.RoomFacility{
			TypeID:     rf.TypeID,
			FacilityID: rf.FacilityID,
			Facility:   facility,
		}

		// Add to room's RoomFacilities
		room.RoomType.RoomFacilities = append(room.RoomType.RoomFacilities, completeRF)
	}

	return &room, nil
}

// GetRoomsByTypeWithDetails retrieves all rooms of a specific room type with facilities
func (r *RoomRepository) GetRoomsByTypeWithDetails(tx *gorm.DB, typeID int) ([]models.Room, error) {
	// Get rooms of this type
	var rooms []models.Room
	if err := tx.Preload("RoomType").Where("type_id = ?", typeID).Find(&rooms).Error; err != nil {
		return nil, err
	}

	if len(rooms) == 0 {
		return []models.Room{}, nil
	}

	// Get all facility IDs for this room type
	var roomFacilities []models.RoomFacility
	if err := tx.Where("type_id = ?", typeID).Find(&roomFacilities).Error; err != nil {
		return nil, err
	}

	// Get all facilities for these IDs
	facilityMap := make(map[int]models.Facility)
	for _, rf := range roomFacilities {
		var facility models.Facility
		if err := tx.Where("fac_id = ?", rf.FacilityID).First(&facility).Error; err == nil {
			facilityMap[rf.FacilityID] = facility
		}
	}

	// Create complete RoomFacility objects with Facility
	completeFacilities := make([]models.RoomFacility, 0)
	for _, rf := range roomFacilities {
		if facility, exists := facilityMap[rf.FacilityID]; exists {
			completeFacilities = append(completeFacilities, models.RoomFacility{
				TypeID:     rf.TypeID,
				FacilityID: rf.FacilityID,
				Facility:   facility,
			})
		}
	}

	// Add the same facilities to all rooms of this type
	for i := range rooms {
		rooms[i].RoomType.RoomFacilities = completeFacilities
	}

	return rooms, nil
}

// GetAllRoomTypes retrieves all room types with their facilities
func (r *RoomRepository) GetAllRoomTypes(tx *gorm.DB) ([]models.RoomType, error) {
	// Get all room types
	var roomTypes []models.RoomType
	if err := tx.Find(&roomTypes).Error; err != nil {
		return nil, err
	}

	// For each room type, get all facilities
	for i, roomType := range roomTypes {
		// Get all facility IDs for this room type
		var roomFacilities []models.RoomFacility
		if err := tx.Where("type_id = ?", roomType.TypeID).Find(&roomFacilities).Error; err != nil {
			return nil, err
		}

		// Initialize the RoomFacilities array
		roomTypes[i].RoomFacilities = make([]models.RoomFacility, 0)

		// For each facility ID, get the complete facility and add to room type
		for _, rf := range roomFacilities {
			var facility models.Facility
			if err := tx.Where("fac_id = ?", rf.FacilityID).First(&facility).Error; err != nil {
				continue // Skip if facility not found
			}

			// Create complete RoomFacility with Facility
			completeRF := models.RoomFacility{
				TypeID:     rf.TypeID,
				FacilityID: rf.FacilityID,
				Facility:   facility,
			}

			// Add to room type's RoomFacilities
			roomTypes[i].RoomFacilities = append(roomTypes[i].RoomFacilities, completeRF)
		}
	}

	return roomTypes, nil
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
