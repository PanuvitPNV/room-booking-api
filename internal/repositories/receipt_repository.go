package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/panuvitpnv/room-booking-api/internal/models"
)

// ReceiptRepository handles database operations for receipts
type ReceiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new receipt repository
func NewReceiptRepository(db *gorm.DB) *ReceiptRepository {
	return &ReceiptRepository{
		db: db,
	}
}

// CreateReceipt creates a payment receipt for a booking with transaction management
func (r *ReceiptRepository) CreateReceipt(tx *gorm.DB, receipt *models.Receipt) error {
	// First check if booking exists and lock the record
	var booking models.Booking
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&booking, receipt.BookingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("booking not found")
		}
		return err
	}

	// Check if a receipt already exists for this booking
	var existingReceipt models.Receipt
	err := tx.Where("booking_id = ?", receipt.BookingID).First(&existingReceipt).Error
	if err == nil {
		return errors.New("payment already processed for this booking")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Validate payment amount
	if receipt.Amount != booking.TotalPrice {
		return errors.New("payment amount does not match booking total price")
	}

	// Get the last used receipt ID number with locking
	var lastRunning models.LastRunning
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lastRunning).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Initialize if not exists
			lastRunning = models.LastRunning{LastRunning: 0, Year: time.Now().Year()}
			if err := tx.Create(&lastRunning).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Update last running with optimistic locking
	result := tx.Model(&models.LastRunning{}).
		Where("last_running = ?", lastRunning.LastRunning).
		Updates(map[string]interface{}{
			"last_running": lastRunning.LastRunning + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed to update last running number due to concurrent modification")
	}

	// Set receipt ID and timestamp
	receipt.ReceiptID = lastRunning.LastRunning + 1
	receipt.IssueDate = time.Now()

	// If payment date is not provided, use current time
	if receipt.PaymentDate.IsZero() {
		receipt.PaymentDate = time.Now()
	}

	// Create the receipt
	if err := tx.Create(receipt).Error; err != nil {
		return err
	}

	return nil
}

// GetReceiptByID retrieves a receipt by ID
func (r *ReceiptRepository) GetReceiptByID(tx *gorm.DB, receiptID int) (*models.Receipt, error) {
	var receipt models.Receipt
	err := tx.Preload("Booking").First(&receipt, receiptID).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetReceiptByBookingID retrieves a receipt by booking ID
func (r *ReceiptRepository) GetReceiptByBookingID(tx *gorm.DB, bookingID int) (*models.Receipt, error) {
	var receipt models.Receipt
	err := tx.Where("booking_id = ?", bookingID).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetAllReceipts retrieves all receipts with pagination
func (r *ReceiptRepository) GetAllReceipts(tx *gorm.DB, page, pageSize int) ([]models.Receipt, int64, error) {
	var receipts []models.Receipt
	var total int64

	// Count total records
	if err := tx.Model(&models.Receipt{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated receipts
	offset := (page - 1) * pageSize
	err := tx.Preload("Booking").
		Offset(offset).
		Limit(pageSize).
		Order("issue_date DESC").
		Find(&receipts).Error

	if err != nil {
		return nil, 0, err
	}

	return receipts, total, nil
}

// CancelReceipt cancels a receipt (for refunds)
func (r *ReceiptRepository) CancelReceipt(tx *gorm.DB, receiptID int) error {
	// Get the receipt with pessimistic locking
	var receipt models.Receipt
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&receipt, receiptID).Error; err != nil {
		return err
	}

	// Create refund receipt (negative amount)
	refundReceipt := models.Receipt{
		BookingID:     receipt.BookingID,
		PaymentDate:   time.Now(),
		PaymentMethod: receipt.PaymentMethod + " (Refund)",
		Amount:        -receipt.Amount,
		IssueDate:     time.Now(),
	}

	// Get the last used receipt ID number with locking
	var lastRunning models.LastRunning
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lastRunning).Error; err != nil {
		return err
	}

	// Update last running with optimistic locking
	result := tx.Model(&models.LastRunning{}).
		Where("last_running = ?", lastRunning.LastRunning).
		Updates(map[string]interface{}{
			"last_running": lastRunning.LastRunning + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed to update last running number due to concurrent modification")
	}

	// Set receipt ID
	refundReceipt.ReceiptID = lastRunning.LastRunning + 1

	// Create the refund receipt
	if err := tx.Create(&refundReceipt).Error; err != nil {
		return err
	}

	return nil
}
