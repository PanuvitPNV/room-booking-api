package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/databases"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"github.com/panuvitpnv/room-booking-api/internal/services"
	"github.com/panuvitpnv/room-booking-api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Improved deadlock test that creates a definite deadlock situation
func main() {
	// Enable deadlock testing mode
	os.Setenv("DEADLOCK_TEST_MODE", "true")
	os.Setenv("ENABLE_DEADLOCK_MODE", "true")

	log.Println("Starting improved deadlock test with DEADLOCK_TEST_MODE=true")

	// Set up services
	db := setupDatabase()
	utils.SetDB(db)

	// Set up repositories and services
	logger := log.New(os.Stdout, "[DEADLOCK_TEST] ", log.LstdFlags)
	lockManager := utils.NewLockManager(500 * time.Millisecond)

	bookingRepo := repositories.NewBookingRepository(db, logger)
	roomRepo := repositories.NewRoomRepository(db)
	bookingService := services.NewBookingService(bookingRepo, roomRepo, lockManager, logger)

	// Run tests
	log.Println("Running guaranteed deadlock test scenario...")
	runGuaranteedDeadlockTest(db)

	log.Println("Running cross update deadlock test...")
	runCrossUpdateDeadlockTest(bookingService)

	log.Println("Running aggressive concurrent booking test...")
	runAggressiveConcurrentBookingTest(bookingService)

	log.Println("Deadlock tests completed")
}

func setupDatabase() *gorm.DB {
	// Load configuration
	cfg := config.ConfigGetting()

	// Connect to database
	db := databases.NewPostgresDatabase(cfg.Database).Connect()
	return db
}

// runGuaranteedDeadlockTest forces a deadlock at the database level
func runGuaranteedDeadlockTest(db *gorm.DB) {
	logger := log.New(os.Stdout, "[GUARANTEED_DEADLOCK] ", log.LstdFlags)
	logger.Println("Setting up test data for guaranteed deadlock")

	// Create two room status entries to lock
	roomStatus1 := models.RoomStatus{
		RoomNum:  101,
		Calendar: time.Now().AddDate(0, 0, 30), // 30 days in future
		Status:   "Available",
	}

	roomStatus2 := models.RoomStatus{
		RoomNum:  102,
		Calendar: time.Now().AddDate(0, 0, 30), // 30 days in future
		Status:   "Available",
	}

	// Save the test data
	db.Save(&roomStatus1)
	db.Save(&roomStatus2)

	logger.Printf("Created room status records for rooms 101 and 102")

	// Channel to signal when deadlock is detected
	deadlockChan := make(chan bool, 2)

	// Create two goroutines that will acquire locks in opposite orders
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("TX1 recovered from panic: %v", r)
				deadlockChan <- true
			}
		}()

		logger.Println("TX1: Starting transaction")
		tx := db.Begin()
		defer tx.Rollback()

		// First, lock room 101
		logger.Println("TX1: Locking room 101")
		var status1 models.RoomStatus
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ?", 101).
			First(&status1).Error; err != nil {
			logger.Printf("TX1: Error locking room 101: %v", err)
			deadlockChan <- true
			return
		}

		logger.Println("TX1: Acquired lock on room 101, sleeping before locking room 102")
		time.Sleep(3 * time.Second) // Sleep to ensure TX2 has locked room 102

		// Now try to lock room 102 (potential deadlock)
		logger.Println("TX1: Now trying to lock room 102...")
		var status2 models.RoomStatus
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ?", 102).
			First(&status2).Error; err != nil {
			logger.Printf("TX1: Error or deadlock detected when locking room 102: %v", err)
			deadlockChan <- true
			return
		}

		// If we get here, we've acquired both locks
		logger.Println("TX1: Successfully acquired both locks (unexpected in deadlock)")
		tx.Commit()
		deadlockChan <- false
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("TX2 recovered from panic: %v", r)
				deadlockChan <- true
			}
		}()

		logger.Println("TX2: Starting transaction")
		tx := db.Begin()
		defer tx.Rollback()

		// Short delay to ensure TX1 starts first
		time.Sleep(500 * time.Millisecond)

		// First, lock room 102
		logger.Println("TX2: Locking room 102")
		var status2 models.RoomStatus
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ?", 102).
			First(&status2).Error; err != nil {
			logger.Printf("TX2: Error locking room 102: %v", err)
			deadlockChan <- true
			return
		}

		logger.Println("TX2: Acquired lock on room 102, sleeping before locking room 101")
		time.Sleep(1 * time.Second) // Sleep to ensure deadlock condition

		// Now try to lock room 101 (this will cause deadlock)
		logger.Println("TX2: Now trying to lock room 101...")
		var status1 models.RoomStatus
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_num = ?", 101).
			First(&status1).Error; err != nil {
			logger.Printf("TX2: Error or deadlock detected when locking room 101: %v", err)
			deadlockChan <- true
			return
		}

		// If we get here, we've acquired both locks
		logger.Println("TX2: Successfully acquired both locks (unexpected in deadlock)")
		tx.Commit()
		deadlockChan <- false
	}()

	// Wait for result with timeout
	timeout := time.After(20 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case deadlockDetected := <-deadlockChan:
			if deadlockDetected {
				logger.Println("Deadlock was detected and handled!")
			} else {
				logger.Println("No deadlock occurred (unexpected)")
			}
		case <-timeout:
			logger.Println("Test timed out - deadlock may have occurred but wasn't resolved")
			return
		}
	}
}

// runCrossUpdateDeadlockTest tests a classic deadlock scenario
func runCrossUpdateDeadlockTest(bookingService *services.BookingService) {
	logger := log.New(os.Stdout, "[CROSS_UPDATE_TEST] ", log.LstdFlags)
	ctx := context.Background()

	// Create two initial bookings
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

	// Create first booking
	err := bookingService.CreateBooking(ctx, booking1)
	if err != nil {
		logger.Printf("Failed to create booking 1: %v", err)
		return
	}
	logger.Printf("Created booking %d for room %d", booking1.BookingID, booking1.RoomNum)

	// Create second booking
	err = bookingService.CreateBooking(ctx, booking2)
	if err != nil {
		logger.Printf("Failed to create booking 2: %v", err)
		return
	}
	logger.Printf("Created booking %d for room %d", booking2.BookingID, booking2.RoomNum)

	// Channel to detect deadlocks
	deadlockChan := make(chan bool, 2)

	// Set up two goroutines to update bookings in opposite order
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("TX1 recovered from panic: %v", r)
				deadlockChan <- true
				return
			}
		}()

		logger.Println("TX1: Updating booking 1, then booking 2")

		// Update booking 1
		err := bookingService.UpdateBooking(ctx, booking1.BookingID,
			booking1.CheckInDate.AddDate(0, 0, 1),
			booking1.CheckOutDate.AddDate(0, 0, 1))

		if err != nil {
			logger.Printf("TX1: Failed to update booking 1: %v", err)
			deadlockChan <- true
			return
		}

		logger.Println("TX1: Successfully updated booking 1, now updating booking 2")

		// Update booking 2
		err = bookingService.UpdateBooking(ctx, booking2.BookingID,
			booking2.CheckInDate.AddDate(0, 0, 1),
			booking2.CheckOutDate.AddDate(0, 0, 1))

		if err != nil {
			logger.Printf("TX1: Failed to update booking 2: %v", err)
			deadlockChan <- true
		} else {
			logger.Println("TX1: Successfully updated both bookings")
			deadlockChan <- false
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("TX2 recovered from panic: %v", r)
				deadlockChan <- true
				return
			}
		}()

		// Small delay to make sure TX1 starts first
		time.Sleep(100 * time.Millisecond)

		logger.Println("TX2: Updating booking 2, then booking 1")

		// Update booking 2 first (opposite order)
		err := bookingService.UpdateBooking(ctx, booking2.BookingID,
			booking2.CheckInDate.AddDate(0, 0, 2),
			booking2.CheckOutDate.AddDate(0, 0, 2))

		if err != nil {
			logger.Printf("TX2: Failed to update booking 2: %v", err)
			deadlockChan <- true
			return
		}

		logger.Println("TX2: Successfully updated booking 2, now updating booking 1")

		// Now try to update booking 1
		err = bookingService.UpdateBooking(ctx, booking1.BookingID,
			booking1.CheckInDate.AddDate(0, 0, 2),
			booking1.CheckOutDate.AddDate(0, 0, 2))

		if err != nil {
			logger.Printf("TX2: Failed to update booking 1: %v", err)
			deadlockChan <- true
		} else {
			logger.Println("TX2: Successfully updated both bookings")
			deadlockChan <- false
		}
	}()

	// Wait for result with timeout
	timeout := time.After(20 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case deadlockDetected := <-deadlockChan:
			if deadlockDetected {
				logger.Println("Deadlock was detected and handled!")
			} else {
				logger.Println("Transaction completed without deadlock")
			}
		case <-timeout:
			logger.Println("Test timed out - deadlock may have occurred but wasn't resolved")
			return
		}
	}
}

// runAggressiveConcurrentBookingTest tests multiple clients trying to book the same room
// with more aggressive locking to trigger deadlocks
func runAggressiveConcurrentBookingTest(bookingService *services.BookingService) {
	logger := log.New(os.Stdout, "[AGGRESSIVE_BOOKING_TEST] ", log.LstdFlags)
	ctx := context.Background()

	// Create several room status records for the same dates to increase contention
	specificDate := time.Now().AddDate(0, 0, 14)

	// Number of concurrent booking attempts
	concurrentBookings := 10
	var wg sync.WaitGroup
	wg.Add(concurrentBookings)

	// Run concurrent bookings with the same dates to maximize contention
	deadlockCount := 0
	var countMutex sync.Mutex

	for i := 0; i < concurrentBookings; i++ {
		go func(clientID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Printf("Client %d recovered from panic: %v", clientID, r)
					countMutex.Lock()
					deadlockCount++
					countMutex.Unlock()
				}
			}()

			// Use the same room and same dates for all bookings to maximize contention
			roomNum := 101

			// All clients try to book the same dates
			checkInDate := specificDate
			checkOutDate := checkInDate.AddDate(0, 0, 3)

			booking := &models.Booking{
				BookingName:  "Concurrent Test",
				RoomNum:      roomNum,
				CheckInDate:  checkInDate,
				CheckOutDate: checkOutDate,
			}

			logger.Printf("Client %d: Attempting to book room %d from %s to %s",
				clientID, roomNum, checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02"))

			err := bookingService.CreateBooking(ctx, booking)
			if err != nil {
				logger.Printf("Client %d: Booking failed: %v", clientID, err)

				// Check if it's a deadlock error
				if err != nil && strings.Contains(err.Error(), "deadlock") {
					countMutex.Lock()
					deadlockCount++
					countMutex.Unlock()
				}
			} else {
				logger.Printf("Client %d: Successfully booked room %d, booking ID: %d",
					clientID, roomNum, booking.BookingID)

				// Immediately try to update the booking to increase contention
				err = bookingService.UpdateBooking(ctx, booking.BookingID,
					booking.CheckInDate.AddDate(0, 0, 1),
					booking.CheckOutDate.AddDate(0, 0, 1))

				if err != nil {
					logger.Printf("Client %d: Update failed: %v", clientID, err)

					// Check if it's a deadlock error
					if err != nil && strings.Contains(err.Error(), "deadlock") {
						countMutex.Lock()
						deadlockCount++
						countMutex.Unlock()
					}
				} else {
					logger.Printf("Client %d: Successfully updated booking %d",
						clientID, booking.BookingID)
				}
			}
		}(i)
	}

	wg.Wait()
	logger.Printf("Concurrent booking test completed. Deadlocks detected: %d", deadlockCount)
}
