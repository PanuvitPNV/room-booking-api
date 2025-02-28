package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReceiptRepository handles database operations for receipts
type ReceiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new ReceiptRepository
func NewReceiptRepository(db *gorm.DB) *ReceiptRepository {
	return &ReceiptRepository{
		db: db,
	}
}

// GetReceiptByID retrieves a receipt by ID
func (r *ReceiptRepository) GetReceiptByID(tx *gorm.DB, receiptID int) (*models.Receipt, error) {
	var receipt models.Receipt
	err := tx.Preload("Booking").
		First(&receipt, "receipt_id = ?", receiptID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("receipt not found")
		}
		return nil, err
	}
	return &receipt, nil
}

// GetReceiptByBookingID retrieves a receipt by booking ID
func (r *ReceiptRepository) GetReceiptByBookingID(tx *gorm.DB, bookingID int) (*models.Receipt, error) {
	var receipt models.Receipt
	err := tx.Preload("Booking").
		First(&receipt, "booking_id = ?", bookingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("receipt not found")
		}
		return nil, err
	}
	return &receipt, nil
}

// CreateReceipt creates a new receipt with concurrency control
func (r *ReceiptRepository) CreateReceipt(ctx context.Context, tx *gorm.DB, receipt *models.Receipt) error {
	// Check if booking exists
	var booking models.Booking
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&booking, "booking_id = ?", receipt.BookingID).Error; err != nil {
		return err
	}

	// Check if receipt already exists for this booking
	var count int64
	if err := tx.Model(&models.Receipt{}).
		Where("booking_id = ?", receipt.BookingID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("receipt already exists for this booking")
	}

	// Set issue date if not provided
	if receipt.IssueDate.IsZero() {
		receipt.IssueDate = time.Now()
	}

	// Create receipt
	return tx.Create(receipt).Error
}

// UpdateReceipt updates a receipt with concurrency control
func (r *ReceiptRepository) UpdateReceipt(ctx context.Context, tx *gorm.DB, receipt *models.Receipt) error {
	// Check if receipt exists and lock it
	var existingReceipt models.Receipt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&existingReceipt, "receipt_id = ?", receipt.ReceiptID).Error; err != nil {
		return err
	}

	// Update receipt
	return tx.Save(receipt).Error
}

// DeleteReceipt deletes a receipt
func (r *ReceiptRepository) DeleteReceipt(ctx context.Context, tx *gorm.DB, receiptID int) error {
	// Check if receipt exists and lock it
	var receipt models.Receipt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&receipt, "receipt_id = ?", receiptID).Error; err != nil {
		return err
	}

	// Delete receipt
	return tx.Delete(&receipt).Error
}
