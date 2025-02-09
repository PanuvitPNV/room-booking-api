// internal/repository/guest_repository.go
package repository

import (
	"context"
	"errors"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type guestRepository struct {
	BaseRepository
}

func NewGuestRepository(db *gorm.DB) GuestRepository {
	return &guestRepository{BaseRepository{db: db}}
}

func (r *guestRepository) Create(ctx context.Context, guest *models.Guest) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check for duplicate email
		var count int64
		if err := db.Model(&models.Guest{}).
			Where("email = ?", guest.Email).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateEntry
		}

		// Check for duplicate phone
		if err := db.Model(&models.Guest{}).
			Where("phone = ?", guest.Phone).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateEntry
		}

		// Create guest with locking to prevent race conditions
		return db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Create(guest).Error
	})
}

func (r *guestRepository) GetByID(ctx context.Context, guestID int) (*models.Guest, error) {
	db := r.getDB(ctx)
	var guest models.Guest

	err := db.Preload("Bookings.Room.RoomType").
		First(&guest, guestID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &guest, nil
}

func (r *guestRepository) Update(ctx context.Context, guest *models.Guest) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Lock the record for update
		existingGuest := &models.Guest{}
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(existingGuest, guest.GuestID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRecordNotFound
			}
			return err
		}

		// Check for duplicate email if email is being changed
		if guest.Email != existingGuest.Email {
			var count int64
			if err := db.Model(&models.Guest{}).
				Where("email = ? AND guest_id != ?", guest.Email, guest.GuestID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDuplicateEntry
			}
		}

		// Check for duplicate phone if phone is being changed
		if guest.Phone != existingGuest.Phone {
			var count int64
			if err := db.Model(&models.Guest{}).
				Where("phone = ? AND guest_id != ?", guest.Phone, guest.GuestID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDuplicateEntry
			}
		}

		// Update guest
		return db.Save(guest).Error
	})
}

func (r *guestRepository) Delete(ctx context.Context, guestID int) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check if guest has any active bookings
		var bookingCount int64
		if err := db.Model(&models.Booking{}).
			Where("guest_id = ?", guestID).
			Count(&bookingCount).Error; err != nil {
			return err
		}

		if bookingCount > 0 {
			return errors.New("cannot delete guest with existing bookings")
		}

		// Delete guest
		result := db.Delete(&models.Guest{}, guestID)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}

		return nil
	})
}

func (r *guestRepository) List(ctx context.Context) ([]models.Guest, error) {
	db := r.getDB(ctx)
	var guests []models.Guest

	err := db.Find(&guests).Error
	if err != nil {
		return nil, err
	}

	return guests, nil
}

// Additional helper methods

func (r *guestRepository) GetByEmail(ctx context.Context, email string) (*models.Guest, error) {
	db := r.getDB(ctx)
	var guest models.Guest

	err := db.Where("email = ?", email).First(&guest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &guest, nil
}

func (r *guestRepository) GetByPhone(ctx context.Context, phone string) (*models.Guest, error) {
	db := r.getDB(ctx)
	var guest models.Guest

	err := db.Where("phone = ?", phone).First(&guest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &guest, nil
}
