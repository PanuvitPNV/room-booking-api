// pkg/migration/migration.go
package main

import (
	"log"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
)

func main() {
	conf := config.ConfigGetting()
	db := databases.NewPostgresDatabase(conf.Database)
	gormDB := db.Connect()

	// Drop existing triggers if they exist
	gormDB.Exec("DROP TRIGGER IF EXISTS update_room_status ON bookings")
	gormDB.Exec("DROP FUNCTION IF EXISTS update_room_status_for_booking()")

	// Drop existing tables in correct order (because of foreign key dependencies)
	if err := gormDB.Migrator().DropTable(
		&models.RoomStatus{},
		&models.Booking{},
		&models.Guest{},
		&models.Room{},
		&models.RoomType{},
	); err != nil {
		log.Printf("Error dropping tables: %v", err)
	}

	// Disable foreign key check during migrations
	gormDB.Exec("SET CONSTRAINTS ALL DEFERRED")

	// Create room_types table
	if err := gormDB.Table("room_types").AutoMigrate(&models.RoomType{}); err != nil {
		log.Fatalf("Error creating room_types table: %v", err)
	}

	// Create rooms table with foreign key
	if err := gormDB.Table("rooms").AutoMigrate(&models.Room{}); err != nil {
		log.Fatalf("Error creating rooms table: %v", err)
	}

	// Add foreign key constraint
	if err := gormDB.Exec(`ALTER TABLE rooms 
       ADD CONSTRAINT fk_rooms_room_type 
       FOREIGN KEY (type_id) 
       REFERENCES room_types(type_id)`).Error; err != nil {
		log.Fatalf("Error adding foreign key to rooms: %v", err)
	}

	// Create guests table
	if err := gormDB.Table("guests").AutoMigrate(&models.Guest{}); err != nil {
		log.Fatalf("Error creating guests table: %v", err)
	}

	// Create bookings table with foreign keys
	if err := gormDB.Table("bookings").AutoMigrate(&models.Booking{}); err != nil {
		log.Fatalf("Error creating bookings table: %v", err)
	}

	// Add foreign key constraints
	if err := gormDB.Exec(`ALTER TABLE bookings 
       ADD CONSTRAINT fk_bookings_room 
       FOREIGN KEY (room_num) 
       REFERENCES rooms(room_num)`).Error; err != nil {
		log.Fatalf("Error adding foreign key to bookings (room): %v", err)
	}

	if err := gormDB.Exec(`ALTER TABLE bookings 
       ADD CONSTRAINT fk_bookings_guest 
       FOREIGN KEY (guest_id) 
       REFERENCES guests(guest_id)`).Error; err != nil {
		log.Fatalf("Error adding foreign key to bookings (guest): %v", err)
	}

	// Create room_statuses table with foreign keys
	if err := gormDB.Table("room_statuses").AutoMigrate(&models.RoomStatus{}); err != nil {
		log.Fatalf("Error creating room_statuses table: %v", err)
	}

	// Add foreign key constraints
	if err := gormDB.Exec(`ALTER TABLE room_statuses 
       ADD CONSTRAINT fk_room_statuses_room 
       FOREIGN KEY (room_num) 
       REFERENCES rooms(room_num)`).Error; err != nil {
		log.Fatalf("Error adding foreign key to room_statuses (room): %v", err)
	}

	if err := gormDB.Exec(`ALTER TABLE room_statuses 
       ADD CONSTRAINT fk_room_statuses_booking 
       FOREIGN KEY (booking_id) 
       REFERENCES bookings(booking_id)`).Error; err != nil {
		log.Fatalf("Error adding foreign key to room_statuses (booking): %v", err)
	}

	// Create function and trigger for automatic room status updates
	createRoomStatusFunction := `
   CREATE OR REPLACE FUNCTION update_room_status_for_booking()
   RETURNS TRIGGER AS $$
   BEGIN
       -- When a new booking is created or updated
       INSERT INTO room_statuses (room_num, calendar, status, booking_id)
       SELECT 
           NEW.room_num,
           generate_series(
               NEW.check_in_date::date,
               NEW.check_out_date::date - INTERVAL '1 day',
               INTERVAL '1 day'
           )::date,
           'Occupied',
           NEW.booking_id;

       RETURN NEW;
   END;
   $$ LANGUAGE plpgsql;
   `

	createTrigger := `
   CREATE TRIGGER update_room_status
   AFTER INSERT OR UPDATE ON bookings
   FOR EACH ROW
   EXECUTE FUNCTION update_room_status_for_booking();
   `

	// Execute the function and trigger creation
	if err := gormDB.Exec(createRoomStatusFunction).Error; err != nil {
		log.Fatalf("Error creating room status function: %v", err)
	}

	if err := gormDB.Exec(createTrigger).Error; err != nil {
		log.Fatalf("Error creating room status trigger: %v", err)
	}

	// Re-enable foreign key checks
	gormDB.Exec("SET CONSTRAINTS ALL IMMEDIATE")

	log.Println("Migration completed successfully!")
}
