package services

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"gorm.io/gorm"
)

// BookingService handles business logic for bookings
type BookingService struct {
	db          *gorm.DB
	bookingRepo *repositories.BookingRepository
}

// NewBookingService creates a new BookingService
func NewBookingService(db *gorm.DB, bookingRepo *repositories.BookingRepository) *BookingService {
	return &BookingService{
		db:          db,
		bookingRepo: bookingRepo,
	}
}

// BookingRequest represents a request to create or update a booking
type BookingRequest struct {
	BookingName  string    `json:"booking_name" validate:"required"`
	RoomNum      int       `json:"room_num" validate:"required"`
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// PaymentRequest represents a payment for a booking
type PaymentRequest struct {
	PaymentMethod string    `json:"payment_method" validate:"required,oneof=Credit Debit Bank Transfer"`
	PaymentDate   time.Time `json:"payment_date" validate:"required"`
}

// CheckRoomAvailability checks if a room is available for the requested dates
func (s *BookingService) CheckRoomAvailability(ctx context.Context, tx *gorm.DB, roomNum int, checkIn, checkOut time.Time) (bool, error) {
	return s.bookingRepo.CheckRoomAvailability(ctx, tx, roomNum, checkIn, checkOut)
}

// GetBooking retrieves a booking by ID
func (s *BookingService) GetBooking(tx *gorm.DB, bookingID int) (*models.Booking, error) {
	return s.bookingRepo.GetBookingByID(tx, bookingID)
}

// UpdateBooking updates an existing booking
func (s *BookingService) UpdateBooking(ctx context.Context, tx *gorm.DB, bookingID int, req BookingRequest) (*models.Booking, error) {
	// Basic validation
	if req.CheckInDate.After(req.CheckOutDate) || req.CheckInDate.Equal(req.CheckOutDate) {
		return nil, errors.New("check-in date must be before check-out date")
	}

	if req.CheckInDate.Before(time.Now()) {
		return nil, errors.New("cannot update booking to start in the past")
	}

	// Get existing booking
	existingBooking, err := s.bookingRepo.GetBookingByID(tx, bookingID)
	if err != nil {
		return nil, err
	}

	// Update fields
	existingBooking.BookingName = req.BookingName
	existingBooking.RoomNum = req.RoomNum
	existingBooking.CheckInDate = req.CheckInDate
	existingBooking.CheckOutDate = req.CheckOutDate

	// Use the repository to handle database operations with concurrency control
	err = s.bookingRepo.UpdateBooking(ctx, tx, existingBooking)
	if err != nil {
		return nil, err
	}

	// Retrieve the updated booking
	return s.bookingRepo.GetBookingByID(tx, bookingID)
}

// CancelBooking cancels a booking
func (s *BookingService) CancelBooking(ctx context.Context, tx *gorm.DB, bookingID int) error {
	return s.bookingRepo.CancelBooking(ctx, tx, bookingID)
}

// SearchAvailableRooms finds available rooms for a specific date range and guest count
func (s *BookingService) SearchAvailableRooms(tx *gorm.DB, checkInDate, checkOutDate time.Time, guestCount int) ([]models.Room, error) {
	// Basic validation
	if checkInDate.After(checkOutDate) || checkInDate.Equal(checkOutDate) {
		return nil, errors.New("check-in date must be before check-out date")
	}

	return s.bookingRepo.SearchAvailableRooms(tx, checkInDate, checkOutDate, guestCount)
}

// CreateBookingWithPayment creates a booking and processes payment in a single atomic transaction
// Implements "first to pay gets the room" business logic
func (s *BookingService) CreateBookingWithPayment(
	ctx context.Context,
	bookingReq BookingRequest,
	paymentReq PaymentRequest,
) (*models.Booking, *models.Receipt, error) {
	// Validate booking request
	if bookingReq.CheckInDate.After(bookingReq.CheckOutDate) || bookingReq.CheckInDate.Equal(bookingReq.CheckOutDate) {
		return nil, nil, errors.New("check-in date must be before check-out date")
	}

	if bookingReq.CheckInDate.Before(time.Now()) {
		return nil, nil, errors.New("cannot book a room in the past")
	}

	var booking *models.Booking
	var receipt *models.Receipt

	// Execute the entire operation in a single transaction to ensure atomicity
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// First check room availability with locking to prevent race conditions
		available, err := s.bookingRepo.CheckRoomAvailability(
			ctx,
			tx,
			bookingReq.RoomNum,
			bookingReq.CheckInDate,
			bookingReq.CheckOutDate,
		)
		if err != nil {
			return err
		}

		if !available {
			return errors.New("room is not available for the selected dates")
		}

		// Create booking model
		booking = &models.Booking{
			BookingName:  bookingReq.BookingName,
			RoomNum:      bookingReq.RoomNum,
			CheckInDate:  bookingReq.CheckInDate,
			CheckOutDate: bookingReq.CheckOutDate,
			BookingDate:  time.Now(),
		}

		// Create receipt model
		receipt = &models.Receipt{
			PaymentDate:   paymentReq.PaymentDate,
			PaymentMethod: paymentReq.PaymentMethod,
			IssueDate:     time.Now(),
		}

		// Create booking with payment in a single operation
		// This ensures that payment and booking are created atomically
		if err := s.bookingRepo.CreateBooking(ctx, tx, booking, receipt); err != nil {
			return err
		}

		// Get the complete booking with all relationships
		completeBooking, err := s.bookingRepo.GetBookingByID(tx, booking.BookingID)
		if err != nil {
			return err
		}

		// Update our reference to include all loaded relationships
		*booking = *completeBooking

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return booking, receipt, nil
}
