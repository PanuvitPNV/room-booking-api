package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// ReceiptHandler handles HTTP requests related to receipts
type ReceiptHandler struct {
	receiptService *services.ReceiptService
}

// NewReceiptHandler creates a new receipt handler
func NewReceiptHandler(receiptService *services.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// CreateReceiptRequest represents a request to create a receipt
type CreateReceiptRequest struct {
	BookingID     int    `json:"booking_id" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required,oneof=Credit Debit Bank Transfer"`
	Amount        int    `json:"amount" validate:"required,gt=0"`
}

// CreateReceipt handles the creation of a payment receipt
// @Summary Create a payment receipt
// @Description Process payment for a booking
// @Tags receipts
// @Accept json
// @Produce json
// @Param receipt body CreateReceiptRequest true "Receipt details"
// @Success 201 {object} models.Receipt
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /receipts [post]
func (h *ReceiptHandler) CreateReceipt(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateReceiptRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	receipt := models.Receipt{
		BookingID:     req.BookingID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		PaymentDate:   time.Now(),
	}

	if err := h.receiptService.CreateReceipt(ctx, &receipt); err != nil {
		log.Errorf("Failed to create receipt: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create receipt: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, receipt)
}

// GetReceipt retrieves a receipt by ID
// @Summary Get a receipt
// @Description Get a receipt by ID
// @Tags receipts
// @Produce json
// @Param id path int true "Receipt ID"
// @Success 200 {object} models.Receipt
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /receipts/{id} [get]
func (h *ReceiptHandler) GetReceipt(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid receipt ID",
		})
	}

	receipt, err := h.receiptService.GetReceiptByID(ctx, id)
	if err != nil {
		log.Errorf("Failed to get receipt: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Receipt not found",
		})
	}

	return c.JSON(http.StatusOK, receipt)
}

// GetReceiptByBooking retrieves a receipt by booking ID
// @Summary Get receipt by booking
// @Description Get a receipt associated with a booking
// @Tags receipts
// @Produce json
// @Param bookingId path int true "Booking ID"
// @Success 200 {object} models.Receipt
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /receipts/booking/{bookingId} [get]
func (h *ReceiptHandler) GetReceiptByBooking(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("bookingId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid booking ID",
		})
	}

	receipt, err := h.receiptService.GetReceiptByBookingID(ctx, id)
	if err != nil {
		log.Errorf("Failed to get receipt: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Receipt not found for booking",
		})
	}

	return c.JSON(http.StatusOK, receipt)
}

// ProcessRefundRequest represents a request to process a refund
type ProcessRefundRequest struct {
	BookingID int `json:"booking_id" validate:"required"`
}

// ProcessRefund processes a refund for a booking
// @Summary Process a refund
// @Description Process a refund for a booking
// @Tags receipts
// @Accept json
// @Produce json
// @Param refund body ProcessRefundRequest true "Refund details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /receipts/refund [post]
func (h *ReceiptHandler) ProcessRefund(c echo.Context) error {
	ctx := c.Request().Context()

	var req ProcessRefundRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	if err := h.receiptService.ProcessRefund(ctx, req.BookingID); err != nil {
		log.Errorf("Failed to process refund: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process refund: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Refund processed successfully",
	})
}

// GetAllReceipts retrieves all receipts with pagination
// @Summary Get all receipts
// @Description Get all receipts with pagination
// @Tags receipts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /receipts [get]
func (h *ReceiptHandler) GetAllReceipts(c echo.Context) error {
	ctx := c.Request().Context()

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.QueryParam("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	receipts, total, err := h.receiptService.GetAllReceipts(ctx, page, pageSize)
	if err != nil {
		log.Errorf("Failed to get receipts: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get receipts: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"receipts":   receipts,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetReceiptsByDateRangeRequest represents a request to find receipts in a date range
type GetReceiptsByDateRangeRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// GetReceiptsByDateRange retrieves receipts for a date range
// @Summary Get receipts by date range
// @Description Get all receipts within a date range
// @Tags receipts
// @Accept json
// @Produce json
// @Param dates body GetReceiptsByDateRangeRequest true "Date range"
// @Success 200 {array} models.Receipt
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /receipts/by-date [post]
func (h *ReceiptHandler) GetReceiptsByDateRange(c echo.Context) error {
	ctx := c.Request().Context()

	var req GetReceiptsByDateRangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data: " + err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	receipts, err := h.receiptService.GetReceiptsByDateRange(ctx, req.StartDate, req.EndDate)
	if err != nil {
		log.Errorf("Failed to get receipts: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get receipts: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, receipts)
}
