package repository

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type roomStatusRepository struct {
	BaseRepository
}

func NewRoomStatusRepository(db *gorm.DB) RoomStatusRepository {
	return &roomStatusRepository{BaseRepository{db: db}}
}

func (r *roomStatusRepository) Create(ctx context.Context, status *models.RoomStatus) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Verify room exists
		var room models.Room
		if err := db.First(&room, "room_num = ?", status.RoomNum).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid room number")
			}
			return err
		}

		// Verify booking exists if booking_id is provided
		if status.BookingID != nil {
			var booking models.Booking
			if err := db.First(&booking, *status.BookingID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("invalid booking ID")
				}
				return err
			}
		}

		// Create status with locking to prevent concurrent modifications
		return db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "room_num"},
				{Name: "calendar"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
		}).Create(status).Error
	})
}

func (r *roomStatusRepository) GetByRoomAndDate(ctx context.Context, roomNum int, date time.Time) (*models.RoomStatus, error) {
	db := r.getDB(ctx)
	var status models.RoomStatus

	err := db.Preload("Room.RoomType").
		Preload("Booking").
		Where("room_num = ? AND calendar = ?", roomNum, date).
		First(&status).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &status, nil
}

func (r *roomStatusRepository) Update(ctx context.Context, status *models.RoomStatus) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Lock the record for update
		existingStatus := &models.RoomStatus{}
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ? AND calendar = ?", status.RoomNum, status.Calendar).
			First(existingStatus).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRecordNotFound
			}
			return err
		}

		// Verify booking exists if booking_id is being updated
		if status.BookingID != nil {
			var booking models.Booking
			if err := db.First(&booking, *status.BookingID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("invalid booking ID")
				}
				return err
			}
		}

		// Update status
		return db.Save(status).Error
	})
}

func (r *roomStatusRepository) List(ctx context.Context, date time.Time) ([]models.RoomStatus, error) {
	db := r.getDB(ctx)
	var statuses []models.RoomStatus

	err := db.Preload("Room.RoomType").
		Preload("Booking").
		Where("calendar = ?", date).
		Find(&statuses).Error

	return statuses, err
}

func (r *roomStatusRepository) GetRoomStatusRange(ctx context.Context, roomNum int, startDate, endDate time.Time) ([]models.RoomStatus, error) {
	db := r.getDB(ctx)
	var statuses []models.RoomStatus

	err := db.Preload("Room.RoomType").
		Preload("Booking").
		Where("room_num = ? AND calendar BETWEEN ? AND ?", roomNum, startDate, endDate).
		Order("calendar ASC").
		Find(&statuses).Error

	return statuses, err
}

// Additional helper methods

func (r *roomStatusRepository) GetAllRoomStatusForDate(ctx context.Context, date time.Time) ([]models.RoomStatus, error) {
	db := r.getDB(ctx)
	var statuses []models.RoomStatus

	err := db.Preload("Room.RoomType").
		Preload("Booking").
		Where("calendar = ?", date).
		Order("room_num ASC").
		Find(&statuses).Error

	return statuses, err
}

func (r *roomStatusRepository) CreateBulkStatus(ctx context.Context, statuses []models.RoomStatus) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Create all statuses with conflict handling
		return db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "room_num"},
				{Name: "calendar"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
		}).Create(&statuses).Error
	})
}

func (r *roomStatusRepository) DeleteRoomStatusRange(ctx context.Context, roomNum int, startDate, endDate time.Time) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		return db.Where("room_num = ? AND calendar BETWEEN ? AND ?",
			roomNum, startDate, endDate).
			Delete(&models.RoomStatus{}).Error
	})
}

func (r *roomStatusRepository) GetOccupancyRate(ctx context.Context, startDate, endDate time.Time) (map[int]float64, error) {
	db := r.getDB(ctx)

	type Result struct {
		RoomNum  int
		Total    int64
		Occupied int64
	}

	var results []Result
	err := db.Model(&models.RoomStatus{}).
		Select("room_num, COUNT(*) as total, SUM(CASE WHEN status = 'Occupied' THEN 1 ELSE 0 END) as occupied").
		Where("calendar BETWEEN ? AND ?", startDate, endDate).
		Group("room_num").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	occupancyRates := make(map[int]float64)
	for _, result := range results {
		if result.Total > 0 {
			occupancyRates[result.RoomNum] = float64(result.Occupied) / float64(result.Total) * 100
		}
	}

	return occupancyRates, nil
}
