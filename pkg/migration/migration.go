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

	log.Println("Starting database migration...")

	// Drop all existing triggers and functions first
	dropStatements := []string{
		"DROP TRIGGER IF EXISTS update_room_status ON bookings CASCADE",
		"DROP TRIGGER IF EXISTS populate_room_status ON rooms CASCADE",
		"DROP FUNCTION IF EXISTS update_room_status_for_booking() CASCADE",
		"DROP FUNCTION IF EXISTS populate_room_status_for_new_room() CASCADE",
	}

	for _, stmt := range dropStatements {
		if err := gormDB.Exec(stmt).Error; err != nil {
			log.Printf("Warning while executing %s: %v", stmt, err)
		}
	}

	// Drop existing tables in correct order
	tables := []interface{}{
		&models.RoomStatus{},
		&models.Booking{},
		&models.Guest{},
		&models.Room{},
		&models.RoomType{},
	}

	for _, table := range tables {
		if err := gormDB.Migrator().DropTable(table); err != nil {
			log.Printf("Error dropping table: %v", err)
		}
	}

	log.Println("Dropped existing tables and triggers")

	// Disable foreign key check during migrations
	gormDB.Exec("SET CONSTRAINTS ALL DEFERRED")

	// Create tables
	log.Println("Creating tables...")

	// Create room_types table
	if err := gormDB.AutoMigrate(&models.RoomType{}); err != nil {
		log.Fatalf("Error creating room_types table: %v", err)
	}

	// Create rooms table with foreign key
	if err := gormDB.AutoMigrate(&models.Room{}); err != nil {
		log.Fatalf("Error creating rooms table: %v", err)
	}

	// Add foreign key constraint for rooms
	if err := gormDB.Exec(`ALTER TABLE rooms 
        ADD CONSTRAINT fk_rooms_room_type 
        FOREIGN KEY (type_id) 
        REFERENCES room_types(type_id)`).Error; err != nil {
		log.Fatalf("Error adding foreign key to rooms: %v", err)
	}

	// Create guests table
	if err := gormDB.AutoMigrate(&models.Guest{}); err != nil {
		log.Fatalf("Error creating guests table: %v", err)
	}

	// Create bookings table
	if err := gormDB.AutoMigrate(&models.Booking{}); err != nil {
		log.Fatalf("Error creating bookings table: %v", err)
	}

	// Add foreign key constraints for bookings
	if err := gormDB.Exec(`ALTER TABLE bookings 
        ADD CONSTRAINT fk_bookings_room 
        FOREIGN KEY (room_num) 
        REFERENCES rooms(room_num)`).Error; err != nil {
		log.Fatalf("Error adding room foreign key to bookings: %v", err)
	}

	if err := gormDB.Exec(`ALTER TABLE bookings 
        ADD CONSTRAINT fk_bookings_guest 
        FOREIGN KEY (guest_id) 
        REFERENCES guests(guest_id)`).Error; err != nil {
		log.Fatalf("Error adding guest foreign key to bookings: %v", err)
	}

	// Create room_statuses table
	if err := gormDB.AutoMigrate(&models.RoomStatus{}); err != nil {
		log.Fatalf("Error creating room_statuses table: %v", err)
	}

	// Add foreign key constraints for room_statuses
	if err := gormDB.Exec(`ALTER TABLE room_statuses 
        ADD CONSTRAINT fk_room_statuses_room 
        FOREIGN KEY (room_num) 
        REFERENCES rooms(room_num)`).Error; err != nil {
		log.Fatalf("Error adding room foreign key to room_statuses: %v", err)
	}

	if err := gormDB.Exec(`ALTER TABLE room_statuses 
        ADD CONSTRAINT fk_room_statuses_booking 
        FOREIGN KEY (booking_id) 
        REFERENCES bookings(booking_id)`).Error; err != nil {
		log.Fatalf("Error adding booking foreign key to room_statuses: %v", err)
	}

	log.Println("Tables created successfully")

	log.Println("Creating functions and triggers...")

	// Create function for populating room status for new rooms
	createRoomStatusPopulationFunction := `
    CREATE OR REPLACE FUNCTION populate_room_status_for_new_room()
    RETURNS TRIGGER AS $$
    DECLARE
        start_date DATE;
        end_date DATE;
        current_year INT;
    BEGIN
        -- Get current year
        current_year := EXTRACT(YEAR FROM CURRENT_DATE);
        
        -- Set date range for current year
        start_date := DATE_TRUNC('year', CURRENT_DATE);
        end_date := (DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year' - INTERVAL '1 day')::date;

        -- Insert room status records for the entire year
        INSERT INTO room_statuses (room_num, calendar, status)
        SELECT 
            NEW.room_num,
            generate_series(
                start_date,
                end_date,
                INTERVAL '1 day'
            )::date,
            'Available';

        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;
    `

	// Create function for updating room status when booking is made
	createBookingStatusFunction := `
		CREATE OR REPLACE FUNCTION update_room_status_for_booking()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Update existing room status records for the booking period
			-- Include the check-out date by removing the "< check_out_date" condition
			UPDATE room_statuses
			SET status = 'Occupied',
				booking_id = NEW.booking_id
			WHERE room_num = NEW.room_num
			AND calendar >= NEW.check_in_date::date
			AND calendar <= NEW.check_out_date::date;  -- Changed from < to <=

			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
    `

	// Create triggers
	createRoomTrigger := `
    CREATE TRIGGER populate_room_status
    AFTER INSERT ON rooms
    FOR EACH ROW
    EXECUTE FUNCTION populate_room_status_for_new_room();
    `

	createBookingTrigger := `
    CREATE TRIGGER update_room_status
    AFTER INSERT OR UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_room_status_for_booking();
    `

	// Create indexes
	createIndexes := []string{
		"CREATE INDEX idx_room_status_room_date ON room_statuses(room_num, calendar)",
		"CREATE INDEX idx_bookings_date_range ON bookings(check_in_date, check_out_date)",
		"CREATE INDEX idx_room_status_status ON room_statuses(status)",
	}

	// Execute all SQL statements
	statements := []string{
		createRoomStatusPopulationFunction,
		createBookingStatusFunction,
		createRoomTrigger,
		createBookingTrigger,
	}

	statements = append(statements, createIndexes...)

	for _, stmt := range statements {
		if err := gormDB.Exec(stmt).Error; err != nil {
			log.Fatalf("Error executing statement: %v\nStatement: %s", err, stmt)
		}
	}

	// Add constraints
	constraints := []string{
		// Ensure check_out_date is after check_in_date
		`ALTER TABLE bookings 
			ADD CONSTRAINT check_dates 
			CHECK (check_out_date > check_in_date)`,

		// Ensure booking dates are within 2025
		`ALTER TABLE bookings 
		ADD CONSTRAINT check_booking_year 
		CHECK (EXTRACT(YEAR FROM check_in_date) = EXTRACT(YEAR FROM CURRENT_DATE) 
			AND EXTRACT(YEAR FROM check_out_date) = EXTRACT(YEAR FROM CURRENT_DATE))`,

		// Ensure valid status values
		`ALTER TABLE room_statuses 
			ADD CONSTRAINT check_status 
			CHECK (status IN ('Available', 'Occupied'))`,
	}

	for _, constraint := range constraints {
		if err := gormDB.Exec(constraint).Error; err != nil {
			log.Printf("Warning while adding constraint: %v", err)
		}
	}

	// Re-enable foreign key checks
	gormDB.Exec("SET CONSTRAINTS ALL IMMEDIATE")

	log.Println("Migration completed successfully!")
	log.Println("Created:")
	log.Println("- 5 tables (room_types, rooms, guests, bookings, room_statuses)")
	log.Println("- 2 triggers (populate_room_status, update_room_status)")
	log.Println("- 3 indexes")
	log.Println("- Multiple constraints for data integrity")
}
