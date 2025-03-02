package data

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/panuvitpnv/room-booking-api/internal/models"
)

// SeedDatabase populates the database with initial data
func SeedDatabase(db *gorm.DB) error {
	log.Println("Seeding database...")

	// Create room types
	roomTypes := []models.RoomType{
		{
			TypeID:        1,
			Name:          "Standard",
			Description:   "A comfortable standard room with essential amenities",
			Area:          25,
			PricePerNight: 1000,
			NoOfGuest:     2,
		},
		{
			TypeID:        2,
			Name:          "Deluxe",
			Description:   "A spacious deluxe room with premium amenities",
			Area:          35,
			PricePerNight: 1500,
			NoOfGuest:     2,
		},
		{
			TypeID:        3,
			Name:          "Suite",
			Description:   "A luxurious suite with separate living area",
			Area:          50,
			PricePerNight: 2500,
			NoOfGuest:     4,
		},
		{
			TypeID:        4,
			Name:          "Family Room",
			Description:   "A large room suitable for families with children",
			Area:          45,
			PricePerNight: 2000,
			NoOfGuest:     4,
		},
	}

	// Create facilities
	facilities := []models.Facility{
		{FacilityID: 1, Name: "Wi-Fi"},
		{FacilityID: 2, Name: "Air Conditioning"},
		{FacilityID: 3, Name: "TV"},
		{FacilityID: 4, Name: "Mini Bar"},
		{FacilityID: 5, Name: "Coffee Machine"},
		{FacilityID: 6, Name: "Balcony"},
		{FacilityID: 7, Name: "Bathtub"},
		{FacilityID: 8, Name: "Kitchen"},
	}

	// Create room-facility mappings
	roomFacilities := []models.RoomFacility{
		{TypeID: 1, FacilityID: 1}, // Standard has Wi-Fi
		{TypeID: 1, FacilityID: 2}, // Standard has Air Conditioning
		{TypeID: 1, FacilityID: 3}, // Standard has TV

		{TypeID: 2, FacilityID: 1}, // Deluxe has Wi-Fi
		{TypeID: 2, FacilityID: 2}, // Deluxe has Air Conditioning
		{TypeID: 2, FacilityID: 3}, // Deluxe has TV
		{TypeID: 2, FacilityID: 4}, // Deluxe has Mini Bar
		{TypeID: 2, FacilityID: 5}, // Deluxe has Coffee Machine

		{TypeID: 3, FacilityID: 1}, // Suite has Wi-Fi
		{TypeID: 3, FacilityID: 2}, // Suite has Air Conditioning
		{TypeID: 3, FacilityID: 3}, // Suite has TV
		{TypeID: 3, FacilityID: 4}, // Suite has Mini Bar
		{TypeID: 3, FacilityID: 5}, // Suite has Coffee Machine
		{TypeID: 3, FacilityID: 6}, // Suite has Balcony
		{TypeID: 3, FacilityID: 7}, // Suite has Bathtub

		{TypeID: 4, FacilityID: 1}, // Family Room has Wi-Fi
		{TypeID: 4, FacilityID: 2}, // Family Room has Air Conditioning
		{TypeID: 4, FacilityID: 3}, // Family Room has TV
		{TypeID: 4, FacilityID: 8}, // Family Room has Kitchen
	}

	// Create rooms
	rooms := []models.Room{}

	// Create 5 Standard rooms (101-105)
	for i := 1; i <= 5; i++ {
		rooms = append(rooms, models.Room{
			RoomNum: 100 + i,
			TypeID:  1,
		})
	}

	// Create 5 Deluxe rooms (201-205)
	for i := 1; i <= 5; i++ {
		rooms = append(rooms, models.Room{
			RoomNum: 200 + i,
			TypeID:  2,
		})
	}

	// Create 3 Suite rooms (301-303)
	for i := 1; i <= 3; i++ {
		rooms = append(rooms, models.Room{
			RoomNum: 300 + i,
			TypeID:  3,
		})
	}

	// Create 3 Family rooms (401-403)
	for i := 1; i <= 3; i++ {
		rooms = append(rooms, models.Room{
			RoomNum: 400 + i,
			TypeID:  4,
		})
	}

	// Initialize Last Running number for IDs
	lastRunning := models.LastRunning{
		LastRunning: 1000,
		Year:        time.Now().Year(),
	}

	// Use transactions to ensure data consistency during seeding
	return db.Transaction(func(tx *gorm.DB) error {
		// Check if data already exists
		var count int64
		tx.Model(&models.RoomType{}).Count(&count)
		if count > 0 {
			log.Println("Database already has data, skipping seed")
			return nil
		}

		// Create room types with error handling
		for _, roomType := range roomTypes {
			if err := tx.Create(&roomType).Error; err != nil {
				return fmt.Errorf("failed to create room type: %w", err)
			}
		}
		log.Println("Room types created")

		// Create facilities
		for _, facility := range facilities {
			if err := tx.Create(&facility).Error; err != nil {
				return fmt.Errorf("failed to create facility: %w", err)
			}
		}
		log.Println("Facilities created")

		// Create room-facility mappings
		for _, rf := range roomFacilities {
			if err := tx.Create(&rf).Error; err != nil {
				return fmt.Errorf("failed to create room-facility mapping: %w", err)
			}
		}
		log.Println("Room-facility mappings created")

		// Create rooms
		for _, room := range rooms {
			if err := tx.Create(&room).Error; err != nil {
				return fmt.Errorf("failed to create room: %w", err)
			}
		}
		log.Println("Rooms created")

		// Initialize LastRunning
		if err := tx.Create(&lastRunning).Error; err != nil {
			return fmt.Errorf("failed to initialize last running number: %w", err)
		}
		log.Println("Last running number initialized")

		return nil
	})
}
