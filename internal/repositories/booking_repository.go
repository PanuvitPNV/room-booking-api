package repositories

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/utils"
)

// BookingRepository handles database operations for bookings
type BookingRepository struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *gorm.DB, logger *log.Logger) *BookingRepository {
	return &BookingRepository{
		db:     db,
		logger: logger,
	}
}

// IsRoomAvailable checks if a room is available for the specified dates
func (r *BookingRepository) IsRoomAvailable(tx *gorm.DB, roomNum int, checkIn, checkOut time.Time) (bool, error) {
	if utils.DeadlockTesting.Enabled {
		// DEADLOCK MODE: Use a separate query with FOR UPDATE to get all the room statuses first
		var roomStatuses []models.RoomStatus

		// First, lock some rows to create potential deadlock conditions without using aggregate functions
		err := tx.Model(&models.RoomStatus{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ? AND calendar >= ?",
					roomNum, checkIn.Format("2006-01-02")).
			Limit(1). // Just get one row to lock - enough for deadlock potential
			Find(&roomStatuses).Error

		if err != nil {
			return false, err
		}

		// Add artificial delay after acquiring locks to increase deadlock chances
		utils.DelayIfTesting(100 * time.Millisecond)
	}

	// Normal behavior - perform the count query
	var count int64
	err := tx.Model(&models.RoomStatus{}).
		Where("room_num = ? AND calendar >= ? AND calendar < ? AND status = ?",
			roomNum, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02"), "Occupied").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// CreateBooking creates a new booking with semantic booking ID
func (r *BookingRepository) CreateBooking(tx *gorm.DB, booking *models.Booking) error {
	// Generate semantic booking ID with random suffix
	generateBookingID := func() int {
		baseID := time.Now().Year()*100000000 + int(time.Now().Month())*1000000 + time.Now().Day()*10000 + booking.RoomNum
		randomSuffix := rand.Intn(900) + 100 // 3-digit random number (100-999)
		return baseID*1000 + randomSuffix
	}

	for retries := 0; retries < 5; retries++ {
		booking.BookingID = generateBookingID()

		// DEADLOCK SCENARIO 3: Use stronger lock level on room lookup
		var roomType models.RoomType
		if err := tx.Model(&models.Room{}).
			Clauses(clause.Locking{Strength: "UPDATE"}). // Add explicit lock
			Select("room_types.price_per_night").
			Joins("JOIN room_types ON rooms.type_id = room_types.type_id").
			Where("rooms.room_num = ?", booking.RoomNum).
			First(&roomType).Error; err != nil {
			return err
		}

		// DEADLOCK SCENARIO 4: Add delay after acquiring locks
		time.Sleep(150 * time.Millisecond)

		nights := int(booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24)
		if nights < 1 {
			return errors.New("check-out date must be at least one day after check-in date")
		}

		booking.TotalPrice = roomType.PricePerNight * nights
		booking.BookingDate = time.Now()

		// Try to create the booking
		if err := tx.Create(booking).Error; err != nil {
			// Retry on duplicate key violation
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				continue
			}
			return err
		}

		r.logger.Printf("Created booking %d for room %d", booking.BookingID, booking.RoomNum)

		// Update room statuses for each day of the booking
		current := booking.CheckInDate
		// Use less than to exclude the checkout day (changed from the original)
		for current.Before(booking.CheckOutDate) {
			status := models.RoomStatus{
				RoomNum:   booking.RoomNum,
				Calendar:  current,
				Status:    "Occupied",
				BookingID: &booking.BookingID,
			}

			// DEADLOCK SCENARIO 5: Use a more aggressive conflict resolution strategy that can lead to deadlocks
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "room_num"}, {Name: "calendar"}},
				DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
			}).Create(&status).Error; err != nil {
				return err
			}

			current = current.AddDate(0, 0, 1)
		}

		return nil
	}

	return errors.New("failed to generate unique booking ID after multiple retries")
}

// GetBookingByID retrieves a booking by its ID
func (r *BookingRepository) GetBookingByID(tx *gorm.DB, bookingID int) (*models.Booking, error) {
	var booking models.Booking

	// DEADLOCK SCENARIO 6: Use pessimistic locking to hold locks longer
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Room.RoomType").
		First(&booking, "booking_id = ?", bookingID).Error

	if err != nil {
		return nil, err
	}

	// DEADLOCK SCENARIO 7: Add artificial delay while holding the lock
	time.Sleep(200 * time.Millisecond)

	return &booking, nil
}

// CancelBooking cancels a booking and releases the room
func (r *BookingRepository) CancelBooking(tx *gorm.DB, bookingID int) error {
	var booking models.Booking

	// DEADLOCK SCENARIO 8: Use explicit FOR UPDATE lock
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		return err
	}

	r.logger.Printf("Cancelling booking %d for room %d", bookingID, booking.RoomNum)

	// DEADLOCK SCENARIO 9: Check room status with additional locks
	var roomStatus models.RoomStatus
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("room_num = ? AND booking_id = ?", booking.RoomNum, bookingID).
		First(&roomStatus).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var receiptCount int64
	if err := tx.Model(&models.Receipt{}).
		Where("booking_id = ?", bookingID).
		Count(&receiptCount).Error; err != nil {
		return err
	}

	if receiptCount > 0 {
		return errors.New("cannot cancel a booking that has already been paid for")
	}

	// DEADLOCK SCENARIO 10: Add delay between operations to increase deadlock chance
	time.Sleep(250 * time.Millisecond)

	if err := tx.Model(&models.RoomStatus{}).
		Where("booking_id = ?", bookingID).
		Updates(map[string]interface{}{
			"status":     "Available",
			"booking_id": nil,
		}).Error; err != nil {
		return err
	}

	if err := tx.Delete(&booking).Error; err != nil {
		return err
	}

	return nil
}

// UpdateBooking updates an existing booking with optimistic concurrency control
func (r *BookingRepository) UpdateBooking(tx *gorm.DB, bookingID int, updateData map[string]interface{}) error {
	var booking models.Booking

	if utils.DeadlockTesting.Enabled {
		// DEADLOCK MODE: Use pessimistic locking with FOR UPDATE
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&booking, "booking_id = ?", bookingID).Error; err != nil {
			return err
		}

		r.logger.Printf("Updating booking %d for room %d", bookingID, booking.RoomNum)

		// In test mode, lock the room status records first to create deadlock potential
		var roomStatuses []models.RoomStatus
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ? AND booking_id = ?", booking.RoomNum, bookingID).
			Find(&roomStatuses).Error; err != nil {
			return err
		}

		r.logger.Printf("Acquired lock on %d room status records for booking %d", len(roomStatuses), bookingID)

		// Add a delay to increase chance of deadlock
		utils.DelayIfTesting(200 * time.Millisecond)

		// Update without optimistic concurrency control in test mode
		if err := tx.Model(&models.Booking{}).
			Where("booking_id = ?", bookingID).
			Updates(updateData).Error; err != nil {
			return err
		}

		return nil
	} else {
		// NORMAL MODE: Use optimistic concurrency control
		if err := tx.First(&booking, "booking_id = ?", bookingID).Error; err != nil {
			return err
		}

		currentUpdatedAt := booking.UpdatedAt

		result := tx.Model(&models.Booking{}).
			Where("booking_id = ? AND updated_at = ?", bookingID, currentUpdatedAt).
			Updates(updateData)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("booking %d was updated by another transaction, please retry", bookingID)
		}

		return nil
	}
}

// GetAvailableRoomsByDateRange finds available rooms for a date range
func (r *BookingRepository) GetAvailableRoomsByDateRange(tx *gorm.DB, startDate, endDate time.Time) ([]models.Room, error) {
	// First, get all rooms
	var rooms []models.Room

	if err := tx.Preload("RoomType").Find(&rooms).Error; err != nil {
		return nil, err
	}

	// Filter out rooms that are already booked during the requested period
	var availableRooms []models.Room
	for _, room := range rooms {
		// Check if the context is done/canceled between room checks
		if tx.Statement.Context != nil {
			select {
			case <-tx.Statement.Context.Done():
				return nil, tx.Statement.Context.Err()
			default:
				// Continue processing
			}
		}

		var occupiedDaysCount int64
		err := tx.Model(&models.RoomStatus{}).
			Where("room_num = ? AND calendar >= ? AND calendar < ? AND status = ?",
				room.RoomNum, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), "Occupied").
			Count(&occupiedDaysCount).Error

		if err != nil {
			return nil, err
		}

		if occupiedDaysCount == 0 {
			availableRooms = append(availableRooms, room)
		}

		// In deadlock test mode, add a very small delay (not enough to cause timeouts)
		if os.Getenv("DEADLOCK_TEST_MODE") == "true" {
			time.Sleep(5 * time.Millisecond)
		}
	}

	return availableRooms, nil
}
