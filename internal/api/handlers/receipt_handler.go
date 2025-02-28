package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/services"
)

// ReceiptHandler handles HTTP requests related to receipts
type ReceiptHandler struct {
	receiptService *services.ReceiptService
}

// NewReceiptHandler creates a new ReceiptHandler
func NewReceiptHandler(receiptService *services.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// RegisterRoutes registers all receipt routes
func (h *ReceiptHandler) RegisterRoutes(e *echo.Echo) {
	receipts := e.Group("/api/receipts")

	receipts.GET("/:id", h.GetReceiptByID)
	receipts.GET("/booking/:bookingId", h.GetReceiptByBookingID)
	receipts.POST("", h.CreateReceipt)
	receipts.PUT("/:id", h.UpdateReceipt)
	receipts.DELETE("/:id", h.DeleteReceipt)
}

// GetReceiptByID godoc
// @Summary      Get receipt by ID
// @Description  Retrieve a receipt by its ID
// @Tags         receipts
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Receipt ID"
// @Success      200  {object}  models.Receipt
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Receipt not found"
// @Router       /receipts/{id} [get]
func (h *ReceiptHandler) GetReceiptByID(c echo.Context) error {
	// Parse receipt ID from path
	receiptIDStr := c.Param("id")
	receiptID, err := strconv.Atoi(receiptIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receipt ID"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to get receipt
	receipt, err := h.receiptService.GetReceiptByID(tx, receiptID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Receipt not found: " + err.Error()})
	}

	return c.JSON(http.StatusOK, receipt)
}

// GetReceiptByBookingID godoc
// @Summary      Get receipt by booking ID
// @Description  Retrieve a receipt associated with a specific booking
// @Tags         receipts
// @Accept       json
// @Produce      json
// @Param        bookingId   path      integer  true  "Booking ID"
// @Success      200  {object}  models.Receipt
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Receipt not found"
// @Router       /receipts/booking/{bookingId} [get]
func (h *ReceiptHandler) GetReceiptByBookingID(c echo.Context) error {
	// Parse booking ID from path
	bookingIDStr := c.Param("bookingId")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to get receipt
	receipt, err := h.receiptService.GetReceiptByBookingID(tx, bookingID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Receipt not found: " + err.Error()})
	}

	return c.JSON(http.StatusOK, receipt)
}

// ReceiptRequest represents the request model for receipt operations
type ReceiptRequest struct {
	BookingID     int       `json:"booking_id" validate:"required"`
	PaymentDate   time.Time `json:"payment_date" validate:"required"`
	PaymentMethod string    `json:"payment_method" validate:"required,oneof=Credit Debit Bank Transfer"`
	Amount        int       `json:"amount" validate:"required,min=1"`
}

// CreateReceipt godoc
// @Summary      Create a new receipt
// @Description  Create a payment receipt for a booking with transaction control
// @Tags         receipts
// @Accept       json
// @Produce      json
// @Param        receipt  body      ReceiptRequest  true  "Receipt details"
// @Success      201      {object}  models.Receipt
// @Failure      400      {object}  map[string]string  "Bad request"
// @Router       /receipts [post]
func (h *ReceiptHandler) CreateReceipt(c echo.Context) error {
	// Parse request body
	var req ReceiptRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Basic validation
	if req.BookingID <= 0 || req.PaymentDate.IsZero() || req.PaymentMethod == "" || req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing or invalid required fields"})
	}

	// Create receipt model
	receipt := &models.Receipt{
		BookingID:     req.BookingID,
		PaymentDate:   req.PaymentDate,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		IssueDate:     time.Now(),
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to create receipt
	createdReceipt, err := h.receiptService.CreateReceipt(c.Request().Context(), tx, receipt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to create receipt: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, createdReceipt)
}

// UpdateReceipt godoc
// @Summary      Update a receipt
// @Description  Update an existing receipt with transaction control
// @Tags         receipts
// @Accept       json
// @Produce      json
// @Param        id       path      integer         true  "Receipt ID"
// @Param        receipt  body      ReceiptRequest  true  "Updated receipt details"
// @Success      200      {object}  models.Receipt
// @Failure      400      {object}  map[string]string  "Bad request"
// @Failure      404      {object}  map[string]string  "Receipt not found"
// @Router       /receipts/{id} [put]
func (h *ReceiptHandler) UpdateReceipt(c echo.Context) error {
	// Parse receipt ID from path
	receiptIDStr := c.Param("id")
	receiptID, err := strconv.Atoi(receiptIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receipt ID"})
	}

	// Parse request body
	var req ReceiptRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Get existing receipt
	existingReceipt, err := h.receiptService.GetReceiptByID(tx, receiptID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Receipt not found: " + err.Error()})
	}

	// Update fields
	existingReceipt.PaymentDate = req.PaymentDate
	existingReceipt.PaymentMethod = req.PaymentMethod
	existingReceipt.Amount = req.Amount

	// Call service to update receipt
	updatedReceipt, err := h.receiptService.UpdateReceipt(c.Request().Context(), tx, existingReceipt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to update receipt: " + err.Error()})
	}

	return c.JSON(http.StatusOK, updatedReceipt)
}

// DeleteReceipt godoc
// @Summary      Delete a receipt
// @Description  Delete an existing receipt with transaction control
// @Tags         receipts
// @Accept       json
// @Produce      json
// @Param        id   path      integer  true  "Receipt ID"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Router       /receipts/{id} [delete]
func (h *ReceiptHandler) DeleteReceipt(c echo.Context) error {
	// Parse receipt ID from path
	receiptIDStr := c.Param("id")
	receiptID, err := strconv.Atoi(receiptIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receipt ID"})
	}

	// Get database transaction from context
	tx := middleware.GetTransaction(c)

	// Call service to delete receipt
	err = h.receiptService.DeleteReceipt(c.Request().Context(), tx, receiptID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to delete receipt: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Receipt deleted successfully"})
}
