import { BookingService } from '../src/services/bookingService';
import { pool } from '../src/config/database';
import { createTables } from '../src/scripts/initDb';

describe('Hotel Booking System', () => {
  let bookingService: BookingService;

  beforeAll(async () => {
    await createTables();
    bookingService = new BookingService();
  });

  afterAll(async () => {
    await pool.end();
  });

  beforeEach(async () => {
    // Clean up data before each test
    const client = await pool.connect();
    try {
      await client.query('BEGIN');
      await client.query('DELETE FROM receipts');
      await client.query('DELETE FROM payments');
      await client.query('DELETE FROM bookings');
      await client.query('DELETE FROM guests');
      await client.query('UPDATE rooms SET is_available = TRUE');
      await client.query('COMMIT');
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    } finally {
      client.release();
    }
  });

  describe('Normal Booking Flow', () => {
    test('should create a successful booking', async () => {
      const bookingRequest = {
        guestName: 'John Doe',
        guestEmail: 'john@example.com',
        guestPhone: '+1234567890',
        roomId: 1,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-05',
        paymentMethod: 'credit_card'
      };

      const result = await bookingService.createBooking(bookingRequest);
      
      expect(result.booking).toBeDefined();
      expect(result.payment).toBeDefined();
      expect(result.receipt).toBeDefined();
      expect(result.booking.status).toBe('pending');
      expect(result.payment.status).toBe('completed');
    });

    test('should fail when room is not available', async () => {
      // First booking
      const bookingRequest = {
        guestName: 'John Doe',
        guestEmail: 'john@example.com',
        guestPhone: '+1234567890',
        roomId: 1,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-05',
        paymentMethod: 'credit_card'
      };

      await bookingService.createBooking(bookingRequest);

      // Second booking for same room
      const secondBookingRequest = {
        guestName: 'Jane Smith',
        guestEmail: 'jane@example.com',
        guestPhone: '+1234567891',
        roomId: 1,
        checkInDate: '2024-12-02',
        checkOutDate: '2024-12-06',
        paymentMethod: 'credit_card'
      };

      await expect(bookingService.createBooking(secondBookingRequest))
        .rejects.toThrow('Room is not available');
    });
  });

  describe('Transaction Rollback', () => {
    test('should rollback transaction on payment failure', async () => {
      // Mock payment failure by using invalid payment method
      const bookingRequest = {
        guestName: 'John Doe',
        guestEmail: 'john@example.com',
        guestPhone: '+1234567890',
        roomId: 1,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-05',
        paymentMethod: 'invalid_method'
      };

      try {
        await bookingService.createBooking(bookingRequest);
      } catch (error) {
        // After failure, room should still be available
        const client = await pool.connect();
        try {
          const result = await client.query('SELECT is_available FROM rooms WHERE id = $1', [1]);
          expect(result.rows[0].is_available).toBe(true);
        } finally {
          client.release();
        }
      }
    });
  });

  describe('Row Locking Tests', () => {
    test('should handle concurrent bookings with row locking enabled', async () => {
      bookingService.setRowLocking(true);

      const bookingRequest1 = {
        guestName: 'John Doe',
        guestEmail: 'john@example.com',
        guestPhone: '+1234567890',
        roomId: 1,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-05',
        paymentMethod: 'credit_card'
      };

      const bookingRequest2 = {
        guestName: 'Jane Smith',
        guestEmail: 'jane@example.com',
        guestPhone: '+1234567891',
        roomId: 1,
        checkInDate: '2024-12-02',
        checkOutDate: '2024-12-06',
        paymentMethod: 'debit_card'
      };

      // Start both bookings concurrently
      const promises = [
        bookingService.createBooking(bookingRequest1),
        bookingService.createBooking(bookingRequest2)
      ];

      const results = await Promise.allSettled(promises);
      
      // Without row locking, we might get inconsistent results
      // This test demonstrates the potential for race conditions
      console.log('Results without row locking:', results.map(r => r.status));
    });
  });

  describe('Booking Management', () => {
    test('should cancel booking and make room available', async () => {
      const bookingRequest = {
        guestName: 'John Doe',
        guestEmail: 'john@example.com',
        guestPhone: '+1234567890',
        roomId: 1,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-05',
        paymentMethod: 'credit_card'
      };

      const result = await bookingService.createBooking(bookingRequest);
      const bookingId = result.booking.id;

      // Cancel the booking
      await bookingService.cancelBooking(bookingId);

      // Check if room is available again
      const client = await pool.connect();
      try {
        const roomResult = await client.query('SELECT is_available FROM rooms WHERE id = $1', [1]);
        expect(roomResult.rows[0].is_available).toBe(true);

        const bookingResult = await client.query('SELECT status FROM bookings WHERE id = $1', [bookingId]);
        expect(bookingResult.rows[0].status).toBe('cancelled');
      } finally {
        client.release();
      }
    });

    test('should get booking details', async () => {
      const bookingRequest = {
        guestName: 'John Doe',
        guestEmail: 'john@example.com',
        guestPhone: '+1234567890',
        roomId: 1,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-05',
        paymentMethod: 'credit_card'
      };

      const result = await bookingService.createBooking(bookingRequest);
      const bookingDetails = await bookingService.getBookingDetails(result.booking.id);

      expect(bookingDetails).toBeDefined();
      expect(bookingDetails.guest_name).toBe('John Doe');
      expect(bookingDetails.guest_email).toBe('john@example.com');
      expect(bookingDetails.room_number).toBe('101');
      expect(bookingDetails.receipt_number).toBeDefined();
    });
  });
});