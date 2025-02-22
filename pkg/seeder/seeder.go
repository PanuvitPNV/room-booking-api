package main

import (
	"log"
	"time"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
)

func main() {
	conf := config.ConfigGetting()
	db := databases.NewPostgresDatabase(conf.Database)
	gormDB := db.Connect()

	log.Println("Starting data seeding...")

	// Clean existing data
	gormDB.Exec("TRUNCATE room_statuses, bookings, guests, rooms, room_types CASCADE")

	// 1. Seed RoomTypes
	roomTypes := []models.RoomType{
		{
			Name:          "Standard Twin",
			Description:   "Two single beds with city view",
			Area:          25,
			Highlight:     "Perfect for friends traveling together",
			Facility:      "TV, Air Conditioning, Wi-Fi, Mini Fridge",
			PricePerNight: 1500,
			Capacity:      2,
		},
		{
			Name:          "Deluxe King",
			Description:   "Large room with king-size bed",
			Area:          35,
			Highlight:     "Romantic mountain view",
			Facility:      "TV, Air Conditioning, Wi-Fi, Mini Bar, Coffee Maker, Bathtub",
			PricePerNight: 2500,
			Capacity:      2,
		},
		{
			Name:          "Family Suite",
			Description:   "Two bedrooms with living area",
			Area:          50,
			Highlight:     "Spacious layout perfect for families",
			Facility:      "2 TVs, Air Conditioning, Wi-Fi, Kitchen, Dining Area, Washing Machine",
			PricePerNight: 4000,
			Capacity:      4,
		},
	}

	if err := gormDB.Create(&roomTypes).Error; err != nil {
		log.Fatalf("Error seeding room types: %v", err)
	}
	log.Println("Room types seeded successfully")

	// 2. Seed Rooms (this will trigger automatic room status creation for 1 year)
	rooms := []models.Room{
		{RoomNum: 201, TypeID: roomTypes[0].TypeID}, // Standard Twin
		{RoomNum: 202, TypeID: roomTypes[0].TypeID}, // Standard Twin
		{RoomNum: 301, TypeID: roomTypes[1].TypeID}, // Deluxe King
		{RoomNum: 302, TypeID: roomTypes[1].TypeID}, // Deluxe King
		{RoomNum: 401, TypeID: roomTypes[2].TypeID}, // Family Suite
	}

	// Create rooms one by one to ensure trigger works properly
	for _, room := range rooms {
		if err := gormDB.Create(&room).Error; err != nil {
			log.Printf("Error seeding room %d: %v", room.RoomNum, err)
		}
	}
	log.Println("Rooms seeded successfully")

	// 3. Seed Guests
	guests := []models.Guest{
		{
			FirstName:   "Alice",
			LastName:    "Johnson",
			DateOfBirth: time.Date(1990, 3, 15, 0, 0, 0, 0, time.UTC),
			Email:       "alice.j@example.com",
			Phone:       "111-222-3333",
		},
		{
			FirstName:   "Bob",
			LastName:    "Smith",
			DateOfBirth: time.Date(1985, 7, 22, 0, 0, 0, 0, time.UTC),
			Email:       "bob.smith@example.com",
			Phone:       "444-555-6666",
		},
		{
			FirstName:   "Carol",
			LastName:    "Williams",
			DateOfBirth: time.Date(1992, 12, 5, 0, 0, 0, 0, time.UTC),
			Email:       "carol.w@example.com",
			Phone:       "777-888-9999",
		},
	}

	if err := gormDB.Create(&guests).Error; err != nil {
		log.Fatalf("Error seeding guests: %v", err)
	}
	log.Println("Guests seeded successfully")

	// 4. Create bookings (will automatically update room statuses)
	// Use dates within the current year
	currentYear := time.Now().Year()
	bookings := []models.Booking{
		{
			RoomNum:      201,
			GuestID:      guests[0].GuestID,
			CheckInDate:  time.Date(currentYear, 2, 15, 14, 0, 0, 0, time.UTC),
			CheckOutDate: time.Date(currentYear, 2, 18, 12, 0, 0, 0, time.UTC),
			TotalPrice:   roomTypes[0].PricePerNight * 3,
		},
		{
			RoomNum:      301,
			GuestID:      guests[1].GuestID,
			CheckInDate:  time.Date(currentYear, 3, 1, 14, 0, 0, 0, time.UTC),
			CheckOutDate: time.Date(currentYear, 3, 3, 12, 0, 0, 0, time.UTC),
			TotalPrice:   roomTypes[1].PricePerNight * 2,
		},
		{
			RoomNum:      401,
			GuestID:      guests[2].GuestID,
			CheckInDate:  time.Date(currentYear, 4, 10, 14, 0, 0, 0, time.UTC),
			CheckOutDate: time.Date(currentYear, 4, 15, 12, 0, 0, 0, time.UTC),
			TotalPrice:   roomTypes[2].PricePerNight * 5,
		},
	}

	for _, booking := range bookings {
		if err := gormDB.Create(&booking).Error; err != nil {
			log.Printf("Error creating booking: %v", err)
		}
	}
	log.Println("Bookings seeded successfully")

	// Print summary of seeded data
	var roomTypeCount, roomCount, guestCount, bookingCount, statusCount int64
	gormDB.Model(&models.RoomType{}).Count(&roomTypeCount)
	gormDB.Model(&models.Room{}).Count(&roomCount)
	gormDB.Model(&models.Guest{}).Count(&guestCount)
	gormDB.Model(&models.Booking{}).Count(&bookingCount)
	gormDB.Model(&models.RoomStatus{}).Count(&statusCount)

	log.Printf("\nSeeded Data Summary:")
	log.Printf("Room Types: %d", roomTypeCount)
	log.Printf("Rooms: %d", roomCount)
	log.Printf("Guests: %d", guestCount)
	log.Printf("Bookings: %d", bookingCount)
	log.Printf("Room Statuses: %d (should be rooms × 365)", statusCount)

	// Verify room statuses
	var availableCount, occupiedCount int64
	gormDB.Model(&models.RoomStatus{}).Where("status = ?", "Available").Count(&availableCount)
	gormDB.Model(&models.RoomStatus{}).Where("status = ?", "Occupied").Count(&occupiedCount)
	log.Printf("Available Room Statuses: %d", availableCount)
	log.Printf("Occupied Room Statuses: %d", occupiedCount)
}
