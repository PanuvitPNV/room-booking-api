package test

import (
	"context"
	"testing"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockBookingRepository is a mock for BookingRepository
type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepository) GetByID(ctx context.Context, bookingID int) (*models.Booking, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) CheckRoomAvailability(ctx context.Context, roomNum int, checkIn, checkOut time.Time) (bool, error) {
	args := m.Called(ctx, roomNum, checkIn, checkOut)
	return args.Bool(0), args.Error(1)
}

// Helper function to create test models
func CreateTestBooking() *models.Booking {
	return &models.Booking{
		BookingID:    1,
		RoomNum:      101,
		GuestID:      1,
		CheckInDate:  time.Now().Add(24 * time.Hour),
		CheckOutDate: time.Now().Add(72 * time.Hour),
		TotalPrice:   300,
	}
}

func CreateTestGuest() *models.Guest {
	return &models.Guest{
		GuestID:     1,
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Email:       "john@example.com",
		Phone:       "1234567890",
	}
}

func CreateTestRoom() *models.Room {
	return &models.Room{
		RoomNum:  101,
		TypeID:   1,
		RoomType: CreateTestRoomType(),
	}
}

func CreateTestRoomType() models.RoomType {
	return models.RoomType{
		TypeID:        1,
		Name:          "Deluxe",
		Description:   "Deluxe Room",
		Area:          30,
		PricePerNight: 100,
		Capacity:      2,
	}
}

// Assert functions
func AssertBookingEqual(t *testing.T, expected, actual *models.Booking) {
	t.Helper()
	if expected.BookingID != actual.BookingID {
		t.Errorf("BookingID mismatch: expected %d, got %d", expected.BookingID, actual.BookingID)
	}
	if expected.RoomNum != actual.RoomNum {
		t.Errorf("RoomNum mismatch: expected %d, got %d", expected.RoomNum, actual.RoomNum)
	}
	if expected.GuestID != actual.GuestID {
		t.Errorf("GuestID mismatch: expected %d, got %d", expected.GuestID, actual.GuestID)
	}
	// Add more field comparisons as needed
}
