package main

import (
	"fmt"
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

	// Test 1: Check Room Types
	fmt.Println("\n=== Test 1: Room Types ===")
	var roomTypes []models.RoomType
	if err := gormDB.Find(&roomTypes).Error; err != nil {
		log.Fatalf("Error fetching room types: %v", err)
	}
	for _, rt := range roomTypes {
		fmt.Printf("Type ID: %d, Name: %s, Price: %d, Capacity: %d\n",
			rt.TypeID, rt.Name, rt.PricePerNight, rt.Capacity)
	}

	// Test 2: Check Rooms and their Types
	fmt.Println("\n=== Test 2: Rooms with Types ===")
	var rooms []models.Room
	if err := gormDB.Preload("RoomType").Find(&rooms).Error; err != nil {
		log.Fatalf("Error fetching rooms: %v", err)
	}
	for _, room := range rooms {
		fmt.Printf("Room %d is a %s\n", room.RoomNum, room.RoomType.Name)
	}

	// Test 3: Check Guests
	fmt.Println("\n=== Test 3: Guests ===")
	var guests []models.Guest
	if err := gormDB.Find(&guests).Error; err != nil {
		log.Fatalf("Error fetching guests: %v", err)
	}
	for _, guest := range guests {
		fmt.Printf("Guest ID: %d, Name: %s %s, Email: %s\n",
			guest.GuestID, guest.FirstName, guest.LastName, guest.Email)
	}

	// Test 4: Check Existing Bookings
	fmt.Println("\n=== Test 4: Current Bookings ===")
	var bookings []models.Booking
	if err := gormDB.Preload("Room").Preload("Guest").Find(&bookings).Error; err != nil {
		log.Fatalf("Error fetching bookings: %v", err)
	}
	for _, booking := range bookings {
		fmt.Printf("Booking ID: %d\n", booking.BookingID)
		fmt.Printf("Room: %d\n", booking.RoomNum)
		fmt.Printf("Guest: %s %s\n", booking.Guest.FirstName, booking.Guest.LastName)
		fmt.Printf("Check-in: %v\n", booking.CheckInDate.Format("2006-01-02"))
		fmt.Printf("Check-out: %v\n", booking.CheckOutDate.Format("2006-01-02"))
		fmt.Printf("Total Price: %d\n\n", booking.TotalPrice)
	}

	// Test 5: Check Room Status for next 7 days
	fmt.Println("\n=== Test 5: Room Status Next 7 Days ===")
	var roomStatuses []models.RoomStatus
	if err := gormDB.Where("calendar BETWEEN ? AND ?",
		time.Now(),
		time.Now().AddDate(0, 0, 7)).
		Order("room_num, calendar").
		Find(&roomStatuses).Error; err != nil {
		log.Fatalf("Error fetching room statuses: %v", err)
	}
	for _, status := range roomStatuses {
		fmt.Printf("Room %d on %v: %s\n",
			status.RoomNum,
			status.Calendar.Format("2006-01-02"),
			status.Status)
	}

	// Test 6: Try to create an overlapping booking (should fail)
	fmt.Println("\n=== Test 6: Overlapping Booking Test ===")
	newBooking := models.Booking{
		RoomNum:      201, // Try to book a room that's already booked
		GuestID:      1,
		CheckInDate:  time.Now().AddDate(0, 0, 1),
		CheckOutDate: time.Now().AddDate(0, 0, 2),
		TotalPrice:   1500,
	}
	err := gormDB.Create(&newBooking).Error
	if err != nil {
		fmt.Printf("Expected error occurred (this is good): %v\n", err)
	} else {
		fmt.Println("Warning: Overlapping booking was created!")
	}

	// Test 7: Check Available Rooms for specific dates
	fmt.Println("\n=== Test 7: Available Rooms for Next Week ===")
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := checkIn.AddDate(0, 0, 3)

	var availableRooms []models.Room
	if err := gormDB.
		Joins("LEFT JOIN room_statuses ON rooms.room_num = room_statuses.room_num").
		Where("room_statuses.room_num IS NULL OR room_statuses.status = ? OR room_statuses.calendar NOT BETWEEN ? AND ?",
			"Available", checkIn, checkOut).
		Preload("RoomType").
		Distinct().
		Find(&availableRooms).Error; err != nil {
		log.Fatalf("Error checking available rooms: %v", err)
	}

	fmt.Printf("Available rooms for %v to %v:\n",
		checkIn.Format("2006-01-02"),
		checkOut.Format("2006-01-02"))
	for _, room := range availableRooms {
		fmt.Printf("Room %d (%s) - %d per night\n",
			room.RoomNum,
			room.RoomType.Name,
			room.RoomType.PricePerNight)
	}

	fmt.Println("\nAll tests completed!")
}
