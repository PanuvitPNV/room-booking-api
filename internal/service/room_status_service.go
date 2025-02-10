package service

import (
	"context"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repository"
)

type roomStatusService struct {
	roomStatusRepo repository.RoomStatusRepository
	roomRepo       repository.RoomRepository
}

func NewRoomStatusService(
	roomStatusRepo repository.RoomStatusRepository,
	roomRepo repository.RoomRepository,
) RoomStatusService {
	return &roomStatusService{
		roomStatusRepo: roomStatusRepo,
		roomRepo:       roomRepo,
	}
}

func (s *roomStatusService) GetRoomStatus(ctx context.Context, roomNum int, date time.Time) (*models.RoomStatus, error) {
	// Verify room exists
	_, err := s.roomRepo.GetByNum(ctx, roomNum)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	return s.roomStatusRepo.GetByRoomAndDate(ctx, roomNum, date)
}

func (s *roomStatusService) UpdateRoomStatus(ctx context.Context, status *models.RoomStatus) error {
	// Verify room exists
	_, err := s.roomRepo.GetByNum(ctx, status.RoomNum)
	if err != nil {
		return ErrRoomNotFound
	}

	// Validate status value using constants
	if status.Status != StatusAvailable && status.Status != StatusOccupied {
		return ErrInvalidData
	}

	return s.roomStatusRepo.Update(ctx, status)
}

func (s *roomStatusService) GetRoomStatusRange(ctx context.Context, roomNum int, startDate, endDate time.Time) ([]models.RoomStatus, error) {
	// Verify room exists
	_, err := s.roomRepo.GetByNum(ctx, roomNum)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	// Validate date range
	if startDate.After(endDate) {
		return nil, ErrInvalidDateRange
	}

	return s.roomStatusRepo.GetRoomStatusRange(ctx, roomNum, startDate, endDate)
}

func (s *roomStatusService) GetAllRoomStatus(ctx context.Context, date time.Time) ([]models.RoomStatus, error) {
	return s.roomStatusRepo.List(ctx, date)
}

// Additional helper methods for internal use
func (s *roomStatusService) updateRoomStatusForBooking(ctx context.Context, booking *models.Booking, status string) error {
	currentDate := booking.CheckInDate
	for currentDate.Before(booking.CheckOutDate) {
		roomStatus := &models.RoomStatus{
			RoomNum:   booking.RoomNum,
			Calendar:  currentDate,
			Status:    status,
			BookingID: &booking.BookingID,
		}

		if err := s.roomStatusRepo.Create(ctx, roomStatus); err != nil {
			return err
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}
	return nil
}

func (s *roomStatusService) clearRoomStatus(ctx context.Context, roomNum int, startDate, endDate time.Time) error {
	currentDate := startDate
	for currentDate.Before(endDate) {
		roomStatus := &models.RoomStatus{
			RoomNum:  roomNum,
			Calendar: currentDate,
			Status:   StatusAvailable, // Using constant
		}

		if err := s.roomStatusRepo.Create(ctx, roomStatus); err != nil {
			return err
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}
	return nil
}
