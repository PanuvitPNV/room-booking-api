package repository

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type bookingRepository struct {
	BaseRepository
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{BaseRepository{db: db}}
}

func (r *bookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check room availability with lock
		available, err := r.checkRoomAvailabilityWithLock(txCtx, booking.RoomNum, booking.CheckInDate, booking.CheckOutDate)
		if err != nil {
			return err
		}
		if !available {
			return errors.New("room is not available for the selected dates")
		}

		// Create booking
		if err := db.Create(booking).Error; err != nil {
			return err
		}

		// Update room status
		return r.updateRoomStatus(txCtx, booking)
	})
}

func (r *bookingRepository) GetByID(ctx context.Context, bookingID int) (*models.Booking, error) {
	db := r.getDB(ctx)
	var booking models.Booking

	err := db.Preload("Room.RoomType").
		Preload("Guest").
		First(&booking, bookingID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &booking, nil
}

func (r *bookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Lock the existing booking record
		existingBooking := &models.Booking{}
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(existingBooking, booking.BookingID).Error
		if err != nil {
			return err
		}

		// Check availability for new dates if they changed
		if !existingBooking.CheckInDate.Equal(booking.CheckInDate) ||
			!existingBooking.CheckOutDate.Equal(booking.CheckOutDate) {

			available, err := r.checkRoomAvailabilityWithLock(txCtx, booking.RoomNum, booking.CheckInDate, booking.CheckOutDate)
			if err != nil {
				return err
			}
			if !available {
				return errors.New("room is not available for the new dates")
			}

			// Delete old room statuses
			if err := db.Where("booking_id = ?", booking.BookingID).
				Delete(&models.RoomStatus{}).Error; err != nil {
				return err
			}

			// Create new room statuses
			if err := r.updateRoomStatus(txCtx, booking); err != nil {
				return err
			}
		}

		// Update booking
		return db.Save(booking).Error
	})
}

func (r *bookingRepository) Delete(ctx context.Context, bookingID int) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Delete associated room statuses first
		if err := db.Where("booking_id = ?", bookingID).
			Delete(&models.RoomStatus{}).Error; err != nil {
			return err
		}

		// Delete the booking
		return db.Delete(&models.Booking{}, bookingID).Error
	})
}

func (r *bookingRepository) List(ctx context.Context) ([]models.Booking, error) {
	db := r.getDB(ctx)
	var bookings []models.Booking

	err := db.Preload("Room.RoomType").
		Preload("Guest").
		Find(&bookings).Error

	return bookings, err
}

func (r *bookingRepository) GetGuestBookings(ctx context.Context, guestID int) ([]models.Booking, error) {
	db := r.getDB(ctx)
	var bookings []models.Booking

	err := db.Preload("Room.RoomType").
		Preload("Guest").
		Where("guest_id = ?", guestID).
		Find(&bookings).Error

	return bookings, err
}

func (r *bookingRepository) GetRoomBookings(ctx context.Context, roomNum int) ([]models.Booking, error) {
	db := r.getDB(ctx)
	var bookings []models.Booking

	err := db.Preload("Guest").
		Preload("Room.RoomType").
		Where("room_num = ?", roomNum).
		Order("check_in_date ASC").
		Find(&bookings).Error

	return bookings, err
}

func (r *bookingRepository) CheckRoomAvailability(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, error) {
	db := r.getDB(ctx)
	var count int64

	err := db.Model(&models.Booking{}).
		Where("room_num = ? AND "+
			"((check_in_date < ? AND check_out_date > ?) OR "+
			"(check_in_date BETWEEN ? AND ?) OR "+
			"(check_out_date BETWEEN ? AND ?))",
			roomNum,
			checkOut, checkIn,
			checkIn, checkOut,
			checkIn, checkOut).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// Private helper methods

func (r *bookingRepository) checkRoomAvailabilityWithLock(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, error) {
	db := r.getDB(ctx)

	var count int64
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&models.Booking{}).
		Where("room_num = ? AND "+
			"((check_in_date < ? AND check_out_date > ?) OR "+
			"(check_in_date BETWEEN ? AND ?) OR "+
			"(check_out_date BETWEEN ? AND ?))",
			roomNum,
			checkOut, checkIn,
			checkIn, checkOut,
			checkIn, checkOut).
		Count(&count).Error

	return count == 0, err
}

func (r *bookingRepository) updateRoomStatus(ctx context.Context, booking *models.Booking) error {
	db := r.getDB(ctx)

	// Generate dates between check-in and check-out
	var statuses []models.RoomStatus
	for d := booking.CheckInDate; d.Before(booking.CheckOutDate); d = d.AddDate(0, 0, 1) {
		statuses = append(statuses, models.RoomStatus{
			RoomNum:   booking.RoomNum,
			Calendar:  d,
			Status:    "Occupied",
			BookingID: &booking.BookingID,
		})
	}

	// Use upsert to handle existing records
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "room_num"}, {Name: "calendar"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
	}).Create(&statuses).Error
}
