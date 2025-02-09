package repository

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type roomRepository struct {
	BaseRepository
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{BaseRepository{db: db}}
}

func (r *roomRepository) Create(ctx context.Context, room *models.Room) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check if room number already exists
		var count int64
		if err := db.Model(&models.Room{}).
			Where("room_num = ?", room.RoomNum).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateEntry
		}

		// Verify room type exists
		var roomType models.RoomType
		if err := db.First(&roomType, room.TypeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid room type")
			}
			return err
		}

		// Create room with locking
		return db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Create(room).Error
	})
}

func (r *roomRepository) GetByNum(ctx context.Context, roomNum int) (*models.Room, error) {
	db := r.getDB(ctx)
	var room models.Room

	err := db.Preload("RoomType").
		First(&room, "room_num = ?", roomNum).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &room, nil
}

func (r *roomRepository) Update(ctx context.Context, room *models.Room) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Lock the record for update
		existingRoom := &models.Room{}
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(existingRoom, "room_num = ?", room.RoomNum).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRecordNotFound
			}
			return err
		}

		// Verify room type exists if it's being changed
		if room.TypeID != existingRoom.TypeID {
			var roomType models.RoomType
			if err := db.First(&roomType, room.TypeID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("invalid room type")
				}
				return err
			}
		}

		// Update room
		return db.Save(room).Error
	})
}

func (r *roomRepository) Delete(ctx context.Context, roomNum int) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check if room has any bookings
		var bookingCount int64
		if err := db.Model(&models.Booking{}).
			Where("room_num = ?", roomNum).
			Count(&bookingCount).Error; err != nil {
			return err
		}

		if bookingCount > 0 {
			return errors.New("cannot delete room with existing bookings")
		}

		// Delete room
		result := db.Delete(&models.Room{}, "room_num = ?", roomNum)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}

		return nil
	})
}

func (r *roomRepository) List(ctx context.Context) ([]models.Room, error) {
	db := r.getDB(ctx)
	var rooms []models.Room

	err := db.Preload("RoomType").Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) GetByType(ctx context.Context, typeID int) ([]models.Room, error) {
	db := r.getDB(ctx)
	var rooms []models.Room

	err := db.Preload("RoomType").
		Where("type_id = ?", typeID).
		Find(&rooms).Error

	return rooms, err
}

func (r *roomRepository) GetAvailableRooms(ctx context.Context, checkIn, checkOut time.Time) ([]models.Room, error) {
	db := r.getDB(ctx)
	var rooms []models.Room

	err := db.Preload("RoomType").
		Where("room_num NOT IN (?)",
			db.Table("bookings").
				Select("room_num").
				Where("(check_in_date < ? AND check_out_date > ?) OR "+
					"(check_in_date BETWEEN ? AND ?) OR "+
					"(check_out_date BETWEEN ? AND ?)",
					checkOut, checkIn,
					checkIn, checkOut,
					checkIn, checkOut)).
		Find(&rooms).Error

	return rooms, err
}

func (r *roomRepository) GetAvailableRoomsByType(ctx context.Context, typeID int, checkIn, checkOut time.Time) ([]models.Room, error) {
	db := r.getDB(ctx)
	var rooms []models.Room

	err := db.Preload("RoomType").
		Where("type_id = ?", typeID).
		Where("room_num NOT IN (?)",
			db.Table("bookings").
				Select("room_num").
				Where("(check_in_date < ? AND check_out_date > ?) OR "+
					"(check_in_date BETWEEN ? AND ?) OR "+
					"(check_out_date BETWEEN ? AND ?)",
					checkOut, checkIn,
					checkIn, checkOut,
					checkIn, checkOut)).
		Find(&rooms).Error

	return rooms, err
}

// Helper method to check if a room is available for given dates
func (r *roomRepository) isRoomAvailable(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, error) {
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

	return count == 0, err
}
