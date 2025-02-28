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

	// Drop all existing triggers and functions first to avoid conflicts
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

	// Drop existing tables in correct order to respect foreign key constraints
	tables := []interface{}{
		&models.RoomStatus{},
		&models.Receipt{},
		&models.Booking{},
		&models.Room{},
		&models.RoomFacility{},
		&models.Facility{},
		&models.RoomType{},
		&models.LastRunning{},
	}

	for _, table := range tables {
		if err := gormDB.Migrator().DropTable(table); err != nil {
			log.Printf("Error dropping table: %v", err)
		}
	}

	log.Println("Dropped existing tables and triggers")

	// Start transaction for schema creation
	tx := gormDB.Begin()
	if tx.Error != nil {
		log.Fatalf("Failed to begin transaction: %v", tx.Error)
	}

	// Disable foreign key checks during migrations
	tx.Exec("SET CONSTRAINTS ALL DEFERRED")

	// Create tables
	log.Println("Creating tables...")

	// Create room_types table
	if err := tx.AutoMigrate(&models.RoomType{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating room_types table: %v", err)
	}

	// Create facilities table
	if err := tx.AutoMigrate(&models.Facility{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating facilities table: %v", err)
	}

	// Create room_facilities table
	if err := tx.AutoMigrate(&models.RoomFacility{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating room_facilities table: %v", err)
	}

	// Add foreign key constraints for room_facilities
	if err := tx.Exec(`ALTER TABLE room_facilities 
        ADD CONSTRAINT fk_room_facilities_room_type 
        FOREIGN KEY (type_id) 
        REFERENCES room_types(type_id)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding room_type foreign key to room_facilities: %v", err)
	}

	if err := tx.Exec(`ALTER TABLE room_facilities 
        ADD CONSTRAINT fk_room_facilities_facility 
        FOREIGN KEY (fac_id) 
        REFERENCES facilities(fac_id)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding facility foreign key to room_facilities: %v", err)
	}

	// Create rooms table with foreign key
	if err := tx.AutoMigrate(&models.Room{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating rooms table: %v", err)
	}

	// Add foreign key constraint for rooms
	if err := tx.Exec(`ALTER TABLE rooms 
        ADD CONSTRAINT fk_rooms_room_type 
        FOREIGN KEY (type_id) 
        REFERENCES room_types(type_id)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding foreign key to rooms: %v", err)
	}

	// Create bookings table
	if err := tx.AutoMigrate(&models.Booking{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating bookings table: %v", err)
	}

	// Add foreign key constraints for bookings
	if err := tx.Exec(`ALTER TABLE bookings 
        ADD CONSTRAINT fk_bookings_room 
        FOREIGN KEY (room_num) 
        REFERENCES rooms(room_num)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding room foreign key to bookings: %v", err)
	}

	// Add unique index to ensure no duplicate bookings for same room and dates
	if err := tx.Exec(`CREATE UNIQUE INDEX idx_no_duplicate_booking
        ON bookings(room_num, check_in_date, check_out_date)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding unique index for booking dates: %v", err)
	}

	// Create receipts table
	if err := tx.AutoMigrate(&models.Receipt{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating receipts table: %v", err)
	}

	// Add foreign key constraint for receipts
	if err := tx.Exec(`ALTER TABLE receipts 
        ADD CONSTRAINT fk_receipts_booking 
        FOREIGN KEY (booking_id) 
        REFERENCES bookings(booking_id)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding booking foreign key to receipts: %v", err)
	}

	// Create room_statuses table
	if err := tx.AutoMigrate(&models.RoomStatus{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating room_statuses table: %v", err)
	}

	// Add foreign key constraints for room_statuses
	if err := tx.Exec(`ALTER TABLE room_statuses 
        ADD CONSTRAINT fk_room_statuses_room 
        FOREIGN KEY (room_num) 
        REFERENCES rooms(room_num)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding room foreign key to room_statuses: %v", err)
	}

	if err := tx.Exec(`ALTER TABLE room_statuses 
        ADD CONSTRAINT fk_room_statuses_booking 
        FOREIGN KEY (booking_id) 
        REFERENCES bookings(booking_id)`).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Error adding booking foreign key to room_statuses: %v", err)
	}

	// Create LastRunning table
	if err := tx.AutoMigrate(&models.LastRunning{}); err != nil {
		tx.Rollback()
		log.Fatalf("Error creating last_running table: %v", err)
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
    DECLARE
        is_available BOOLEAN;
    BEGIN
        -- First, check if all the dates in the booking range are available
        SELECT COUNT(*) = 0 INTO is_available
        FROM room_statuses
        WHERE room_num = NEW.room_num
          AND calendar >= NEW.check_in_date::date
          AND calendar <= NEW.check_out_date::date
          AND status != 'Available';
        
        -- If not all dates are available, raise an exception to prevent the booking
        IF NOT is_available THEN
            RAISE EXCEPTION 'Room % is not available for the requested dates', NEW.room_num;
        END IF;

        -- If all dates are available, update the room statuses
        UPDATE room_statuses
        SET status = 'Occupied',
            booking_id = NEW.booking_id
        WHERE room_num = NEW.room_num
          AND calendar >= NEW.check_in_date::date
          AND calendar <= NEW.check_out_date::date;  

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
    AFTER INSERT ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_room_status_for_booking();
    `

	// Create indexes for performance optimization
	createIndexes := []string{
		"CREATE INDEX idx_room_status_room_date ON room_statuses(room_num, calendar)",
		"CREATE INDEX idx_bookings_date_range ON bookings(check_in_date, check_out_date)",
		"CREATE INDEX idx_room_status_status ON room_statuses(status)",
		"CREATE INDEX idx_room_type_id ON rooms(type_id)",
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
		if err := tx.Exec(stmt).Error; err != nil {
			tx.Rollback()
			log.Fatalf("Error executing statement: %v\nStatement: %s", err, stmt)
		}
	}

	// Add constraints for data integrity
	constraints := []string{
		// Ensure check_out_date is after check_in_date
		`ALTER TABLE bookings 
         ADD CONSTRAINT check_dates 
         CHECK (check_out_date > check_in_date)`,

		// Ensure valid status values for room_statuses
		`ALTER TABLE room_statuses 
         ADD CONSTRAINT check_room_status 
         CHECK (status IN ('Available', 'Occupied', 'Reserved', 'Maintenance'))`,

		// Add constraint for payment methods
		`ALTER TABLE receipts 
         ADD CONSTRAINT check_payment_method 
         CHECK (payment_method IN ('Credit', 'Debit', 'Bank Transfer'))`,
	}

	for _, constraint := range constraints {
		if err := tx.Exec(constraint).Error; err != nil {
			log.Printf("Warning while adding constraint: %v", err)
			// Not rolling back for constraint warnings as they might be just duplicates
		}
	}

	// Add optimistic locking for concurrency control
	if err := tx.Exec(`ALTER TABLE bookings ADD COLUMN version INTEGER DEFAULT 1`).Error; err != nil {
		log.Printf("Warning: Could not add version column for optimistic locking: %v", err)
	}

	// Re-enable foreign key checks
	tx.Exec("SET CONSTRAINTS ALL IMMEDIATE")

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("Migration completed successfully!")
	log.Println("Created:")
	log.Println("- 8 tables (room_types, facilities, room_facilities, rooms, bookings, receipts, room_statuses, last_running)")
	log.Println("- 2 triggers (populate_room_status, update_room_status)")
	log.Println("- 5 indexes for query optimization")
	log.Println("- Multiple constraints for data integrity")
	log.Println("- Optimistic locking for concurrency control")
}
