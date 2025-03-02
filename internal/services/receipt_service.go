package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"github.com/panuvitpnv/room-booking-api/internal/utils"
)

// ReceiptService handles receipt business logic
type ReceiptService struct {
	receiptRepo *repositories.ReceiptRepository
	bookingRepo *repositories.BookingRepository
	lockManager *utils.LockManager
}

// NewReceiptService creates a new receipt service
func NewReceiptService(
	receiptRepo *repositories.ReceiptRepository,
	bookingRepo *repositories.BookingRepository,
	lockManager *utils.LockManager,
) *ReceiptService {
	return &ReceiptService{
		receiptRepo: receiptRepo,
		bookingRepo: bookingRepo,
		lockManager: lockManager,
	}
}

// CreateReceipt creates a new payment receipt with transaction management
func (s *ReceiptService) CreateReceipt(ctx context.Context, receipt *models.Receipt) error {
	// Validate receipt data
	if receipt.BookingID <= 0 {
		return errors.New("booking ID is required")
	}

	if receipt.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if receipt.PaymentMethod == "" {
		return errors.New("payment method is required")
	}

	// Get booking first to verify it exists and to lock on it
	var booking *models.Booking
	err := utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		booking, err = s.bookingRepo.GetBookingByID(tx, receipt.BookingID)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	// Acquire a lock on the booking to prevent race conditions
	// This ensures only one payment can be processed for a booking at a time
	unlock, err := s.lockManager.AcquireLock("booking", receipt.BookingID)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer unlock()

	// Execute receipt creation with retry for optimistic concurrency
	return utils.RunWithRetry(3, func() error {
		return utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			// Verify that amount matches the booking total
			if receipt.Amount != booking.TotalPrice {
				return fmt.Errorf("payment amount (%d) does not match booking total (%d)",
					receipt.Amount, booking.TotalPrice)
			}

			// Create the receipt
			if err := s.receiptRepo.CreateReceipt(tx, receipt); err != nil {
				return fmt.Errorf("failed to create receipt: %w", err)
			}

			return nil
		})
	})
}

// GetReceiptByID retrieves a receipt by ID
func (s *ReceiptService) GetReceiptByID(ctx context.Context, receiptID int) (*models.Receipt, error) {
	var receipt *models.Receipt
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		receipt, err = s.receiptRepo.GetReceiptByID(tx, receiptID)
		return err
	})

	return receipt, err
}

// GetReceiptByBookingID retrieves a receipt by booking ID
func (s *ReceiptService) GetReceiptByBookingID(ctx context.Context, bookingID int) (*models.Receipt, error) {
	var receipt *models.Receipt
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		receipt, err = s.receiptRepo.GetReceiptByBookingID(tx, bookingID)
		return err
	})

	return receipt, err
}

// ProcessRefund processes a refund for a booking
func (s *ReceiptService) ProcessRefund(ctx context.Context, bookingID int) error {
	// First check if there's a receipt for this booking
	var receipt *models.Receipt
	err := utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		var err error
		receipt, err = s.receiptRepo.GetReceiptByBookingID(tx, bookingID)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to find receipt for booking %d: %w", bookingID, err)
	}

	// Acquire locks for both the booking and the receipt
	resources := []struct {
		Type string
		ID   interface{}
	}{
		{Type: "booking", ID: bookingID},
		{Type: "receipt", ID: receipt.ReceiptID},
	}

	unlock, err := s.lockManager.AcquireMultipleLocks(resources)
	if err != nil {
		return fmt.Errorf("failed to acquire locks: %w", err)
	}
	defer unlock()

	// Execute refund with retry for optimistic concurrency
	return utils.RunWithRetry(3, func() error {
		return utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			// Cancel the receipt (create a refund record)
			if err := s.receiptRepo.CancelReceipt(tx, receipt.ReceiptID); err != nil {
				return fmt.Errorf("failed to create refund receipt: %w", err)
			}

			// Cancel the booking
			if err := s.bookingRepo.CancelBooking(tx, bookingID); err != nil {
				return fmt.Errorf("failed to cancel booking: %w", err)
			}

			return nil
		})
	})
}

// GetAllReceipts gets all receipts with pagination
func (s *ReceiptService) GetAllReceipts(ctx context.Context, page, pageSize int) ([]models.Receipt, int64, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	var receipts []models.Receipt
	var total int64
	var err error

	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		receipts, total, err = s.receiptRepo.GetAllReceipts(tx, page, pageSize)
		return err
	})

	return receipts, total, err
}

// GetReceiptsByDateRange gets all receipts within a date range
func (s *ReceiptService) GetReceiptsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Receipt, error) {
	var receipts []models.Receipt

	err := utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		return tx.
			Preload("Booking").
			Where("payment_date BETWEEN ? AND ?", startDate, endDate).
			Find(&receipts).Error
	})

	return receipts, err
}
