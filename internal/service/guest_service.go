package service

import (
	"context"

	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repository"
)

type guestService struct {
	guestRepo   repository.GuestRepository
	bookingRepo repository.BookingRepository
}

func NewGuestService(guestRepo repository.GuestRepository, bookingRepo repository.BookingRepository) GuestService {
	return &guestService{
		guestRepo:   guestRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *guestService) CreateGuest(ctx context.Context, req *request.CreateGuestRequest) (*models.Guest, error) {
	guest := &models.Guest{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Email:       req.Email,
		Phone:       req.Phone,
	}

	err := s.guestRepo.Create(ctx, guest)
	if err != nil {
		return nil, err
	}

	return guest, nil
}

func (s *guestService) GetGuestByID(ctx context.Context, guestID int) (*models.Guest, error) {
	return s.guestRepo.GetByID(ctx, guestID)
}

func (s *guestService) UpdateGuest(ctx context.Context, guestID int, req *request.UpdateGuestRequest) (*models.Guest, error) {
	guest, err := s.guestRepo.GetByID(ctx, guestID)
	if err != nil {
		return nil, ErrGuestNotFound
	}

	if req.FirstName != "" {
		guest.FirstName = req.FirstName
	}
	if req.LastName != "" {
		guest.LastName = req.LastName
	}
	if !req.DateOfBirth.IsZero() {
		guest.DateOfBirth = req.DateOfBirth
	}
	if req.Email != "" {
		guest.Email = req.Email
	}
	if req.Phone != "" {
		guest.Phone = req.Phone
	}

	err = s.guestRepo.Update(ctx, guest)
	if err != nil {
		return nil, err
	}

	return guest, nil
}

func (s *guestService) DeleteGuest(ctx context.Context, guestID int) error {
	return s.guestRepo.Delete(ctx, guestID)
}

func (s *guestService) ListGuests(ctx context.Context, page, pageSize int) ([]models.Guest, int, error) {
	guests, err := s.guestRepo.List(ctx)
	if err != nil {
		return nil, 0, err
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	total := len(guests)

	if start >= total {
		return []models.Guest{}, total, nil
	}
	if end > total {
		end = total
	}

	return guests[start:end], total, nil
}

func (s *guestService) GetGuestBookingHistory(ctx context.Context, guestID int) ([]models.Booking, error) {
	// Verify guest exists
	_, err := s.guestRepo.GetByID(ctx, guestID)
	if err != nil {
		return nil, ErrGuestNotFound
	}

	return s.bookingRepo.GetGuestBookings(ctx, guestID)
}
