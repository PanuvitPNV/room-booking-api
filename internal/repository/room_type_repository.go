package repository

import (
	"context"
	"errors"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type roomTypeRepository struct {
	BaseRepository
}

func NewRoomTypeRepository(db *gorm.DB) RoomTypeRepository {
	return &roomTypeRepository{BaseRepository{db: db}}
}

func (r *roomTypeRepository) Create(ctx context.Context, roomType *models.RoomType) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check for duplicate name
		var count int64
		if err := db.Model(&models.RoomType{}).
			Where("name = ?", roomType.Name).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateEntry
		}

		// Create room type with locking
		return db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Create(roomType).Error
	})
}

func (r *roomTypeRepository) GetByID(ctx context.Context, typeID int) (*models.RoomType, error) {
	db := r.getDB(ctx)
	var roomType models.RoomType

	err := db.Preload("Rooms").
		First(&roomType, typeID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &roomType, nil
}

func (r *roomTypeRepository) Update(ctx context.Context, roomType *models.RoomType) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Lock the record for update
		existingRoomType := &models.RoomType{}
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(existingRoomType, roomType.TypeID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRecordNotFound
			}
			return err
		}

		// Check for duplicate name if name is being changed
		if roomType.Name != existingRoomType.Name {
			var count int64
			if err := db.Model(&models.RoomType{}).
				Where("name = ? AND type_id != ?", roomType.Name, roomType.TypeID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDuplicateEntry
			}
		}

		// Update room type
		return db.Save(roomType).Error
	})
}

func (r *roomTypeRepository) Delete(ctx context.Context, typeID int) error {
	return r.WithinTransaction(ctx, func(txCtx context.Context) error {
		db := r.getDB(txCtx)

		// Check if there are any rooms of this type
		var roomCount int64
		if err := db.Model(&models.Room{}).
			Where("type_id = ?", typeID).
			Count(&roomCount).Error; err != nil {
			return err
		}

		if roomCount > 0 {
			return errors.New("cannot delete room type with existing rooms")
		}

		// Delete room type
		result := db.Delete(&models.RoomType{}, typeID)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}

		return nil
	})
}

func (r *roomTypeRepository) List(ctx context.Context) ([]models.RoomType, error) {
	db := r.getDB(ctx)
	var roomTypes []models.RoomType

	err := db.Find(&roomTypes).Error
	return roomTypes, err
}

// Additional helper methods

func (r *roomTypeRepository) GetByName(ctx context.Context, name string) (*models.RoomType, error) {
	db := r.getDB(ctx)
	var roomType models.RoomType

	err := db.Where("name = ?", name).First(&roomType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &roomType, nil
}

func (r *roomTypeRepository) ListWithRooms(ctx context.Context) ([]models.RoomType, error) {
	db := r.getDB(ctx)
	var roomTypes []models.RoomType

	err := db.Preload("Rooms").Find(&roomTypes).Error
	return roomTypes, err
}

func (r *roomTypeRepository) GetAvailableTypes(ctx context.Context) ([]models.RoomType, error) {
	db := r.getDB(ctx)
	var roomTypes []models.RoomType

	err := db.Joins("JOIN rooms ON rooms.type_id = room_types.type_id").
		Group("room_types.type_id").
		Having("COUNT(rooms.room_num) > 0").
		Find(&roomTypes).Error

	return roomTypes, err
}

func (r *roomTypeRepository) GetTypeStatistics(ctx context.Context) (map[int]TypeStats, error) {
	db := r.getDB(ctx)

	type Result struct {
		TypeID    int
		RoomCount int64
		AvgPrice  float64
	}

	var results []Result
	err := db.Model(&models.RoomType{}).
		Select("room_types.type_id, COUNT(rooms.room_num) as room_count, AVG(room_types.price_per_night) as avg_price").
		Joins("LEFT JOIN rooms ON rooms.type_id = room_types.type_id").
		Group("room_types.type_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[int]TypeStats)
	for _, result := range results {
		stats[result.TypeID] = TypeStats{
			RoomCount: result.RoomCount,
			AvgPrice:  result.AvgPrice,
		}
	}

	return stats, nil
}

// Type for room type statistics
type TypeStats struct {
	RoomCount int64
	AvgPrice  float64
}
