package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/services"
	"github.com/panuvitpnv/room-booking-api/internal/utils"
)

// TestHandler handles test-related HTTP requests
type TestHandler struct {
	bookingService *services.BookingService
	logger         *log.Logger
}

// NewTestHandler creates a new test handler
func NewTestHandler(bookingService *services.BookingService) *TestHandler {
	logger := log.New(os.Stdout, "[TEST_HANDLER] ", log.LstdFlags)

	return &TestHandler{
		bookingService: bookingService,
		logger:         logger,
	}
}

// RegisterRoutes registers test-related routes
func (h *TestHandler) RegisterRoutes(e *echo.Echo) {
	testGroup := e.Group("/test")

	// Endpoint to trigger a concurrent booking scenario
	testGroup.GET("/concurrent-bookings/:roomNum/:count", h.ConcurrentBookings)

	// Endpoint to trigger a deadlock scenario
	testGroup.GET("/deadlock", h.Deadlock)

	// Status endpoint to verify the test controller is working
	testGroup.GET("/status", h.Status)

	h.logger.Println("Test routes registered")
}

// Status handles GET /test/status
func (h *TestHandler) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "Test controller is operational",
		"mode":   os.Getenv("DEADLOCK_TEST_MODE"),
	})
}

// ConcurrentBookings handles GET /test/concurrent-bookings/:roomNum/:count
func (h *TestHandler) ConcurrentBookings(c echo.Context) error {
	roomNumStr := c.Param("roomNum")
	countStr := c.Param("count")

	roomNum, err := strconv.Atoi(roomNumStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid room number",
		})
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		count = 5 // Default to 5 concurrent bookings
	}

	// Run the concurrent booking test in the background
	go h.runConcurrentBookingTest(roomNum, count)

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "Test started",
		"message": fmt.Sprintf("Running %d concurrent bookings for room %d", count, roomNum),
	})
}

// Deadlock handles GET /test/deadlock
func (h *TestHandler) Deadlock(c echo.Context) error {
	// Run the deadlock test in the background
	go h.runDeadlockTest()

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "Deadlock test started",
		"message": "Running deadlock test in the background",
	})
}

// runConcurrentBookingTest creates multiple concurrent bookings for the same room
func (h *TestHandler) runConcurrentBookingTest(roomNum, count int) {
	h.logger.Printf("Starting concurrent booking test for room %d with %d clients", roomNum, count)

	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(id int) {
			defer wg.Done()

			// Create slightly offset booking dates to allow some bookings to succeed
			checkInDate := time.Now().AddDate(0, 0, id)
			checkOutDate := checkInDate.AddDate(0, 0, 3)

			booking := &models.Booking{
				BookingName:  fmt.Sprintf("Test Customer %d", id),
				RoomNum:      roomNum,
				CheckInDate:  checkInDate,
				CheckOutDate: checkOutDate,
			}

			h.logger.Printf("Client %d: Attempting to book room %d", id, roomNum)

			err := h.bookingService.CreateBooking(context.Background(), booking)
			if err != nil {
				h.logger.Printf("Client %d: Booking failed: %v", id, err)
			} else {
				h.logger.Printf("Client %d: Successfully booked room %d with ID %d",
					id, roomNum, booking.BookingID)
			}
		}(i)
	}

	wg.Wait()
	h.logger.Println("Concurrent bookings test completed")
}

// runDeadlockTest executes a test designed to trigger a true deadlock
func (h *TestHandler) runDeadlockTest() {
	h.logger.Println("Starting true deadlock test with cross-booking updates")

	ctx := context.Background()

	// Create two bookings for two different rooms
	booking1 := &models.Booking{
		BookingName:  "Deadlock Test A",
		RoomNum:      101,
		CheckInDate:  time.Now().AddDate(0, 0, 1),
		CheckOutDate: time.Now().AddDate(0, 0, 3),
	}

	booking2 := &models.Booking{
		BookingName:  "Deadlock Test B",
		RoomNum:      102,
		CheckInDate:  time.Now().AddDate(0, 0, 1),
		CheckOutDate: time.Now().AddDate(0, 0, 3),
	}

	err := h.bookingService.CreateBooking(ctx, booking1)
	if err != nil {
		h.logger.Printf("Failed to create booking 1: %v", err)
		return
	}

	err = h.bookingService.CreateBooking(ctx, booking2)
	if err != nil {
		h.logger.Printf("Failed to create booking 2: %v", err)
		return
	}

	h.logger.Printf("Created bookings: %d and %d", booking1.BookingID, booking2.BookingID)

	// Create room status entries for both bookings to ensure they exist
	err = utils.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Create status entries for booking 1
		status1 := models.RoomStatus{
			RoomNum:   booking1.RoomNum,
			Calendar:  booking1.CheckInDate,
			Status:    "Occupied",
			BookingID: &booking1.BookingID,
		}
		if err := tx.Save(&status1).Error; err != nil {
			return err
		}

		// Create status entries for booking 2
		status2 := models.RoomStatus{
			RoomNum:   booking2.RoomNum,
			Calendar:  booking2.CheckInDate,
			Status:    "Occupied",
			BookingID: &booking2.BookingID,
		}
		if err := tx.Save(&status2).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.Printf("Failed to create room statuses: %v", err)
		return
	}

	h.logger.Println("Created room status entries for both bookings")

	// Wait for 1 second to ensure all previous transactions are completed
	time.Sleep(1 * time.Second)

	// Now try to update both bookings in opposite order from two goroutines
	// Transaction 1: Will lock booking1->booking2
	go func() {
		utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			h.logger.Println("TX1: Locking booking 1, then will try booking 2")

			// First, lock booking 1's room status
			var statuses1 []models.RoomStatus
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("room_num = ? AND booking_id = ?", booking1.RoomNum, booking1.BookingID).
				Find(&statuses1).Error; err != nil {
				h.logger.Printf("TX1: Failed to lock booking 1 statuses: %v", err)
				return err
			}

			h.logger.Println("TX1: Successfully locked booking 1 statuses")

			// Sleep to ensure TX2 has time to lock booking 2
			time.Sleep(500 * time.Millisecond)

			// Now try to lock booking 2's room status (this should deadlock)
			h.logger.Println("TX1: Now trying to lock booking 2 statuses...")
			var statuses2 []models.RoomStatus
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("room_num = ? AND booking_id = ?", booking2.RoomNum, booking2.BookingID).
				Find(&statuses2).Error; err != nil {
				h.logger.Printf("TX1: Failed to lock booking 2 statuses: %v", err)
				return err
			}

			h.logger.Println("TX1: Successfully locked both bookings' statuses (should not happen in deadlock)")
			return nil
		})
	}()

	// Transaction 2: Will lock booking2->booking1 (opposite order)
	go func() {
		utils.WithTransaction(ctx, func(tx *gorm.DB) error {
			h.logger.Println("TX2: Locking booking 2, then will try booking 1")

			// First, lock booking 2's room status
			var statuses2 []models.RoomStatus
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("room_num = ? AND booking_id = ?", booking2.RoomNum, booking2.BookingID).
				Find(&statuses2).Error; err != nil {
				h.logger.Printf("TX2: Failed to lock booking 2 statuses: %v", err)
				return err
			}

			h.logger.Println("TX2: Successfully locked booking 2 statuses")

			// Sleep to ensure TX1 has time to lock booking 1
			time.Sleep(500 * time.Millisecond)

			// Now try to lock booking 1's room status (this should deadlock)
			h.logger.Println("TX2: Now trying to lock booking 1 statuses...")
			var statuses1 []models.RoomStatus
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("room_num = ? AND booking_id = ?", booking1.RoomNum, booking1.BookingID).
				Find(&statuses1).Error; err != nil {
				h.logger.Printf("TX2: Failed to lock booking 1 statuses: %v", err)
				return err
			}

			h.logger.Println("TX2: Successfully locked both bookings' statuses (should not happen in deadlock)")
			return nil
		})
	}()

	// Wait for 15 seconds to see the deadlock unfold (or until PostgreSQL timeout)
	h.logger.Println("Waiting for deadlock detection... (should see deadlock error in logs)")
	time.Sleep(15 * time.Second)
	h.logger.Println("Deadlock test completed")
}
