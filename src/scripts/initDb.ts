import { pool } from '../config/database';
import { logger } from '../utils/logger';

const createTables = async () => {
  const client = await pool.connect();
  
  try {
    await client.query('BEGIN');

    // Create guests table
    await client.query(`
      CREATE TABLE IF NOT EXISTS guests (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        phone VARCHAR(20) NOT NULL,
        booking_count INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);

    // Create rooms table
    await client.query(`
      CREATE TABLE IF NOT EXISTS rooms (
        id SERIAL PRIMARY KEY,
        room_number VARCHAR(10) UNIQUE NOT NULL,
        room_type VARCHAR(50) NOT NULL,
        price_per_night DECIMAL(10,2) NOT NULL,
        is_available BOOLEAN DEFAULT TRUE,
        booking_count INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);

    // Create bookings table
    await client.query(`
      CREATE TABLE IF NOT EXISTS bookings (
        id SERIAL PRIMARY KEY,
        guest_id INTEGER REFERENCES guests(id),
        room_id INTEGER REFERENCES rooms(id),
        check_in_date DATE NOT NULL,
        check_out_date DATE NOT NULL,
        total_amount DECIMAL(10,2) NOT NULL,
        status VARCHAR(20) DEFAULT 'pending',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);

    // Create payments table
    await client.query(`
      CREATE TABLE IF NOT EXISTS payments (
        id SERIAL PRIMARY KEY,
        booking_id INTEGER REFERENCES bookings(id),
        amount DECIMAL(10,2) NOT NULL,
        payment_method VARCHAR(50) NOT NULL,
        status VARCHAR(20) DEFAULT 'pending',
        transaction_id VARCHAR(100) UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);

    // Create receipts table
    await client.query(`
      CREATE TABLE IF NOT EXISTS receipts (
        id SERIAL PRIMARY KEY,
        booking_id INTEGER REFERENCES bookings(id),
        payment_id INTEGER REFERENCES payments(id),
        receipt_number VARCHAR(50) UNIQUE NOT NULL,
        total_amount DECIMAL(10,2) NOT NULL,
        generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);

    // Add missing columns if they don't exist (for existing databases)
    await client.query(`
      ALTER TABLE guests 
      ADD COLUMN IF NOT EXISTS booking_count INTEGER DEFAULT 0
    `);

    await client.query(`
      ALTER TABLE rooms 
      ADD COLUMN IF NOT EXISTS booking_count INTEGER DEFAULT 0
    `);

    // Insert sample rooms
    await client.query(`
      INSERT INTO rooms (room_number, room_type, price_per_night) VALUES
      ('101', 'Standard', 100.00),
      ('102', 'Standard', 100.00),
      ('201', 'Deluxe', 150.00),
      ('202', 'Deluxe', 150.00),
      ('301', 'Suite', 250.00)
      ON CONFLICT (room_number) DO NOTHING
    `);

    // Create indexes for better performance and deadlock testing
    await client.query(`
      CREATE INDEX IF NOT EXISTS idx_guests_email ON guests(email)
    `);

    await client.query(`
      CREATE INDEX IF NOT EXISTS idx_rooms_availability ON rooms(is_available)
    `);

    await client.query(`
      CREATE INDEX IF NOT EXISTS idx_bookings_guest_id ON bookings(guest_id)
    `);

    await client.query(`
      CREATE INDEX IF NOT EXISTS idx_bookings_room_id ON bookings(room_id)
    `);

    await client.query(`
      CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status)
    `);

    await client.query('COMMIT');
    logger.info('Database initialized successfully');
    
  } catch (error) {
    await client.query('ROLLBACK');
    logger.error('Failed to initialize database', { error: error instanceof Error ? error.message : String(error) });
    throw error;
  } finally {
    client.release();
  }
};

// Additional function to populate test data for deadlock testing
const populateTestData = async () => {
  const client = await pool.connect();
  
  try {
    await client.query('BEGIN');

    // Insert test guests
    await client.query(`
      INSERT INTO guests (name, email, phone) VALUES
      ('John Doe', 'john.doe@example.com', '555-0001'),
      ('Jane Smith', 'jane.smith@example.com', '555-0002'),
      ('Bob Johnson', 'bob.johnson@example.com', '555-0003'),
      ('Alice Brown', 'alice.brown@example.com', '555-0004'),
      ('Charlie Wilson', 'charlie.wilson@example.com', '555-0005')
      ON CONFLICT (email) DO NOTHING
    `);

    // Add more test rooms
    await client.query(`
      INSERT INTO rooms (room_number, room_type, price_per_night) VALUES
      ('103', 'Standard', 100.00),
      ('104', 'Standard', 100.00),
      ('203', 'Deluxe', 150.00),
      ('204', 'Deluxe', 150.00),
      ('302', 'Suite', 250.00),
      ('303', 'Suite', 250.00)
      ON CONFLICT (room_number) DO NOTHING
    `);

    await client.query('COMMIT');
    logger.info('Test data populated successfully');
    
  } catch (error) {
    await client.query('ROLLBACK');
    logger.error('Failed to populate test data', { error: error instanceof Error ? error.message : String(error) });
    throw error;
  } finally {
    client.release();
  }
};

// Function to reset counters (useful for testing)
const resetBookingCounters = async () => {
  const client = await pool.connect();
  
  try {
    await client.query('BEGIN');

    await client.query('UPDATE guests SET booking_count = 0');
    await client.query('UPDATE rooms SET booking_count = 0');

    await client.query('COMMIT');
    logger.info('Booking counters reset successfully');
    
  } catch (error) {
    await client.query('ROLLBACK');
    logger.error('Failed to reset booking counters', { error: error instanceof Error ? error.message : String(error) });
    throw error;
  } finally {
    client.release();
  }
};

// Run if called directly
if (require.main === module) {
  createTables()
    .then(() => populateTestData())
    .then(() => {
      logger.info('Database setup complete');
      process.exit(0);
    })
    .catch((error) => {
      logger.error('Database setup failed', { error });
      process.exit(1);
    });
}

export { createTables, populateTestData, resetBookingCounters };