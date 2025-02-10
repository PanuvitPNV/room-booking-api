package service

import (
	"context"
	"sync"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/dto/request"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repository"
)

type bookingService struct {
	bookingRepo    repository.BookingRepository
	roomRepo       repository.RoomRepository
	guestRepo      repository.GuestRepository
	roomStatusRepo repository.RoomStatusRepository
	roomLocks      sync.Map // Map of room numbers to mutex locks
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	roomRepo repository.RoomRepository,
	guestRepo repository.GuestRepository,
	roomStatusRepo repository.RoomStatusRepository,
) BookingService {
	return &bookingService{
		bookingRepo:    bookingRepo,
		roomRepo:       roomRepo,
		guestRepo:      guestRepo,
		roomStatusRepo: roomStatusRepo,
	}
}

func (s *bookingService) getRoomLock(roomNum int) *sync.Mutex {
	lock, _ := s.roomLocks.LoadOrStore(roomNum, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (s *bookingService) CreateBooking(ctx context.Context, req *request.CreateBookingRequest) (*models.Booking, error) {
	// Get room-specific lock for concurrency control
	lock := s.getRoomLock(req.RoomNum)
	lock.Lock()
	defer lock.Unlock()

	var booking *models.Booking
	err := s.bookingRepo.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Verify guest exists
		guest, err := s.guestRepo.GetByID(txCtx, req.GuestID)
		if err != nil {
			if err == repository.ErrRecordNotFound {
				return ErrGuestNotFound
			}
			return err
		}

		// 2. Verify room exists and get room details
		room, err := s.roomRepo.GetByNum(txCtx, req.RoomNum)
		if err != nil {
			if err == repository.ErrRecordNotFound {
				return ErrRoomNotFound
			}
			return err
		}

		// 3. Check room availability
		available, err := s.bookingRepo.CheckRoomAvailability(txCtx, req.RoomNum, req.CheckInDate, req.CheckOutDate)
		if err != nil {
			return err
		}
		if !available {
			return ErrRoomNotAvailable
		}

		// 4. Validate dates
		if req.CheckInDate.Before(time.Now()) {
			return ErrInvalidDateRange
		}
		if !req.CheckOutDate.After(req.CheckInDate) {
			return ErrInvalidDateRange
		}

		// 5. Calculate total price
		nights := int(req.CheckOutDate.Sub(req.CheckInDate).Hours() / 24)
		if nights < 1 {
			return ErrInvalidDateRange
		}
		totalPrice := room.RoomType.PricePerNight * nights

		// 6. Create booking
		booking = &models.Booking{
			RoomNum:      req.RoomNum,
			GuestID:      req.GuestID,
			CheckInDate:  req.CheckInDate,
			CheckOutDate: req.CheckOutDate,
			TotalPrice:   totalPrice,
			Room:         *room,
			Guest:        *guest,
		}

		// 7. Save booking - room status will be updated by trigger
		if err := s.bookingRepo.Create(txCtx, booking); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *bookingService) GetBookingByID(ctx context.Context, bookingID int) (*models.Booking, error) {
	return s.bookingRepo.GetByID(ctx, bookingID)
}

func (s *bookingService) UpdateBooking(ctx context.Context, bookingID int, req *request.UpdateBookingRequest) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, ErrBookingNotFound
	}

	lock := s.getRoomLock(booking.RoomNum)
	lock.Lock()
	defer lock.Unlock()

	err = s.bookingRepo.WithinTransaction(ctx, func(txCtx context.Context) error {
		if !req.CheckInDate.IsZero() {
			booking.CheckInDate = req.CheckInDate
		}
		if !req.CheckOutDate.IsZero() {
			booking.CheckOutDate = req.CheckOutDate
		}

		// Recalculate price if dates changed
		if !req.CheckInDate.IsZero() || !req.CheckOutDate.IsZero() {
			room, err := s.roomRepo.GetByNum(txCtx, booking.RoomNum)
			if err != nil {
				return err
			}
			nights := int(booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24)
			booking.TotalPrice = room.RoomType.PricePerNight * nights
		}

		return s.bookingRepo.Update(txCtx, booking)
	})

	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *bookingService) CancelBooking(ctx context.Context, bookingID int) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	lock := s.getRoomLock(booking.RoomNum)
	lock.Lock()
	defer lock.Unlock()

	return s.bookingRepo.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Update room status back to Available
		currentDate := booking.CheckInDate
		for currentDate.Before(booking.CheckOutDate) {
			status := &models.RoomStatus{
				RoomNum:  booking.RoomNum,
				Calendar: currentDate,
				Status:   "Available",
			}

			if err := s.roomStatusRepo.Create(txCtx, status); err != nil {
				return err
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}

		return s.bookingRepo.Delete(txCtx, bookingID)
	})
}

func (s *bookingService) ListBookings(ctx context.Context, req *request.GetBookingsRequest) ([]models.Booking, int, error) {
	var bookings []models.Booking
	var err error

	if req.GuestID != nil {
		bookings, err = s.bookingRepo.GetGuestBookings(ctx, *req.GuestID)
	} else if req.RoomNum != nil {
		bookings, err = s.bookingRepo.GetRoomBookings(ctx, *req.RoomNum)
	} else {
		bookings, err = s.bookingRepo.List(ctx)
	}

	if err != nil {
		return nil, 0, err
	}

	// Apply date filters if provided
	if req.FromDate != nil || req.ToDate != nil {
		var filtered []models.Booking
		for _, booking := range bookings {
			if req.FromDate != nil && booking.CheckInDate.Before(*req.FromDate) {
				continue
			}
			if req.ToDate != nil && booking.CheckOutDate.After(*req.ToDate) {
				continue
			}
			filtered = append(filtered, booking)
		}
		bookings = filtered
	}

	total := len(bookings)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start >= total {
		return []models.Booking{}, total, nil
	}
	if end > total {
		end = total
	}

	return bookings[start:end], total, nil
}

func (s *bookingService) CheckRoomAvailability(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, *models.Room, error) {
	if checkIn.After(checkOut) || checkIn.Equal(checkOut) {
		return false, nil, ErrInvalidDateRange
	}

	room, err := s.roomRepo.GetByNum(ctx, roomNum)
	if err != nil {
		return false, nil, ErrRoomNotFound
	}

	available, err := s.bookingRepo.CheckRoomAvailability(ctx, roomNum, checkIn, checkOut)
	if err != nil {
		return false, nil, err
	}

	return available, room, nil
}

func (s *bookingService) GetBookingsByGuest(ctx context.Context, guestID int) ([]models.Booking, error) {
	return s.bookingRepo.GetGuestBookings(ctx, guestID)
}

func (s *bookingService) GetBookingsByRoom(ctx context.Context, roomNum int) ([]models.Booking, error) {
	return s.bookingRepo.GetRoomBookings(ctx, roomNum)
}
