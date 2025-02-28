package main

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	// Initialize configuration and database connection
	conf := config.ConfigGetting()
	db := databases.NewPostgresDatabase(conf.Database)
	gormDB := db.Connect()

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	log.Println("Starting database seeding...")

	// Begin transaction
	tx := gormDB.Begin()
	if tx.Error != nil {
		log.Fatalf("Failed to begin transaction: %v", tx.Error)
	}

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

	// Seed sample bookings and receipts
	if err := seedBookingsAndReceipts(tx, rooms); err != nil {
		tx.Rollback()
		log.Fatalf("Error seeding bookings and receipts: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("Database seeding completed successfully!")
	log.Println("Seeded:")
	log.Println("- Room types, facilities, and their relationships")
	log.Println("- Rooms assigned to different room types")
	log.Println("- Sample bookings with receipts")
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

	for _, roomType := range roomTypes {
		if err := tx.Create(&roomType).Error; err != nil {
			return nil, err
		}
	}

	log.Printf("Seeded %d room types", len(roomTypes))
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

	for _, facility := range facilities {
		if err := tx.Create(&facility).Error; err != nil {
			return nil, err
		}
	}

	log.Printf("Seeded %d facilities", len(facilities))
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
			roomFacility := models.RoomFacility{
				TypeID:     roomTypeID,
				FacilityID: facilityID,
			}
			roomFacilities = append(roomFacilities, roomFacility)
		}
	}

	for _, roomFacility := range roomFacilities {
		if err := tx.Create(&roomFacility).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded %d room facility relationships", len(roomFacilities))
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
			room := models.Room{
				RoomNum: roomNum,
				TypeID:  typeID,
			}

			if err := tx.Create(&room).Error; err != nil {
				return nil, err
			}

			rooms = append(rooms, room)
			roomNum++
		}
	}

	log.Printf("Seeded %d rooms", len(rooms))
	return rooms, nil
}

func seedLastRunning(tx *gorm.DB) error {
	lastRunning := models.LastRunning{
		LastRunning: 0,
		Year:        time.Now().Year(),
	}

	if err := tx.Create(&lastRunning).Error; err != nil {
		return err
	}

	log.Println("Seeded last running record")
	return nil
}

func seedBookingsAndReceipts(tx *gorm.DB, rooms []models.Room) error {
	// Generate 20 sample bookings across different rooms
	numBookings := 20
	bookings := make([]models.Booking, 0, numBookings)

	// Generate some past, current, and future bookings
	now := time.Now()
	currentYear := now.Year()
	currentMonth := now.Month()

	guestNames := []string{
		"John Smith", "Jane Doe", "Michael Johnson", "Emily Davis",
		"David Wilson", "Sarah Brown", "Robert Taylor", "Jessica Miller",
		"Thomas Moore", "Jennifer Anderson", "William White", "Lisa Martinez",
		"Daniel Clark", "Mary Rodriguez", "James Lewis", "Patricia Allen",
		"Christopher Young", "Barbara Hall", "Matthew King", "Elizabeth Scott",
	}

	paymentMethods := []string{"Credit", "Debit", "Bank Transfer"}

	// First, check if any bookings already exist
	var existingBookingCount int64
	if err := tx.Model(&models.Booking{}).Count(&existingBookingCount).Error; err != nil {
		return err
	}

	if existingBookingCount > 0 {
		log.Printf("Skipping booking seeding as %d bookings already exist", existingBookingCount)
		return nil
	}

	// Get or create LastRunning for the current year
	var lastRunning models.LastRunning
	result := tx.Where("year = ?", currentYear).First(&lastRunning)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new LastRunning record if it doesn't exist
		lastRunning = models.LastRunning{
			LastRunning: 0,
			Year:        currentYear,
		}
		if err := tx.Create(&lastRunning).Error; err != nil {
			return err
		}
	}

	for i := 0; i < numBookings; i++ {
		// Random room selection
		room := rooms[rand.Intn(len(rooms))]

		// Get room price
		var roomType models.RoomType
		if err := tx.First(&roomType, "type_id = ?", room.TypeID).Error; err != nil {
			return err
		}

		// Create random date range
		var checkInDate, checkOutDate time.Time

		if i < 5 {
			// Past bookings
			monthOffset := rand.Intn(6) // 0-5 months ago
			checkInDate = time.Date(currentYear, currentMonth-time.Month(monthOffset), 1+rand.Intn(15), 14, 0, 0, 0, time.Local)
			stayDuration := 1 + rand.Intn(5) // 1-5 nights
			checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
		} else if i < 15 {
			// Current and near future bookings
			daysOffset := rand.Intn(60) - 10 // -10 to 50 days from now
			checkInDate = now.AddDate(0, 0, daysOffset)
			stayDuration := 1 + rand.Intn(7) // 1-7 nights
			checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
		} else {
			// Far future bookings
			monthOffset := 2 + rand.Intn(4) // 2-5 months in the future
			checkInDate = time.Date(currentYear, currentMonth+time.Month(monthOffset), 1+rand.Intn(20), 14, 0, 0, 0, time.Local)
			stayDuration := 1 + rand.Intn(10) // 1-10 nights
			checkOutDate = checkInDate.AddDate(0, 0, stayDuration)
		}

		// Calculate booking ID (YYYYXXXXXX format)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lastRunning, "year = ?", currentYear).Error; err != nil {
			return err
		}

		lastRunning.LastRunning++
		if err := tx.Save(&lastRunning).Error; err != nil {
			return err
		}

		bookingID := currentYear*1000000 + lastRunning.LastRunning

		// Calculate total price
		nights := int(checkOutDate.Sub(checkInDate).Hours() / 24)
		if nights < 1 {
			nights = 1
		}
		totalPrice := roomType.PricePerNight * nights

		// Check if booking with this ID already exists
		var existingBooking models.Booking
		result := tx.Where("booking_id = ?", bookingID).First(&existingBooking)
		if result.Error == nil {
			// Skip this booking if ID already exists
			log.Printf("Skipping booking with ID %d as it already exists", bookingID)
			continue
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		// Create booking
		booking := models.Booking{
			BookingID:    bookingID,
			BookingName:  guestNames[i%len(guestNames)],
			RoomNum:      room.RoomNum,
			CheckInDate:  checkInDate,
			CheckOutDate: checkOutDate,
			BookingDate:  time.Now().AddDate(0, 0, -rand.Intn(30)), // Booking made 0-30 days ago
			TotalPrice:   totalPrice,
		}

		// Disable triggers temporarily for seeding
		if err := tx.Exec("ALTER TABLE bookings DISABLE TRIGGER ALL").Error; err != nil {
			return err
		}

		if err := tx.Create(&booking).Error; err != nil {
			return err
		}

		// Re-enable triggers
		if err := tx.Exec("ALTER TABLE bookings ENABLE TRIGGER ALL").Error; err != nil {
			return err
		}

		bookings = append(bookings, booking)

		// Update room status for each day of the booking
		current := checkInDate
		for current.Before(checkOutDate) || current.Equal(checkOutDate) {
			status := models.RoomStatus{
				RoomNum:   room.RoomNum,
				Calendar:  current,
				Status:    "Occupied",
				BookingID: &booking.BookingID,
			}

			// Use upsert to handle potential existing records
			if err := tx.Exec(`
                INSERT INTO room_statuses (room_num, calendar, status, booking_id)
                VALUES (?, ?, ?, ?)
                ON CONFLICT (room_num, calendar) 
                DO UPDATE SET status = EXCLUDED.status, booking_id = EXCLUDED.booking_id
            `, status.RoomNum, status.Calendar, status.Status, status.BookingID).Error; err != nil {
				return err
			}

			current = current.AddDate(0, 0, 1)
		}

		// Create receipt for past and some current bookings
		if checkInDate.Before(now) {
			receiptID := 10000 + i
			paymentDate := booking.BookingDate.AddDate(0, 0, rand.Intn(3)) // Payment 0-2 days after booking

			// Check if receipt with this ID already exists
			var existingReceipt models.Receipt
			result := tx.Where("receipt_id = ?", receiptID).First(&existingReceipt)
			if result.Error == nil {
				// Skip creating receipt if ID already exists
				continue
			}

			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return result.Error
			}

			receipt := models.Receipt{
				ReceiptID:     receiptID,
				BookingID:     booking.BookingID,
				PaymentDate:   paymentDate,
				PaymentMethod: paymentMethods[rand.Intn(len(paymentMethods))],
				Amount:        totalPrice,
				IssueDate:     paymentDate,
			}

			if err := tx.Create(&receipt).Error; err != nil {
				return err
			}
		}
	}

	log.Printf("Seeded %d bookings with corresponding room statuses and receipts", len(bookings))
	return nil
}
