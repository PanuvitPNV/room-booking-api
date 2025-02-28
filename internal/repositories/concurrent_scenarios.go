package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ConcurrentScenarios provides examples of transaction and concurrency control scenarios
type ConcurrentScenarios struct {
	db *gorm.DB
}

// NewConcurrentScenarios creates a new ConcurrentScenarios instance
func NewConcurrentScenarios(db *gorm.DB) *ConcurrentScenarios {
	return &ConcurrentScenarios{
		db: db,
	}
}

// LostUpdate demonstrates the lost update problem and its solution
func (cs *ConcurrentScenarios) LostUpdate(ctx context.Context, bookingID int) (string, error) {
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Retrieve original booking
	var originalBooking models.Booking
	if err := cs.db.First(&originalBooking, "booking_id = ?", bookingID).Error; err != nil {
		return "", err
	}

	// Record original values
	originalName := originalBooking.BookingName

	// Start two concurrent transactions
	wg.Add(2)

	// First transaction - updates the booking name to "Alice"
	go func() {
		defer wg.Done()

		// Using sleep to simulate transaction timing
		time.Sleep(100 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Read the booking
			var booking models.Booking
			if err := tx.First(&booking, "booking_id = ?", bookingID).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: Read booking name: %s", booking.BookingName))
			mu.Unlock()

			// Modify the booking name
			booking.BookingName = "Alice"

			// Wait to simulate processing time
			time.Sleep(500 * time.Millisecond)

			// Save the changes
			if err := tx.Save(&booking).Error; err != nil {
				mu.Lock()
				results = append(results, fmt.Sprintf("Transaction 1: Failed to update: %s", err.Error()))
				mu.Unlock()
				return err
			}

			mu.Lock()
			results = append(results, "Transaction 1: Updated booking name to 'Alice'")
			mu.Unlock()
			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Second transaction - updates the booking name to "Bob"
	go func() {
		defer wg.Done()

		// Using sleep to simulate transaction timing
		time.Sleep(200 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Read the booking
			var booking models.Booking
			if err := tx.First(&booking, "booking_id = ?", bookingID).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2: Read booking name: %s", booking.BookingName))
			mu.Unlock()

			// Modify the booking name
			booking.BookingName = "Bob"

			// Save the changes
			if err := tx.Save(&booking).Error; err != nil {
				mu.Lock()
				results = append(results, fmt.Sprintf("Transaction 2: Failed to update: %s", err.Error()))
				mu.Unlock()
				return err
			}

			mu.Lock()
			results = append(results, "Transaction 2: Updated booking name to 'Bob'")
			mu.Unlock()
			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Wait for both transactions to complete
	wg.Wait()

	// Check the final state
	var finalBooking models.Booking
	if err := cs.db.First(&finalBooking, "booking_id = ?", bookingID).Error; err != nil {
		return "", err
	}

	mu.Lock()
	results = append(results, fmt.Sprintf("Final booking name: %s", finalBooking.BookingName))
	results = append(results, fmt.Sprintf("Original name: %s", originalName))
	results = append(results, "In this scenario, a lost update occurred because Transaction 2 didn't know about Transaction 1's changes.")
	resultsStr := ""
	for _, res := range results {
		resultsStr += res + "\n"
	}
	mu.Unlock()

	return resultsStr, nil
}

// LostUpdateWithPessimisticLocking demonstrates preventing lost updates using pessimistic locking
func (cs *ConcurrentScenarios) LostUpdateWithPessimisticLocking(ctx context.Context, bookingID int) (string, error) {
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Retrieve original booking
	var originalBooking models.Booking
	if err := cs.db.First(&originalBooking, "booking_id = ?", bookingID).Error; err != nil {
		return "", err
	}

	// Record original values
	originalName := originalBooking.BookingName

	// Start two concurrent transactions
	wg.Add(2)

	// First transaction - updates the booking name to "Alice" with locking
	go func() {
		defer wg.Done()

		// Using sleep to simulate transaction timing
		time.Sleep(100 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Read the booking with lock
			var booking models.Booking
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&booking, "booking_id = ?", bookingID).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: Read booking name with lock: %s", booking.BookingName))
			mu.Unlock()

			// Modify the booking name
			booking.BookingName = "Alice"

			// Wait to simulate processing time
			time.Sleep(500 * time.Millisecond)

			// Save the changes
			if err := tx.Save(&booking).Error; err != nil {
				mu.Lock()
				results = append(results, fmt.Sprintf("Transaction 1: Failed to update: %s", err.Error()))
				mu.Unlock()
				return err
			}

			mu.Lock()
			results = append(results, "Transaction 1: Updated booking name to 'Alice'")
			mu.Unlock()
			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Second transaction - updates the booking name to "Bob" with locking
	go func() {
		defer wg.Done()

		// Using sleep to simulate transaction timing
		time.Sleep(200 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Read the booking with lock - this will block until Transaction 1 commits
			var booking models.Booking
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&booking, "booking_id = ?", bookingID).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2: Read booking name with lock: %s", booking.BookingName))
			mu.Unlock()

			// Modify the booking name
			booking.BookingName = "Bob"

			// Save the changes
			if err := tx.Save(&booking).Error; err != nil {
				mu.Lock()
				results = append(results, fmt.Sprintf("Transaction 2: Failed to update: %s", err.Error()))
				mu.Unlock()
				return err
			}

			mu.Lock()
			results = append(results, "Transaction 2: Updated booking name to 'Bob'")
			mu.Unlock()
			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Wait for both transactions to complete
	wg.Wait()

	// Check the final state
	var finalBooking models.Booking
	if err := cs.db.First(&finalBooking, "booking_id = ?", bookingID).Error; err != nil {
		return "", err
	}

	mu.Lock()
	results = append(results, fmt.Sprintf("Final booking name: %s", finalBooking.BookingName))
	results = append(results, fmt.Sprintf("Original name: %s", originalName))
	results = append(results, "With pessimistic locking, Transaction 2 had to wait for Transaction 1 to complete.\nTransaction 2 then worked with the updated data, preventing the lost update problem.")
	resultsStr := ""
	for _, res := range results {
		resultsStr += res + "\n"
	}
	mu.Unlock()

	return resultsStr, nil
}

// DirtyRead demonstrates a dirty read problem and its prevention
func (cs *ConcurrentScenarios) DirtyRead(ctx context.Context, bookingID int) (string, error) {
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Retrieve original booking
	var originalBooking models.Booking
	if err := cs.db.First(&originalBooking, "booking_id = ?", bookingID).Error; err != nil {
		return "", err
	}

	// Record original values
	originalPrice := originalBooking.TotalPrice

	// Start two concurrent transactions
	wg.Add(2)

	// First transaction - updates the price but then rolls back
	go func() {
		defer wg.Done()

		tx := cs.db.Begin()

		// Read the booking
		var booking models.Booking
		if err := tx.First(&booking, "booking_id = ?", bookingID).Error; err != nil {
			tx.Rollback()
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1 failed to read: %s", err.Error()))
			mu.Unlock()
			return
		}

		mu.Lock()
		results = append(results, fmt.Sprintf("Transaction 1: Read booking price: %d", booking.TotalPrice))
		mu.Unlock()

		// Modify the price
		newPrice := booking.TotalPrice * 2
		if err := tx.Model(&models.Booking{}).
			Where("booking_id = ?", bookingID).
			Update("total_price", newPrice).Error; err != nil {
			tx.Rollback()
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: Failed to update price: %s", err.Error()))
			mu.Unlock()
			return
		}

		mu.Lock()
		results = append(results, fmt.Sprintf("Transaction 1: Updated price to %d (uncommitted)", newPrice))
		mu.Unlock()

		// Wait to simulate processing time before rolling back
		time.Sleep(500 * time.Millisecond)

		// Rollback the transaction
		tx.Rollback()
		mu.Lock()
		results = append(results, "Transaction 1: Rolled back the price change")
		mu.Unlock()
	}()

	// Second transaction - reads the price during the first transaction's execution
	go func() {
		defer wg.Done()

		// Wait to ensure the first transaction has updated the price but not yet rolled back
		time.Sleep(300 * time.Millisecond)

		// Read the booking with different isolation levels
		// 1. Default isolation level (might read uncommitted data)
		var booking1 models.Booking
		if err := cs.db.First(&booking1, "booking_id = ?", bookingID).Error; err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2 failed to read: %s", err.Error()))
			mu.Unlock()
			return
		}

		mu.Lock()
		results = append(results, fmt.Sprintf("Transaction 2: Read booking price (default isolation): %d", booking1.TotalPrice))
		mu.Unlock()

		// 2. Read committed isolation level
		var booking2 models.Booking
		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Execute with read committed isolation
			if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL READ COMMITTED").Error; err != nil {
				return err
			}

			return tx.First(&booking2, "booking_id = ?", bookingID).Error
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2 failed to read with READ COMMITTED: %s", err.Error()))
			mu.Unlock()
			return
		}

		mu.Lock()
		results = append(results, fmt.Sprintf("Transaction 2: Read booking price (READ COMMITTED): %d", booking2.TotalPrice))
		mu.Unlock()
	}()

	// Wait for both transactions to complete
	wg.Wait()

	// Check the final state
	var finalBooking models.Booking
	if err := cs.db.First(&finalBooking, "booking_id = ?", bookingID).Error; err != nil {
		return "", err
	}

	mu.Lock()
	results = append(results, fmt.Sprintf("Final booking price: %d", finalBooking.TotalPrice))
	results = append(results, fmt.Sprintf("Original price: %d", originalPrice))
	results = append(results, "This demonstrates how the isolation level affects whether a transaction can see uncommitted changes from another transaction.")
	resultsStr := ""
	for _, res := range results {
		resultsStr += res + "\n"
	}
	mu.Unlock()

	return resultsStr, nil
}

// PhantomRead demonstrates a phantom read problem and its prevention
func (cs *ConcurrentScenarios) PhantomRead(ctx context.Context, checkInDate, checkOutDate time.Time) (string, error) {
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Start two concurrent transactions
	wg.Add(2)

	// First transaction - reads available rooms, then reads again
	go func() {
		defer wg.Done()

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// First read of available rooms
			var count1 int64
			err := tx.Model(&models.Room{}).
				Joins("LEFT JOIN room_status ON rooms.room_num = room_status.room_num AND room_status.calendar BETWEEN ? AND ?",
					checkInDate, checkOutDate).
				Where("room_status.status = 'Available' OR room_status.status IS NULL").
				Count(&count1).Error

			if err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: First read found %d available rooms", count1))
			mu.Unlock()

			// Wait to allow Transaction 2 to execute
			time.Sleep(500 * time.Millisecond)

			// Second read of available rooms (might see different results - phantom reads)
			var count2 int64
			err = tx.Model(&models.Room{}).
				Joins("LEFT JOIN room_status ON rooms.room_num = room_status.room_num AND room_status.calendar BETWEEN ? AND ?",
					checkInDate, checkOutDate).
				Where("room_status.status = 'Available' OR room_status.status IS NULL").
				Count(&count2).Error

			if err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: Second read found %d available rooms", count2))
			if count1 != count2 {
				results = append(results, "Transaction 1: Phantom read detected (counts are different)")
			} else {
				results = append(results, "Transaction 1: No phantom read detected (counts are the same)")
			}
			mu.Unlock()

			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Second transaction - creates a new booking
	go func() {
		defer wg.Done()

		// Wait to allow first transaction to do its first read
		time.Sleep(200 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Get an available room
			var room models.Room
			err := tx.Model(&models.Room{}).
				Joins("LEFT JOIN room_status ON rooms.room_num = room_status.room_num AND room_status.calendar BETWEEN ? AND ?",
					checkInDate, checkOutDate).
				Where("room_status.status = 'Available' OR room_status.status IS NULL").
				First(&room).Error

			if err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2: Found available room %d", room.RoomNum))
			mu.Unlock()

			// Create a new booking
			booking := models.Booking{
				BookingName:  "Phantom Test",
				RoomNum:      room.RoomNum,
				CheckInDate:  checkInDate,
				CheckOutDate: checkOutDate,
				BookingDate:  time.Now(),
				TotalPrice:   1000,
			}

			if err := tx.Create(&booking).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2: Created booking %d for room %d", booking.BookingID, room.RoomNum))
			mu.Unlock()

			// Update room status for each day of the booking
			current := checkInDate
			for current.Before(checkOutDate) || current.Equal(checkOutDate) {
				status := models.RoomStatus{
					RoomNum:   room.RoomNum,
					Calendar:  current,
					Status:    "Occupied",
					BookingID: &booking.BookingID,
				}

				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "room_num"}, {Name: "calendar"}},
					DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
				}).Create(&status).Error; err != nil {
					return err
				}

				current = current.AddDate(0, 0, 1)
			}

			mu.Lock()
			results = append(results, "Transaction 2: Updated room status to Occupied")
			mu.Unlock()

			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Wait for both transactions to complete
	wg.Wait()

	mu.Lock()
	results = append(results, "This demonstrates how phantom reads can occur when one transaction reads a set of rows twice, and another transaction adds or removes rows from that set between the reads.")
	resultsStr := ""
	for _, res := range results {
		resultsStr += res + "\n"
	}
	mu.Unlock()

	return resultsStr, nil
}

// SerializationAnomaly demonstrates a serialization anomaly problem and its prevention
func (cs *ConcurrentScenarios) SerializationAnomaly(ctx context.Context) (string, error) {
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Start two concurrent transactions
	wg.Add(2)

	// First transaction - counts total rooms and available rooms
	go func() {
		defer wg.Done()

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Read total rooms count
			var totalRooms int64
			if err := tx.Model(&models.Room{}).Count(&totalRooms).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: Total rooms count: %d", totalRooms))
			mu.Unlock()

			// Wait to allow Transaction 2 to execute
			time.Sleep(300 * time.Millisecond)

			// Calculate available rooms percentage
			var availableRooms int64
			if err := tx.Model(&models.RoomStatus{}).
				Where("status = 'Available'").
				Count(&availableRooms).Error; err != nil {
				return err
			}

			percentage := float64(availableRooms) / float64(totalRooms) * 100

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1: Available rooms: %d", availableRooms))
			results = append(results, fmt.Sprintf("Transaction 1: Available percentage: %.2f%%", percentage))
			mu.Unlock()

			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 1 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Second transaction - adds a new room
	go func() {
		defer wg.Done()

		// Wait to allow first transaction to read total count
		time.Sleep(100 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Find highest room number
			var maxRoom struct {
				MaxRoomNum int
			}
			if err := tx.Model(&models.Room{}).
				Select("COALESCE(MAX(room_num), 0) as max_room_num").
				Scan(&maxRoom).Error; err != nil {
				return err
			}

			// Create a new room
			newRoomNum := maxRoom.MaxRoomNum + 1
			room := models.Room{
				RoomNum: newRoomNum,
				TypeID:  1, // Assuming type ID 1 exists
			}

			if err := tx.Create(&room).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2: Created new room %d", newRoomNum))
			mu.Unlock()

			// Wait to simulate processing before committing
			time.Sleep(100 * time.Millisecond)

			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Transaction 2 failed: %s", err.Error()))
			mu.Unlock()
		}
	}()

	// Wait for both transactions to complete
	wg.Wait()

	mu.Lock()
	results = append(results, "This demonstrates a serialization anomaly where Transaction 1's calculations may be incorrect because Transaction 2 added a room between the two queries.")
	results = append(results, "To prevent this, use 'SERIALIZABLE' isolation level, which would either delay Transaction 2 or fail one of the transactions with a serialization failure.")
	resultsStr := ""
	for _, res := range results {
		resultsStr += res + "\n"
	}
	mu.Unlock()

	return resultsStr, nil
}

// ConcurrentBookings demonstrates how pessimistic locking prevents double booking
// ConcurrentBookings demonstrates how pessimistic locking prevents double booking
func (cs *ConcurrentScenarios) ConcurrentBookings(ctx context.Context, roomNum int, checkInDate, checkOutDate time.Time) (string, error) {
	var wg sync.WaitGroup
	var results []string
	var mu sync.Mutex

	// Start two concurrent bookings for the same room and dates
	wg.Add(2)

	// First booking attempt
	go func() {
		defer wg.Done()

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Lock the room for the date range
			err := tx.Exec(`
				SELECT rs.room_num 
				FROM room_status rs 
				WHERE rs.room_num = ? AND rs.calendar BETWEEN ? AND ?
				FOR UPDATE
			`, roomNum, checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02")).Error

			if err != nil {
				return err
			}

			mu.Lock()
			results = append(results, "Booking 1: Locked room status records")
			mu.Unlock()

			// Check if the room is available for the requested dates
			var conflictCount int64
			err = tx.Model(&models.RoomStatus{}).
				Where("room_num = ? AND calendar BETWEEN ? AND ? AND status = 'Occupied'",
					roomNum,
					checkInDate.Format("2006-01-02"),
					checkOutDate.Format("2006-01-02")).
				Count(&conflictCount).Error

			if err != nil {
				return err
			}

			if conflictCount > 0 {
				mu.Lock()
				results = append(results, "Booking 1: Room is not available for the selected dates")
				mu.Unlock()
				return fmt.Errorf("room is not available for the selected dates")
			}

			mu.Lock()
			results = append(results, "Booking 1: Room is available for booking")
			mu.Unlock()

			// Simulate a short delay in processing
			time.Sleep(500 * time.Millisecond)

			// Create booking
			booking := models.Booking{
				BookingName:  "Guest 1",
				RoomNum:      roomNum,
				CheckInDate:  checkInDate,
				CheckOutDate: checkOutDate,
				BookingDate:  time.Now(),
				TotalPrice:   1000,
			}

			if err := tx.Create(&booking).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Booking 1: Created booking %d", booking.BookingID))
			mu.Unlock()

			// Update room status for each day of the booking
			current := checkInDate
			for current.Before(checkOutDate) || current.Equal(checkOutDate) {
				status := models.RoomStatus{
					RoomNum:   roomNum,
					Calendar:  current,
					Status:    "Occupied",
					BookingID: &booking.BookingID,
				}

				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "room_num"}, {Name: "calendar"}},
					DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
				}).Create(&status).Error; err != nil {
					return err
				}

				current = current.AddDate(0, 0, 1)
			}

			mu.Lock()
			results = append(results, "Booking 1: Updated room status to Occupied")
			mu.Unlock()

			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Booking 1 failed: %s", err.Error()))
			mu.Unlock()
		} else {
			mu.Lock()
			results = append(results, "Booking 1: Successfully booked the room")
			mu.Unlock()
		}
	}()

	// Second booking attempt for the same room and dates
	go func() {
		defer wg.Done()

		// Small delay to ensure the first booking starts first
		time.Sleep(100 * time.Millisecond)

		err := cs.db.Transaction(func(tx *gorm.DB) error {
			// Try to lock the room for the date range
			// This will block until the first transaction completes due to FOR UPDATE
			err := tx.Exec(`
				SELECT rs.room_num 
				FROM room_status rs 
				WHERE rs.room_num = ? AND rs.calendar BETWEEN ? AND ?
				FOR UPDATE
			`, roomNum, checkInDate.Format("2006-01-02"), checkOutDate.Format("2006-01-02")).Error

			if err != nil {
				return err
			}

			mu.Lock()
			results = append(results, "Booking 2: Locked room status records (after Booking 1 released lock)")
			mu.Unlock()

			// Check if the room is available for the requested dates
			var conflictCount int64
			err = tx.Model(&models.RoomStatus{}).
				Where("room_num = ? AND calendar BETWEEN ? AND ? AND status = 'Occupied'",
					roomNum,
					checkInDate.Format("2006-01-02"),
					checkOutDate.Format("2006-01-02")).
				Count(&conflictCount).Error

			if err != nil {
				return err
			}

			if conflictCount > 0 {
				mu.Lock()
				results = append(results, "Booking 2: Room is not available for the selected dates (already booked by Booking 1)")
				mu.Unlock()
				return fmt.Errorf("room is not available for the selected dates")
			}

			mu.Lock()
			results = append(results, "Booking 2: Room is available for booking")
			mu.Unlock()

			// Create booking
			booking := models.Booking{
				BookingName:  "Guest 2",
				RoomNum:      roomNum,
				CheckInDate:  checkInDate,
				CheckOutDate: checkOutDate,
				BookingDate:  time.Now(),
				TotalPrice:   1000,
			}

			if err := tx.Create(&booking).Error; err != nil {
				return err
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("Booking 2: Created booking %d", booking.BookingID))
			mu.Unlock()

			// Update room status for each day of the booking
			current := checkInDate
			for current.Before(checkOutDate) || current.Equal(checkOutDate) {
				status := models.RoomStatus{
					RoomNum:   roomNum,
					Calendar:  current,
					Status:    "Occupied",
					BookingID: &booking.BookingID,
				}

				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "room_num"}, {Name: "calendar"}},
					DoUpdates: clause.AssignmentColumns([]string{"status", "booking_id"}),
				}).Create(&status).Error; err != nil {
					return err
				}

				current = current.AddDate(0, 0, 1)
			}

			mu.Lock()
			results = append(results, "Booking 2: Updated room status to Occupied")
			mu.Unlock()

			return nil
		})

		if err != nil {
			mu.Lock()
			results = append(results, fmt.Sprintf("Booking 2 failed: %s", err.Error()))
			mu.Unlock()
		} else {
			mu.Lock()
			results = append(results, "Booking 2: Successfully booked the room")
			mu.Unlock()
		}
	}()

	// Wait for both transactions to complete
	wg.Wait()

	mu.Lock()
	results = append(results, "This demonstrates how pessimistic locking with FOR UPDATE prevents double booking.")
	results = append(results, "The second booking had to wait for the first one to complete, and then it either succeeded (if the room was still available) or failed (if the first booking took the room).")
	resultsStr := ""
	for _, res := range results {
		resultsStr += res + "\n"
	}
	mu.Unlock()

	return resultsStr, nil
}
