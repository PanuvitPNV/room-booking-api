package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"github.com/panuvitpnv/room-booking-api/internal/utils"
)

// BookingService handles booking business logic
type BookingService struct {
	bookingRepo *repositories.BookingRepository
	roomRepo    *repositories.RoomRepository
	lockManager *utils.LockManager
	logger      *log.Logger
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo *repositories.BookingRepository,
	roomRepo *repositories.RoomRepository,
	lockManager *utils.LockManager,
	logger *log.Logger,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		lockManager: lockManager,
		logger:      logger,
	}
}

// CreateBooking handles the creation of a new booking with transaction management
func (s *BookingService) CreateBooking(ctx context.Context, booking *models.Booking) error {
	// Validate booking dates
	if booking.CheckInDate.IsZero() || booking.CheckOutDate.IsZero() {
		return errors.New("check-in and check-out dates are required")
	}

	if booking.CheckOutDate.Before(booking.CheckInDate) || booking.CheckOutDate.Equal(booking.CheckInDate) {
		return errors.New("check-out date must be after check-in date")
	}

	if booking.CheckInDate.Before(time.Now()) {
		return errors.New("check-in date cannot be in the past")
	}

	// Only acquire application-level lock in normal mode
	if !utils.DeadlockTesting.Enabled {
		// Normal production mode - use application lock
		unlock, err := s.lockManager.AcquireLock("room", booking.RoomNum)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		defer unlock()
	} else {
		// In test mode, log the operation
		if s.logger != nil {
			s.logger.Printf("Creating booking for room %d", booking.RoomNum)
		}
	}

	// Execute booking creation within a transaction with retries
	return utils.RunWithRetry(3, func() error {
		return utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			// In test mode, add delay to increase conflict probability
			if utils.DeadlockTesting.Enabled {
				utils.DelayIfTesting(100 * time.Millisecond)
			}

			// Check if the room exists
			_, err := s.roomRepo.GetRoomByNumber(tx, booking.RoomNum)
			if err != nil {
				return fmt.Errorf("failed to find room: %w", err)
			}

			// Check if the room is available for the requested dates
			available, err := s.bookingRepo.IsRoomAvailable(tx, booking.RoomNum, booking.CheckInDate, booking.CheckOutDate)
			if err != nil {
				return fmt.Errorf("failed to check room availability: %w", err)
			}

			if !available {
				return errors.New("room is not available for the requested dates")
			}

			// Create the booking
			if err := s.bookingRepo.CreateBooking(tx, booking); err != nil {
				return fmt.Errorf("failed to create booking: %w", err)
			}

			// In test mode, add another delay to increase deadlock chance
			if utils.DeadlockTesting.Enabled {
				utils.DelayIfTesting(150 * time.Millisecond)
			}

			return nil
		})
	})
}

// GetBookingByID retrieves a booking by ID
func (s *BookingService) GetBookingByID(ctx context.Context, bookingID int) (*models.Booking, error) {
	var booking *models.Booking
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		booking, err = s.bookingRepo.GetBookingByID(tx, bookingID)
		return err
	})

	return booking, err
}

// CancelBooking cancels a booking with transaction management
func (s *BookingService) CancelBooking(ctx context.Context, bookingID int) error {
	// First get the booking details to know which room to lock
	var booking *models.Booking
	err := utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		booking, err = s.bookingRepo.GetBookingByID(tx, bookingID)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	// DEADLOCK SCENARIO 3: Remove the application-level lock to allow database-level deadlocks
	/*
		unlock, err := s.lockManager.AcquireLock("room", booking.RoomNum)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		defer unlock()
	*/

	s.logger.Printf("Cancelling booking %d for room %d", bookingID, booking.RoomNum)

	// Execute cancellation within a transaction with retries
	return utils.RunWithRetry(3, func() error {
		return utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			// DEADLOCK SCENARIO 4: Add a delay to increase the chance of deadlocks
			time.Sleep(300 * time.Millisecond)

			return s.bookingRepo.CancelBooking(tx, bookingID)
		})
	})
}

// GetAvailableRooms finds available rooms for a date range
func (s *BookingService) GetAvailableRooms(ctx context.Context, startDate, endDate time.Time) ([]models.Room, error) {
	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("start and end dates are required")
	}

	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return nil, errors.New("end date must be after start date")
	}

	var rooms []models.Room
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		rooms, err = s.bookingRepo.GetAvailableRoomsByDateRange(tx, startDate, endDate)
		return err
	})

	return rooms, err
}

// UpdateBooking updates an existing booking with optimistic concurrency control
func (s *BookingService) UpdateBooking(ctx context.Context, bookingID int, newCheckIn, newCheckOut time.Time) error {
	// Validate new dates
	if newCheckIn.IsZero() || newCheckOut.IsZero() {
		return errors.New("check-in and check-out dates are required")
	}

	if newCheckOut.Before(newCheckIn) || newCheckOut.Equal(newCheckIn) {
		return errors.New("check-out date must be after check-in date")
	}

	if newCheckIn.Before(time.Now()) {
		return errors.New("check-in date cannot be in the past")
	}

	// First get the booking to know which room to lock
	var booking *models.Booking
	err := utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		booking, err = s.bookingRepo.GetBookingByID(tx, bookingID)

		// In test mode, add a delay to increase deadlock chance
		if utils.DeadlockTesting.Enabled {
			utils.DelayIfTesting(150 * time.Millisecond)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	if s.logger != nil {
		s.logger.Printf("Updating booking %d for room %d", bookingID, booking.RoomNum)
	}

	// Only acquire application-level lock in normal mode
	if !utils.DeadlockTesting.Enabled {
		// Normal mode - use application lock
		unlock, err := s.lockManager.AcquireLock("room", booking.RoomNum)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		defer unlock()
	}

	// Execute update with retries for optimistic concurrency control
	return utils.RunWithRetry(3, func() error {
		return utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			// In test mode, add a delay inside transaction to increase chance of deadlocks
			if utils.DeadlockTesting.Enabled {
				utils.DelayIfTesting(200 * time.Millisecond)
			}

			// Calculate total price for new dates
			room, err := s.roomRepo.GetRoomByNumber(tx, booking.RoomNum)
			if err != nil {
				return fmt.Errorf("failed to find room: %w", err)
			}

			nights := int(newCheckOut.Sub(newCheckIn).Hours() / 24)
			if nights < 1 {
				return errors.New("booking must be for at least one night")
			}

			newTotalPrice := room.RoomType.PricePerNight * nights

			// DEADLOCK SCENARIO: In test mode, add another delay based on room number
			if utils.DeadlockTesting.Enabled && booking.RoomNum > 100 {
				utils.DelayIfTesting(150 * time.Millisecond)
			}

			// Check if the room is available for the new dates, excluding this booking's current dates
			var occupiedDaysCount int64
			err = tx.Model(&models.RoomStatus{}).
				Where("room_num = ? AND calendar >= ? AND calendar < ? AND status = ? AND (booking_id IS NULL OR booking_id != ?)",
					booking.RoomNum, newCheckIn.Format("2006-01-02"), newCheckOut.Format("2006-01-02"), "Occupied", bookingID).
				Count(&occupiedDaysCount).Error

			if err != nil {
				return fmt.Errorf("failed to check room availability: %w", err)
			}

			if occupiedDaysCount > 0 {
				return errors.New("room is not available for the new dates")
			}

			// First, update room statuses for the old dates to Available
			if err := tx.Model(&models.RoomStatus{}).
				Where("room_num = ? AND booking_id = ?", booking.RoomNum, bookingID).
				Updates(map[string]interface{}{
					"status":     "Available",
					"booking_id": nil,
				}).Error; err != nil {
				return fmt.Errorf("failed to update room statuses: %w", err)
			}

			// In test mode, add another delay to increase deadlock probability
			if utils.DeadlockTesting.Enabled {
				utils.DelayIfTesting(100 * time.Millisecond)
			}

			// Then update booking data
			updateData := map[string]interface{}{
				"check_in_date":  newCheckIn,
				"check_out_date": newCheckOut,
				"total_price":    newTotalPrice,
			}

			if err := s.bookingRepo.UpdateBooking(tx, bookingID, updateData); err != nil {
				return fmt.Errorf("failed to update booking: %w", err)
			}

			// Finally, create new room status entries for the new dates
			current := newCheckIn
			for current.Before(newCheckOut) {
				status := models.RoomStatus{
					RoomNum:   booking.RoomNum,
					Calendar:  current,
					Status:    "Occupied",
					BookingID: &bookingID,
				}

				if err := tx.Save(&status).Error; err != nil {
					return fmt.Errorf("failed to update room status: %w", err)
				}

				current = current.AddDate(0, 0, 1)
			}

			return nil
		})
	})
}

// GetBookingsByDateRange gets all bookings within a date range
func (s *BookingService) GetBookingsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Booking, error) {
	var bookings []models.Booking

	err := utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		// DEADLOCK SCENARIO 10: Add a delay to increase chance of transaction conflict
		time.Sleep(100 * time.Millisecond)

		return tx.
			Preload("Room.RoomType").
			Where("(check_in_date BETWEEN ? AND ?) OR (check_out_date BETWEEN ? AND ?)",
				startDate, endDate, startDate, endDate).
			Find(&bookings).Error
	})

	return bookings, err
}
