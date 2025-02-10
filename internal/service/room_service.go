package service

import (
	"context"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repository"
)

type roomService struct {
	roomRepo     repository.RoomRepository
	roomTypeRepo repository.RoomTypeRepository
}

func NewRoomService(roomRepo repository.RoomRepository, roomTypeRepo repository.RoomTypeRepository) RoomService {
	return &roomService{
		roomRepo:     roomRepo,
		roomTypeRepo: roomTypeRepo,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, req *request.CreateRoomRequest) (*models.Room, error) {
	// Verify room type exists
	_, err := s.roomTypeRepo.GetByID(ctx, req.TypeID)
	if err != nil {
		return nil, ErrRoomTypeNotFound
	}

	room := &models.Room{
		RoomNum: req.RoomNum,
		TypeID:  req.TypeID,
	}

	err = s.roomRepo.Create(ctx, room)
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *roomService) GetRoomByNum(ctx context.Context, roomNum int) (*models.Room, error) {
	return s.roomRepo.GetByNum(ctx, roomNum)
}

func (s *roomService) UpdateRoom(ctx context.Context, roomNum int, req *request.UpdateRoomRequest) (*models.Room, error) {
	// Verify room exists
	room, err := s.roomRepo.GetByNum(ctx, roomNum)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	// Verify new room type exists if changing
	if req.TypeID != room.TypeID {
		_, err := s.roomTypeRepo.GetByID(ctx, req.TypeID)
		if err != nil {
			return nil, ErrRoomTypeNotFound
		}
		room.TypeID = req.TypeID
	}

	err = s.roomRepo.Update(ctx, room)
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *roomService) DeleteRoom(ctx context.Context, roomNum int) error {
	return s.roomRepo.Delete(ctx, roomNum)
}

func (s *roomService) ListRooms(ctx context.Context, page, pageSize int) ([]models.Room, int, error) {
	rooms, err := s.roomRepo.List(ctx)
	if err != nil {
		return nil, 0, err
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	total := len(rooms)

	if start >= total {
		return []models.Room{}, total, nil
	}
	if end > total {
		end = total
	}

	return rooms[start:end], total, nil
}

func (s *roomService) GetRoomsByType(ctx context.Context, typeID int) ([]models.Room, error) {
	return s.roomRepo.GetByType(ctx, typeID)
}

func (s *roomService) GetAvailableRooms(ctx context.Context, checkIn, checkOut time.Time, typeID *int) ([]models.Room, error) {
	if checkIn.After(checkOut) || checkIn.Equal(checkOut) {
		return nil, ErrInvalidDateRange
	}

	if typeID != nil {
		return s.roomRepo.GetAvailableRoomsByType(ctx, *typeID, checkIn, checkOut)
	}
	return s.roomRepo.GetAvailableRooms(ctx, checkIn, checkOut)
}
