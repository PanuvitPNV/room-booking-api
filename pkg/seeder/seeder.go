package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
	"gorm.io/gorm"
)

// Global configuration for database access
var dbConfig *config.Config

func main() {
	// Initialize the random seed
	rand.Seed(time.Now().UnixNano())

	// Get configuration
	dbConfig = config.ConfigGetting()

	// Connect to database
	db := databases.NewPostgresDatabase(dbConfig.Database)
	gormDB := db.Connect()

	log.Println("Starting database seeding with transaction management and concurrency control...")

	// Check if we have existing data
	var existingRoomTypes int64
	gormDB.Model(&models.RoomType{}).Count(&existingRoomTypes)

	if existingRoomTypes > 0 {
		log.Println("Database already has data. Skipping basic seed operations.")
	} else {
		log.Println("Seeding basic data (room types, facilities, etc.)...")

		// Begin the main transaction for basic data
		tx := gormDB.Begin()
		if tx.Error != nil {
			log.Fatalf("Failed to begin transaction: %v", tx.Error)
		}

		// Defer a transaction rollback in case of errors
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Fatalf("Seeding failed with panic: %v", r)
			}
		}()

		// Seed room types
		roomTypes, err := seedRoomTypes(tx)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error seeding room types: %v", err)
		}

		// Seed facilities
		facilities, err := seedFacilities(tx)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error seeding facilities: %v", err)
		}

		// Seed room facilities
		if err := seedRoomFacilities(tx, roomTypes, facilities); err != nil {
			tx.Rollback()
			log.Fatalf("Error seeding room facilities: %v", err)
		}

		// Seed rooms
		rooms, err := seedRooms(tx, roomTypes)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Error seeding rooms: %v", err)
		}

		// Seed last running
		if err := seedLastRunning(tx); err != nil {
			tx.Rollback()
			log.Fatalf("Error seeding last running: %v", err)
		}

		// Commit transaction for basic data
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("Failed to commit basic data transaction: %v", err)
		}

		log.Println("Basic data seeding completed successfully!")

		// Seed bookings with sequential approach (no concurrency)
		if err := seedBookings(gormDB, rooms); err != nil {
			log.Fatalf("Error seeding bookings: %v", err)
		}
	}

	// Run conflict scenario test (if needed)
	if err := seedConflictScenarios(gormDB); err != nil {
		log.Printf("Error setting up conflict scenarios (non-fatal): %v", err)
	}

	log.Println("Database seeding completed successfully!")
	log.Println("Seeded:")
	log.Println("- Room types, facilities, and room-facility relationships")
	log.Println("- Rooms assigned to different room types")
	log.Println("- Realistic bookings with proper date management")
	log.Println("- Receipts for completed bookings")
	log.Println("- Room status entries with proper locking mechanisms")
	log.Println("- Test scenarios for concurrency handling")
}

func seedRoomTypes(tx *gorm.DB) ([]models.RoomType, error) {
	roomTypes := []models.RoomType{
		{
			TypeID:        1,
			Name:          "Standard",
			Description:   "A comfortable standard room with essential amenities",
			Area:          28,
			PricePerNight: 1200,
			NoOfGuest:     2,
		},
		{
			TypeID:        2,
			Name:          "Deluxe",
			Description:   "Spacious room with premium amenities and city view",
			Area:          35,
			PricePerNight: 1800,
			NoOfGuest:     2,
		},
		{
			TypeID:        3,
			Name:          "Suite",
			Description:   "Luxurious suite with separate living area and panoramic views",
			Area:          48,
			PricePerNight: 2500,
			NoOfGuest:     3,
		},
		{
			TypeID:        4,
			Name:          "Family Room",
			Description:   "Spacious room designed for family stays with additional beds",
			Area:          42,
			PricePerNight: 2200,
			NoOfGuest:     4,
		},
		{
			TypeID:        5,
			Name:          "Executive Suite",
			Description:   "Premium suite with executive benefits and luxury amenities",
			Area:          55,
			PricePerNight: 3500,
			NoOfGuest:     2,
		},
	}

	// Use batch insert with transaction for better performance
	if err := tx.CreateInBatches(&roomTypes, len(roomTypes)).Error; err != nil {
		return nil, fmt.Errorf("failed to create room types: %w", err)
	}

	log.Printf("Successfully seeded %d room types", len(roomTypes))
	return roomTypes, nil
}

func seedFacilities(tx *gorm.DB) ([]models.Facility, error) {
	facilities := []models.Facility{
		{FacilityID: 1, Name: "Wi-Fi"},
		{FacilityID: 2, Name: "Air Conditioning"},
		{FacilityID: 3, Name: "Flat-screen TV"},
		{FacilityID: 4, Name: "Mini Bar"},
		{FacilityID: 5, Name: "Coffee Maker"},
		{FacilityID: 6, Name: "Safe"},
		{FacilityID: 7, Name: "Bathtub"},
		{FacilityID: 8, Name: "Rainfall Shower"},
		{FacilityID: 9, Name: "Balcony"},
		{FacilityID: 10, Name: "City View"},
		{FacilityID: 11, Name: "King Size Bed"},
		{FacilityID: 12, Name: "Sofa"},
		{FacilityID: 13, Name: "Work Desk"},
		{FacilityID: 14, Name: "Breakfast Included"},
		{FacilityID: 15, Name: "Room Service"},
	}

	// Using batch insert for efficiency
	if err := tx.CreateInBatches(&facilities, len(facilities)).Error; err != nil {
		return nil, fmt.Errorf("failed to create facilities: %w", err)
	}

	log.Printf("Successfully seeded %d facilities", len(facilities))
	return facilities, nil
}

func seedRoomFacilities(tx *gorm.DB, roomTypes []models.RoomType, facilities []models.Facility) error {
	// Map each room type to a set of facilities
	roomTypeFacilities := map[int][]int{
		1: {1, 2, 3, 6, 13},                                    // Standard
		2: {1, 2, 3, 4, 5, 6, 8, 10, 13},                       // Deluxe
		3: {1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 15},         // Suite
		4: {1, 2, 3, 5, 6, 8, 13, 14},                          // Family Room
		5: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, // Executive Suite
	}

	var roomFacilities []models.RoomFacility

	for roomTypeID, facilityIDs := range roomTypeFacilities {
		for _, facilityID := range facilityIDs {
			roomFacilities = append(roomFacilities, models.RoomFacility{
				TypeID:     roomTypeID,
				FacilityID: facilityID,
			})
		}
	}

	// Using batch insert for efficiency
	if err := tx.CreateInBatches(&roomFacilities, len(roomFacilities)).Error; err != nil {
		return fmt.Errorf("failed to create room facilities: %w", err)
	}

	log.Printf("Successfully seeded %d room facility relationships", len(roomFacilities))
	return nil
}

func seedRooms(tx *gorm.DB, roomTypes []models.RoomType) ([]models.Room, error) {
	// Define how many rooms of each type we want
	roomCounts := map[int]int{
		1: 10, // 10 Standard rooms
		2: 8,  // 8 Deluxe rooms
		3: 5,  // 5 Suites
		4: 5,  // 5 Family rooms
		5: 2,  // 2 Executive suites
	}

	var rooms []models.Room
	roomNum := 101 // Starting room number

	for typeID, count := range roomCounts {
		for i := 0; i < count; i++ {
			rooms = append(rooms, models.Room{
				RoomNum: roomNum,
				TypeID:  typeID,
			})
			roomNum++
		}
	}

	// Using batch insert for efficiency
	if err := tx.CreateInBatches(&rooms, len(rooms)).Error; err != nil {
		return nil, fmt.Errorf("failed to create rooms: %w", err)
	}

	log.Printf("Successfully seeded %d rooms", len(rooms))
	return rooms, nil
}

// Fixed seedLastRunning function that ensures the LastRunning field is properly set
func seedLastRunning(tx *gorm.DB) error {
	// Check if a record already exists for the current year
	currentYear := time.Now().Year()
	var existingRecord models.LastRunning

	result := tx.Where("year = ?", currentYear).First(&existingRecord)
	if result.Error == nil {
		// Record already exists, no need to create a new one
		log.Printf("LastRunning record for year %d already exists with value %d",
			currentYear, existingRecord.LastRunning)
		return nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// If it's an error other than "record not found", return it
		return fmt.Errorf("error checking for existing LastRunning record: %w", result.Error)
	}

	// Create a new record with an explicit initial value
	lastRunning := models.LastRunning{
		LastRunning: 1, // Start with 1 instead of 0 to avoid potential NULL issues
		Year:        currentYear,
	}

	// Use SQL executor directly to ensure the value is set properly
	err := tx.Exec("INSERT INTO last_runnings (last_running, year) VALUES (?, ?)",
		lastRunning.LastRunning, lastRunning.Year).Error

	if err != nil {
		return fmt.Errorf("failed to create last running record: %w", err)
	}

	log.Printf("Successfully seeded last_runnings record with initial value %d for year %d",
		lastRunning.LastRunning, lastRunning.Year)
	return nil
}

// Helper function to generate reservation dates
func generateBookingDates(bookingType string) (time.Time, time.Time, time.Time) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := now.Month()
	bookingDate := now.AddDate(0, 0, -rand.Intn(30)) // Booking made 0-30 days ago

	var checkInDate, checkOutDate time.Time

	switch bookingType {
	case "past":
		// Past bookings (1-3 months ago)
		monthOffset := 1 + rand.Intn(3)
		dayOffset := 1 + rand.Intn(25)
		checkInDate = time.Date(currentYear, currentMonth-time.Month(monthOffset), dayOffset, 14, 0, 0, 0, time.Local)
		stayDuration := 1 + rand.Intn(5) // 1-5 nights
		checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
	case "current":
		// Current bookings (within last 5 days to next 5 days)
		daysOffset := rand.Intn(10) - 5
		checkInDate = now.AddDate(0, 0, daysOffset)
		stayDuration := 1 + rand.Intn(4) // 1-4 nights
		checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
	case "future":
		// Future bookings (next 1-6 months)
		monthOffset := 1 + rand.Intn(6)
		dayOffset := 1 + rand.Intn(25)
		checkInDate = time.Date(currentYear, currentMonth+time.Month(monthOffset), dayOffset, 14, 0, 0, 0, time.Local)
		stayDuration := 1 + rand.Intn(7) // 1-7 nights
		checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
	default:
		// Random bookings throughout the year if type is unknown
		daysOffset := rand.Intn(365) - 60 // -60 days to +305 days from now
		checkInDate = now.AddDate(0, 0, daysOffset)
		stayDuration := 1 + rand.Intn(10) // 1-10 nights
		checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
	}

	// Make sure booking date is before check-in date
	if bookingDate.After(checkInDate.AddDate(0, 0, -1)) {
		bookingDate = checkInDate.AddDate(0, 0, -1-rand.Intn(15))
	}

	return checkInDate, checkOutDate, bookingDate
}

// Non-concurrent booking seeding to avoid connection issues
func seedBookings(db *gorm.DB, rooms []models.Room) error {
	// Guest names for bookings
	guestNames := []string{
		"John Smith", "Jane Doe", "Michael Johnson", "Emily Davis",
		"David Wilson", "Sarah Brown", "Robert Taylor", "Jessica Miller",
		"Thomas Moore", "Jennifer Anderson", "William White", "Lisa Martinez",
		"Daniel Clark", "Mary Rodriguez", "James Lewis", "Patricia Allen",
		"Christopher Young", "Barbara Hall", "Matthew King", "Elizabeth Scott",
		"Charles Harris", "Nancy Jackson", "Brian Thompson", "Susan Wright",
		"Kevin Green", "Karen Walker", "Edward Baker", "Margaret Phillips",
		"George Turner", "Sandra Lee", "Mark Wright", "Michelle Hall",
		"Richard Adams", "Donna Nelson", "Joseph Carter", "Ruth Thomas",
		"Kenneth Lewis", "Carol Young", "Paul Scott", "Sharon Harris",
	}

	paymentMethods := []string{"Credit", "Debit", "Bank Transfer"}

	// Set up booking distribution
	bookingDistribution := map[string]int{
		"past":    15, // Past bookings
		"current": 10, // Current bookings (ongoing or very recent)
		"future":  25, // Future bookings
	}

	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Temporarily disable triggers for bulk seeding
	if err := tx.Exec("ALTER TABLE bookings DISABLE TRIGGER ALL").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to disable triggers: %w", err)
	}

	// Verify or create LastRunning record for current year
	currentYear := time.Now().Year()
	var lastRunning models.LastRunning

	result := tx.Where("year = ?", currentYear).First(&lastRunning)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create if it doesn't exist
			lastRunning = models.LastRunning{
				LastRunning: 0,
				Year:        currentYear,
			}
			if err := tx.Create(&lastRunning).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create last running record: %w", err)
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("error retrieving last running record: %w", result.Error)
		}
	}

	bookingsCreated := 0

	// Process each booking type sequentially
	for bookingType, count := range bookingDistribution {
		for i := 0; i < count; i++ {
			// Get latest LastRunning value
			if err := tx.Where("year = ?", currentYear).First(&lastRunning).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error getting latest last running: %w", err)
			}

			// Increment lastRunning
			lastRunning.LastRunning++
			bookingID := currentYear*1000000 + lastRunning.LastRunning

			// Save updated LastRunning
			if err := tx.Save(&lastRunning).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error updating last running: %w", err)
			}

			// Select a random room
			randomRoomIndex := rand.Intn(len(rooms))
			room := rooms[randomRoomIndex]

			// Generate booking dates
			checkInDate, checkOutDate, bookingDate := generateBookingDates(bookingType)

			// Get room price
			var roomType models.RoomType
			if err := tx.First(&roomType, "type_id = ?", room.TypeID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error retrieving room type: %w", err)
			}

			// Calculate total price
			nights := int(checkOutDate.Sub(checkInDate).Hours() / 24)
			if nights < 1 {
				nights = 1
			}
			totalPrice := roomType.PricePerNight * nights

			// Create booking
			booking := models.Booking{
				BookingID:    bookingID,
				BookingName:  guestNames[rand.Intn(len(guestNames))],
				RoomNum:      room.RoomNum,
				CheckInDate:  checkInDate,
				CheckOutDate: checkOutDate,
				BookingDate:  bookingDate,
				TotalPrice:   totalPrice,
			}

			// Create the booking
			if err := tx.Create(&booking).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating booking: %w", err)
			}

			// Create receipt for past bookings
			if bookingType == "past" || (bookingType == "current" && checkInDate.Before(time.Now())) {
				receiptID := 10000 + bookingID%1000
				paymentDate := bookingDate.AddDate(0, 0, rand.Intn(3)) // Payment 0-2 days after booking

				receipt := models.Receipt{
					ReceiptID:     receiptID,
					BookingID:     booking.BookingID,
					PaymentDate:   paymentDate,
					PaymentMethod: paymentMethods[rand.Intn(len(paymentMethods))],
					Amount:        totalPrice,
					IssueDate:     paymentDate,
				}

				if err := tx.Create(&receipt).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("error creating receipt: %w", err)
				}
			}

			// Update room status for each day of the booking
			current := checkInDate
			for current.Before(checkOutDate) {
				status := models.RoomStatus{
					RoomNum:   room.RoomNum,
					Calendar:  current,
					Status:    "Occupied",
					BookingID: &booking.BookingID,
				}

				// Use upsert to handle conflicts
				if err := tx.Exec(`
					INSERT INTO room_statuses (room_num, calendar, status, booking_id)
					VALUES (?, ?, ?, ?)
					ON CONFLICT (room_num, calendar) 
					DO UPDATE SET status = EXCLUDED.status, booking_id = EXCLUDED.booking_id
				`, status.RoomNum, status.Calendar, status.Status, status.BookingID).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("error updating room status: %w", err)
				}

				current = current.AddDate(0, 0, 1)
			}

			bookingsCreated++
		}
	}

	// Re-enable triggers
	if err := tx.Exec("ALTER TABLE bookings ENABLE TRIGGER ALL").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to re-enable triggers: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully created %d bookings with corresponding room statuses and receipts", bookingsCreated)
	return nil
}

// Function to simulate and test conflict scenarios
func seedConflictScenarios(db *gorm.DB) error {
	log.Println("Setting up booking conflict test scenarios...")

	// Start a new transaction
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction for conflict scenarios: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Conflict scenario failed with panic: %v", r)
		}
	}()

	// 1. Find a room with no future bookings
	var room models.Room
	if err := tx.Joins("LEFT JOIN bookings ON rooms.room_num = bookings.room_num AND bookings.check_out_date > ?",
		time.Now().AddDate(0, 1, 0)).
		Where("bookings.booking_id IS NULL").
		First(&room).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("could not find room for conflict scenario: %w", err)
	}

	// 2. Set up conflicting dates
	startDate := time.Now().AddDate(0, 2, 0) // Two months from now
	endDate := startDate.AddDate(0, 0, 3)    // 3-day stay

	// Get current year for booking ID generation
	currentYear := time.Now().Year()

	// Get the last running number
	var lastRunning models.LastRunning
	if err := tx.Where("year = ?", currentYear).First(&lastRunning).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("could not get last running record: %w", err)
	}

	// Create first booking
	lastRunning.LastRunning++
	bookingID1 := currentYear*1000000 + lastRunning.LastRunning

	if err := tx.Save(&lastRunning).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("could not update last running: %w", err)
	}

	// Disable triggers temporarily for our test
	if err := tx.Exec("ALTER TABLE bookings DISABLE TRIGGER ALL").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to disable triggers: %w", err)
	}

	booking1 := models.Booking{
		BookingID:    bookingID1,
		BookingName:  "Conflict Test 1",
		RoomNum:      room.RoomNum,
		CheckInDate:  startDate,
		CheckOutDate: endDate,
		BookingDate:  time.Now(),
		TotalPrice:   1000, // Placeholder price
	}

	// Create the first booking
	if err := tx.Create(&booking1).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("could not create first test booking: %w", err)
	}
	log.Println("Created first test booking successfully")

	// Update room statuses for first booking
	current := startDate
	for current.Before(endDate) {
		if err := tx.Exec(`
			INSERT INTO room_statuses (room_num, calendar, status, booking_id)
			VALUES (?, ?, ?, ?)
			ON CONFLICT (room_num, calendar) 
			DO UPDATE SET status = EXCLUDED.status, booking_id = EXCLUDED.booking_id
		`, room.RoomNum, current, "Occupied", booking1.BookingID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating room status: %w", err)
		}
		current = current.AddDate(0, 0, 1)
	}

	// Create second booking with same dates (to demonstrate conflict handling)
	lastRunning.LastRunning++
	bookingID2 := currentYear*1000000 + lastRunning.LastRunning

	if err := tx.Save(&lastRunning).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("could not update last running for second booking: %w", err)
	}

	booking2 := models.Booking{
		BookingID:    bookingID2,
		BookingName:  "Conflict Test 2",
		RoomNum:      room.RoomNum,
		CheckInDate:  startDate,
		CheckOutDate: endDate,
		BookingDate:  time.Now(),
		TotalPrice:   1000,
	}

	// Create second booking (this would normally fail due to the unique constraint)
	// We're using it to demonstrate conflict detection in your app
	if err := tx.Create(&booking2).Error; err != nil {
		log.Printf("Second booking creation failed (expected in real production): %v", err)
	} else {
		log.Println("Created second conflicting test booking")
	}

	// Re-enable triggers
	if err := tx.Exec("ALTER TABLE bookings ENABLE TRIGGER ALL").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to re-enable triggers: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit conflict scenario transaction: %w", err)
	}

	log.Printf("Successfully set up conflict scenario for room %d from %s to %s",
		room.RoomNum, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	return nil
}
