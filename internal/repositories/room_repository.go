package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/panuvitpnv/room-booking-api/internal/models"
)

// RoomRepository handles database operations for rooms
type RoomRepository struct {
	db *gorm.DB
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

// GetAllRooms retrieves all rooms with their types and facilities
func (r *RoomRepository) GetAllRooms(tx *gorm.DB) ([]models.Room, error) {
	var rooms []models.Room
	// Preload room type and facilities through RoomType relationship
	err := tx.Preload("RoomType.RoomFacilities.Facility").Preload("RoomType").Find(&rooms).Error
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

// In your room repository
func (r *RoomRepository) GetRoomByNumber(tx *gorm.DB, roomNum int) (*models.Room, error) {
	var room models.Room
	// Explicitly include the Facility object in the preload chain
	err := tx.Preload("RoomType").
		Preload("RoomType.RoomFacilities").
		Preload("RoomType.RoomFacilities.Facility"). // This is the key line
		First(&room, roomNum).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// GetRoomWithFacilities retrieves a room with detailed facility information
func (r *RoomRepository) GetRoomWithFacilities(tx *gorm.DB, roomNum int) (*models.Room, []models.Facility, error) {
	var room models.Room

	// First get the room with its type
	err := tx.Preload("RoomType").First(&room, roomNum).Error
	if err != nil {
		return nil, nil, err
	}

	// Then get facilities for this room type
	var facilities []models.Facility
	err = tx.Table("facilities").
		Joins("JOIN room_facilities ON facilities.fac_id = room_facilities.fac_id").
		Where("room_facilities.type_id = ?", room.TypeID).
		Find(&facilities).Error

	if err != nil {
		return nil, nil, err
	}

	return &room, facilities, nil
}

// GetRoomStatusForDateRange gets the status of a room for each day in a date range
func (r *RoomRepository) GetRoomStatusForDateRange(tx *gorm.DB, roomNum int, startDate, endDate time.Time) ([]models.RoomStatus, error) {
	var statuses []models.RoomStatus

	err := tx.Where("room_num = ? AND calendar >= ? AND calendar < ?",
		roomNum, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("calendar ASC").
		Find(&statuses).Error

	if err != nil {
		return nil, err
	}

	// If we don't have status records for some days, fill in with default available status
	result := make([]models.RoomStatus, 0)
	currentDate := startDate

	for currentDate.Before(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		found := false

		for _, status := range statuses {
			if status.Calendar.Format("2006-01-02") == dateStr {
				result = append(result, status)
				found = true
				break
			}
		}

		if !found {
			// Create default available status for this date
			result = append(result, models.RoomStatus{
				RoomNum:  roomNum,
				Calendar: currentDate,
				Status:   "Available",
			})
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}

// GetRoomTypes retrieves all room types with their facilities
func (r *RoomRepository) GetRoomTypes(tx *gorm.DB) ([]models.RoomType, error) {
	var roomTypes []models.RoomType
	err := tx.Preload("RoomFacilities.Facility").Find(&roomTypes).Error
	if err != nil {
		return nil, err
	}
	return roomTypes, nil
}

// GetRoomAvailabilitySummary gets availability summary for all rooms in a date range
func (r *RoomRepository) GetRoomAvailabilitySummary(tx *gorm.DB, startDate, endDate time.Time) (map[int]map[string]int, error) {
	// Initialize result map: room number -> status counts
	result := make(map[int]map[string]int)

	// Get all rooms
	var rooms []models.Room
	if err := tx.Find(&rooms).Error; err != nil {
		return nil, err
	}

	// Initialize counters for each room
	for _, room := range rooms {
		result[room.RoomNum] = map[string]int{
			"Available": 0,
			"Occupied":  0,
		}
	}

	// Query existing status records
	var statuses []models.RoomStatus
	err := tx.Where("calendar >= ? AND calendar < ?",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Find(&statuses).Error

	if err != nil {
		return nil, err
	}

	// Count days by status for each room
	for _, status := range statuses {
		if _, exists := result[status.RoomNum]; exists {
			result[status.RoomNum][status.Status]++
		}
	}

	// Calculate total days in range
	totalDays := int(endDate.Sub(startDate).Hours() / 24)

	// Fill in missing days as Available
	for roomNum, counts := range result {
		recordedDays := counts["Available"] + counts["Occupied"]
		if recordedDays < totalDays {
			result[roomNum]["Available"] += (totalDays - recordedDays)
		}
	}

	return result, nil
}

// GetRoomsByType retrieves all rooms of a specific type with facilities
func (r *RoomRepository) GetRoomsByType(tx *gorm.DB, typeID int) ([]models.Room, error) {
	var rooms []models.Room
	err := tx.Where("type_id = ?", typeID).
		Preload("RoomType.RoomFacilities.Facility").
		Preload("RoomType").
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetFacilitiesByRoomType retrieves all facilities for a specific room type
func (r *RoomRepository) GetFacilitiesByRoomType(tx *gorm.DB, typeID int) ([]models.Facility, error) {
	var facilities []models.Facility

	err := tx.Table("facilities").
		Joins("JOIN room_facilities ON facilities.fac_id = room_facilities.fac_id").
		Where("room_facilities.type_id = ?", typeID).
		Find(&facilities).Error

	if err != nil {
		return nil, err
	}

	return facilities, nil
}

// GetAllFacilities retrieves all available facilities
func (r *RoomRepository) GetAllFacilities(tx *gorm.DB) ([]models.Facility, error) {
	var facilities []models.Facility
	err := tx.Find(&facilities).Error
	if err != nil {
		return nil, err
	}
	return facilities, nil
}
