package service

import (
	"context"

	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repository"
)

type roomTypeService struct {
	roomTypeRepo repository.RoomTypeRepository
}

func NewRoomTypeService(roomTypeRepo repository.RoomTypeRepository) RoomTypeService {
	return &roomTypeService{
		roomTypeRepo: roomTypeRepo,
	}
}

func (s *roomTypeService) CreateRoomType(ctx context.Context, req *request.CreateRoomTypeRequest) (*models.RoomType, error) {
	roomType := &models.RoomType{
		Name:          req.Name,
		Description:   req.Description,
		Area:          req.Area,
		Highlight:     req.Highlight,
		Facility:      req.Facility,
		PricePerNight: req.PricePerNight,
		Capacity:      req.Capacity,
	}

	err := s.roomTypeRepo.Create(ctx, roomType)
	if err != nil {
		return nil, err
	}

	return roomType, nil
}

func (s *roomTypeService) GetRoomTypeByID(ctx context.Context, typeID int) (*models.RoomType, error) {
	return s.roomTypeRepo.GetByID(ctx, typeID)
}

func (s *roomTypeService) UpdateRoomType(ctx context.Context, typeID int, req *request.UpdateRoomTypeRequest) (*models.RoomType, error) {
	roomType, err := s.roomTypeRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		roomType.Name = req.Name
	}
	if req.Description != "" {
		roomType.Description = req.Description
	}
	if req.Area != 0 {
		roomType.Area = req.Area
	}
	if req.Highlight != "" {
		roomType.Highlight = req.Highlight
	}
	if req.Facility != "" {
		roomType.Facility = req.Facility
	}
	if req.PricePerNight != 0 {
		roomType.PricePerNight = req.PricePerNight
	}
	if req.Capacity != 0 {
		roomType.Capacity = req.Capacity
	}

	err = s.roomTypeRepo.Update(ctx, roomType)
	if err != nil {
		return nil, err
	}

	return roomType, nil
}

func (s *roomTypeService) DeleteRoomType(ctx context.Context, typeID int) error {
	return s.roomTypeRepo.Delete(ctx, typeID)
}

func (s *roomTypeService) ListRoomTypes(ctx context.Context, page, pageSize int) ([]models.RoomType, int, error) {
	roomTypes, err := s.roomTypeRepo.List(ctx)
	if err != nil {
		return nil, 0, err
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	total := len(roomTypes)

	if start >= total {
		return []models.RoomType{}, total, nil
	}
	if end > total {
		end = total
	}

	return roomTypes[start:end], total, nil
}
