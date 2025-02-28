package services

import (
	"context"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"gorm.io/gorm"
)

// ReceiptService handles business logic for receipts
type ReceiptService struct {
	db          *gorm.DB
	receiptRepo *repositories.ReceiptRepository
}

// NewReceiptService creates a new ReceiptService
func NewReceiptService(db *gorm.DB, receiptRepo *repositories.ReceiptRepository) *ReceiptService {
	return &ReceiptService{
		db:          db,
		receiptRepo: receiptRepo,
	}
}

// GetReceiptByID retrieves a receipt by ID
func (s *ReceiptService) GetReceiptByID(tx *gorm.DB, receiptID int) (*models.Receipt, error) {
	return s.receiptRepo.GetReceiptByID(tx, receiptID)
}

// GetReceiptByBookingID retrieves a receipt by booking ID
func (s *ReceiptService) GetReceiptByBookingID(tx *gorm.DB, bookingID int) (*models.Receipt, error) {
	return s.receiptRepo.GetReceiptByBookingID(tx, bookingID)
}

// CreateReceipt creates a new receipt for a booking
func (s *ReceiptService) CreateReceipt(ctx context.Context, tx *gorm.DB, receipt *models.Receipt) (*models.Receipt, error) {
	// Create receipt
	if err := s.receiptRepo.CreateReceipt(ctx, tx, receipt); err != nil {
		return nil, err
	}

	// Retrieve the created receipt
	return s.receiptRepo.GetReceiptByID(tx, receipt.ReceiptID)
}

// UpdateReceipt updates an existing receipt
func (s *ReceiptService) UpdateReceipt(ctx context.Context, tx *gorm.DB, receipt *models.Receipt) (*models.Receipt, error) {
	// Update receipt
	if err := s.receiptRepo.UpdateReceipt(ctx, tx, receipt); err != nil {
		return nil, err
	}

	// Retrieve the updated receipt
	return s.receiptRepo.GetReceiptByID(tx, receipt.ReceiptID)
}

// DeleteReceipt deletes a receipt
func (s *ReceiptService) DeleteReceipt(ctx context.Context, tx *gorm.DB, receiptID int) error {
	return s.receiptRepo.DeleteReceipt(ctx, tx, receiptID)
}
