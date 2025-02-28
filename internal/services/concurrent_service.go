package services

import (
	"context"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"gorm.io/gorm"
)

// ConcurrentService handles concurrency demonstration scenarios
type ConcurrentService struct {
	db                  *gorm.DB
	concurrentScenarios *repositories.ConcurrentScenarios
}

// NewConcurrentService creates a new ConcurrentService
func NewConcurrentService(db *gorm.DB, concurrentScenarios *repositories.ConcurrentScenarios) *ConcurrentService {
	return &ConcurrentService{
		db:                  db,
		concurrentScenarios: concurrentScenarios,
	}
}

// DemoLostUpdate demonstrates the lost update problem
func (s *ConcurrentService) DemoLostUpdate(ctx context.Context, bookingID int) (string, error) {
	return s.concurrentScenarios.LostUpdate(ctx, bookingID)
}

// DemoLostUpdateWithLocking demonstrates prevention of the lost update problem using pessimistic locking
func (s *ConcurrentService) DemoLostUpdateWithLocking(ctx context.Context, bookingID int) (string, error) {
	return s.concurrentScenarios.LostUpdateWithPessimisticLocking(ctx, bookingID)
}

// DemoDirtyRead demonstrates the dirty read problem
func (s *ConcurrentService) DemoDirtyRead(ctx context.Context, bookingID int) (string, error) {
	return s.concurrentScenarios.DirtyRead(ctx, bookingID)
}

// DemoPhantomRead demonstrates the phantom read problem
func (s *ConcurrentService) DemoPhantomRead(ctx context.Context, checkInStr, checkOutStr string) (string, error) {
	// Parse date strings
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return "", err
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return "", err
	}

	return s.concurrentScenarios.PhantomRead(ctx, checkIn, checkOut)
}

// DemoSerializationAnomaly demonstrates a serialization anomaly
func (s *ConcurrentService) DemoSerializationAnomaly(ctx context.Context) (string, error) {
	return s.concurrentScenarios.SerializationAnomaly(ctx)
}

// DemoConcurrentBookings demonstrates concurrent booking attempts for the same room
func (s *ConcurrentService) DemoConcurrentBookings(ctx context.Context, roomNum int, checkInStr, checkOutStr string) (string, error) {
	// Parse date strings
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return "", err
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return "", err
	}

	return s.concurrentScenarios.ConcurrentBookings(ctx, roomNum, checkIn, checkOut)
}
