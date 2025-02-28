package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/utils/concurrency"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BookingRepository handles database operations for bookings
type BookingRepository struct {
	db       *gorm.DB
	roomLock *concurrency.RoomLock
}

// NewBookingRepository creates a new BookingRepository
func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{
		db:       db,
		roomLock: concurrency.NewRoomLock(db),
	}
}

// CheckRoomAvailability checks if a room is available for the requested dates
func (r *BookingRepository) CheckRoomAvailability(ctx context.Context, tx *gorm.DB, roomNum int, checkInDate, checkOutDate time.Time) (bool, error) {
	// Lock the room for the date range to prevent concurrent availability checks
	if err := r.roomLock.LockRoomDateRange(ctx, tx, roomNum, checkInDate, checkOutDate); err != nil {
		return false, err
	}

	// Check if the room is available for the requested dates
	var conflictCount int64
	err := tx.Model(&models.RoomStatus{}).
		Where("room_num = ? AND calendar BETWEEN ? AND ? AND status = 'Occupied'",
			roomNum,
			checkInDate.Format("2006-01-02"),
			checkOutDate.Format("2006-01-02")).
		Count(&conflictCount).Error

	if err != nil {
		return false, err
	}

	return conflictCount == 0, nil
}

// CreateBooking creates a new booking with payment in a single transaction (payment-first logic)
func (r *BookingRepository) CreateBooking(ctx context.Context, tx *gorm.DB, booking *models.Booking, payment *models.Receipt) error {
	// Check if the room exists
	var room models.Room
	if err := tx.First(&room, "room_num = ?", booking.RoomNum).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("room not found")
		}
		return err
	}

	// Check room availability with locking
	available, err := r.CheckRoomAvailability(ctx, tx, booking.RoomNum, booking.CheckInDate, booking.CheckOutDate)
	if err != nil {
		return err
	}

	if !available {
		return errors.New("room is not available for the selected dates")
	}

	// Generate booking ID (year + running number)
	var lastRunning models.LastRunning
	currentYear := time.Now().Year()

	// Get or create LastRunning record for current year with row-level locking
	err = tx.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).FirstOrCreate(&lastRunning, models.LastRunning{Year: currentYear}).Error
	if err != nil {
		return err
	}

	// Increment the running number atomically
	lastRunning.LastRunning++
	if err := tx.Save(&lastRunning).Error; err != nil {
		return err
	}

	// Set the booking ID (format: YYYYXXXXXX)
	booking.BookingID = currentYear*1000000 + lastRunning.LastRunning

	// Set current timestamp for booking date if not provided
	if booking.BookingDate.IsZero() {
		booking.BookingDate = time.Now()
	}

	// Calculate the total price based on the room type's price per night and stay duration
	var roomType models.RoomType
	if err := tx.Model(&models.RoomType{}).
		Select("price_per_night").
		Joins("JOIN rooms ON rooms.type_id = room_types.type_id").
		Where("rooms.room_num = ?", booking.RoomNum).
		First(&roomType).Error; err != nil {
		return err
	}

	// Calculate number of nights
	nights := int(booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24)
	if nights < 1 {
		nights = 1 // Minimum 1 night
	}
	booking.TotalPrice = roomType.PricePerNight * nights

	// Create the booking within the transaction
	if err := tx.Create(booking).Error; err != nil {
		return err
	}

	// Process payment
	if payment != nil {
		payment.BookingID = booking.BookingID
		payment.Amount = booking.TotalPrice

		if payment.IssueDate.IsZero() {
			payment.IssueDate = time.Now()
		}

		if err := tx.Create(payment).Error; err != nil {
			return err
		}
	}

	// Update room status for each day of the booking
	current := booking.CheckInDate
	for current.Before(booking.CheckOutDate) || current.Equal(booking.CheckOutDate) {
		status := models.RoomStatus{
			RoomNum:   booking.RoomNum,
			Calendar:  current,
			Status:    "Occupied",
			BookingID: &booking.BookingID,
		}

		// Use upsert to handle potential existing records
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

// GetBookingByID retrieves a booking by its ID
func (r *BookingRepository) GetBookingByID(tx *gorm.DB, bookingID int) (*models.Booking, error) {
	var booking models.Booking
	err := tx.Preload("Room.RoomType").
		Preload("Receipt").
		First(&booking, "booking_id = ?", bookingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}
	return &booking, nil
}

// UpdateBooking updates an existing booking with optimistic concurrency control
func (r *BookingRepository) UpdateBooking(ctx context.Context, tx *gorm.DB, booking *models.Booking) error {
	// Get the current booking to check for changes
	var existingBooking models.Booking
	if err := tx.First(&existingBooking, "booking_id = ?", booking.BookingID).Error; err != nil {
		return err
	}

	// Lock the room for the booking date range to prevent concurrent modifications
	if err := r.roomLock.LockRoomDateRange(ctx, tx, booking.RoomNum, booking.CheckInDate, booking.CheckOutDate); err != nil {
		return err
	}

	// Check if dates have changed
	datesChanged := !existingBooking.CheckInDate.Equal(booking.CheckInDate) ||
		!existingBooking.CheckOutDate.Equal(booking.CheckOutDate) ||
		existingBooking.RoomNum != booking.RoomNum

	if datesChanged {
		// Check if the room is available for the new dates
		var conflictCount int64
		err := tx.Model(&models.RoomStatus{}).
			Where("room_num = ? AND calendar BETWEEN ? AND ? AND status = 'Occupied' AND booking_id != ?",
				booking.RoomNum,
				booking.CheckInDate.Format("2006-01-02"),
				booking.CheckOutDate.Format("2006-01-02"),
				booking.BookingID).
			Count(&conflictCount).Error

		if err != nil {
			return err
		}

		if conflictCount > 0 {
			return errors.New("room is not available for the selected dates")
		}

		// Clear old room status entries
		if err := tx.Where("booking_id = ?", booking.BookingID).Delete(&models.RoomStatus{}).Error; err != nil {
			return err
		}

		// Calculate the total price based on the room type's price per night and stay duration
		var roomType models.RoomType
		if err := tx.Model(&models.RoomType{}).
			Select("price_per_night").
			Joins("JOIN rooms ON rooms.type_id = room_types.type_id").
			Where("rooms.room_num = ?", booking.RoomNum).
			First(&roomType).Error; err != nil {
			return err
		}

		// Calculate number of nights
		nights := int(booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24)
		if nights < 1 {
			nights = 1 // Minimum 1 night
		}
		booking.TotalPrice = roomType.PricePerNight * nights

		// Create new room status entries
		current := booking.CheckInDate
		for current.Before(booking.CheckOutDate) || current.Equal(booking.CheckOutDate) {
			status := models.RoomStatus{
				RoomNum:   booking.RoomNum,
				Calendar:  current,
				Status:    "Occupied",
				BookingID: &booking.BookingID,
			}

			// Use upsert to handle potential existing records
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "room_num"}, {Name: "calendar"}},
				DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
			}).Create(&status).Error; err != nil {
				return err
			}

			current = current.AddDate(0, 0, 1)
		}
	}

	// Update the booking
	return tx.Save(booking).Error
}

// CancelBooking cancels a booking and frees up the room
func (r *BookingRepository) CancelBooking(ctx context.Context, tx *gorm.DB, bookingID int) error {
	// Lock the booking record for update
	var booking models.Booking
	if err := tx.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		return err
	}

	// Lock the room date range
	if err := r.roomLock.LockRoomDateRange(ctx, tx, booking.RoomNum, booking.CheckInDate, booking.CheckOutDate); err != nil {
		return err
	}

	// Check if there's a receipt (payment) for this booking
	var receiptCount int64
	if err := tx.Model(&models.Receipt{}).Where("booking_id = ?", bookingID).Count(&receiptCount).Error; err != nil {
		return err
	}

	if receiptCount > 0 {
		// If paid, we might want to handle refunds or keep a record
		// Here we'll just set the status to Cancelled but keep the record
		return tx.Model(&models.RoomStatus{}).
			Where("booking_id = ?", bookingID).
			Update("status", "Available").
			Update("booking_id", nil).Error
	} else {
		// If not paid, we can completely delete the booking
		// First clear room status entries
		if err := tx.Where("booking_id = ?", bookingID).Delete(&models.RoomStatus{}).Error; err != nil {
			return err
		}

		// Then delete the booking itself
		return tx.Delete(&booking).Error
	}
}

// SearchAvailableRooms finds rooms available for a given date range
func (r *BookingRepository) SearchAvailableRooms(tx *gorm.DB, checkInDate, checkOutDate time.Time, guestCount int) ([]models.Room, error) {
	var availableRooms []models.Room

	// Find rooms that are available for the entire date range and can accommodate the guest count
	query := tx.Model(&models.Room{}).
		Distinct("rooms.*").
		Joins("JOIN room_types ON rooms.type_id = room_types.type_id").
		Where("room_types.\"no_of_guest\" >= ?", guestCount).
		// Subquery to exclude rooms that have any Occupied status during the requested period
		Where("NOT EXISTS (SELECT 1 FROM room_statuses WHERE room_statuses.room_num = rooms.room_num "+
			"AND room_statuses.calendar BETWEEN ? AND ? AND room_statuses.status = 'Occupied')",
			checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02")).
		Preload("RoomType").
		Order("rooms.room_num")

	err := query.Find(&availableRooms).Error
	return availableRooms, err
}
