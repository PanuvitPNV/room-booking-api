import { PoolClient } from 'pg';
import { getClient } from '../config/database';
import { logger } from '../utils/logger';
import { Booking, Guest, Room, Payment, Receipt } from '../types';

interface BookingRequest {
  guestName: string;
  guestEmail: string;
  guestPhone: string;
  roomId: number;
  checkInDate: string;
  checkOutDate: string;
  paymentMethod: string;
}

interface BookingResponse {
  booking: Booking;
  payment: Payment;
  receipt: Receipt;
}

export class BookingService {
  private enableRowLocking: boolean = true;

  setRowLocking(enabled: boolean) {
    this.enableRowLocking = enabled;
    logger.info(`Row locking ${enabled ? 'enabled' : 'disabled'}`);
  }

  async createBooking(request: BookingRequest): Promise<BookingResponse> {
    const client = await getClient();
    
    try {
      await client.query('BEGIN');
      logger.info('Transaction started', { bookingRequest: request });

      // Step 1: Create or get guest
      const guest = await this.createOrGetGuest(client, {
        name: request.guestName,
        email: request.guestEmail,
        phone: request.guestPhone
      });

      // Step 2: Check room availability with optional locking
      const room = await this.checkRoomAvailability(client, request.roomId);
      
      // Step 3: Calculate total amount
      const checkIn = new Date(request.checkInDate);
      const checkOut = new Date(request.checkOutDate);
      const nights = Math.ceil((checkOut.getTime() - checkIn.getTime()) / (1000 * 60 * 60 * 24));
      const totalAmount = room.price_per_night * nights;

      // Step 4: Create booking
      const booking = await this.createBookingRecord(client, {
        guestId: guest.id,
        roomId: request.roomId,
        checkInDate: request.checkInDate,
        checkOutDate: request.checkOutDate,
        totalAmount
      });

      // Step 5: Update room availability
      await this.updateRoomAvailability(client, request.roomId, false);

      // Step 6: Process payment
      const payment = await this.processPayment(client, {
        bookingId: booking.id,
        amount: totalAmount,
        paymentMethod: request.paymentMethod
      });

      // Step 7: Generate receipt
      const receipt = await this.generateReceipt(client, booking.id, payment.id, totalAmount);

      // Step 8: Update booking statistics (NEW - potential deadlock scenario)
      await this.updateBookingStatistics(client, request.roomId, guest.id);

      await client.query('COMMIT');
      logger.info('Transaction committed successfully', { bookingId: booking.id });

      return { booking, payment, receipt };

    } catch (error) {
      await client.query('ROLLBACK');
      if (error instanceof Error) {
        logger.error('Transaction rolled back', { error: error.message });
      } else {
        logger.error('Transaction rolled back', { error: String(error) });
      }
      throw error;
    } finally {
      client.release();
    }
  }

  private async createOrGetGuest(client: PoolClient, guestData: Partial<Guest>): Promise<Guest> {
    // Check if guest exists
    const existingGuest = await client.query(
      'SELECT * FROM guests WHERE email = $1',
      [guestData.email]
    );

    if (existingGuest.rows.length > 0) {
      return existingGuest.rows[0];
    }

    // Create new guest
    const result = await client.query(
      `INSERT INTO guests (name, email, phone) 
       VALUES ($1, $2, $3) 
       RETURNING *`,
      [guestData.name, guestData.email, guestData.phone]
    );

    logger.info('New guest created', { guestId: result.rows[0].id });
    return result.rows[0];
  }

  private async checkRoomAvailability(client: PoolClient, roomId: number): Promise<Room> {
    const lockClause = this.enableRowLocking ? 'FOR UPDATE' : '';
    
    const result = await client.query(
      `SELECT * FROM rooms WHERE id = $1 ${lockClause}`,
      [roomId]
    );

    if (result.rows.length === 0) {
      throw new Error('Room not found');
    }

    const room = result.rows[0];
    if (!room.is_available) {
      throw new Error('Room is not available');
    }

    logger.info('Room availability checked', { 
      roomId, 
      available: room.is_available,
      lockingEnabled: this.enableRowLocking 
    });

    return room;
  }

  private async createBookingRecord(client: PoolClient, data: {
    guestId: number;
    roomId: number;
    checkInDate: string;
    checkOutDate: string;
    totalAmount: number;
  }): Promise<Booking> {
    const result = await client.query(
      `INSERT INTO bookings (guest_id, room_id, check_in_date, check_out_date, total_amount, status) 
       VALUES ($1, $2, $3, $4, $5, 'pending') 
       RETURNING *`,
      [data.guestId, data.roomId, data.checkInDate, data.checkOutDate, data.totalAmount]
    );

    logger.info('Booking record created', { bookingId: result.rows[0].id });
    return result.rows[0];
  }

  private async updateRoomAvailability(client: PoolClient, roomId: number, isAvailable: boolean): Promise<void> {
    await client.query(
      'UPDATE rooms SET is_available = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2',
      [isAvailable, roomId]
    );

    logger.info('Room availability updated', { roomId, isAvailable });
  }

  private async processPayment(client: PoolClient, data: {
    bookingId: number;
    amount: number;
    paymentMethod: string;
  }): Promise<Payment> {
    const transactionId = `TXN_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    // Simulate payment processing delay
    await new Promise(resolve => setTimeout(resolve, 100));

    const result = await client.query(
      `INSERT INTO payments (booking_id, amount, payment_method, status, transaction_id) 
       VALUES ($1, $2, $3, 'completed', $4) 
       RETURNING *`,
      [data.bookingId, data.amount, data.paymentMethod, transactionId]
    );

    logger.info('Payment processed', { paymentId: result.rows[0].id, transactionId });
    return result.rows[0];
  }

  private async generateReceipt(client: PoolClient, bookingId: number, paymentId: number, totalAmount: number): Promise<Receipt> {
    const receiptNumber = `RCP_${Date.now()}_${Math.random().toString(36).substr(2, 6)}`;
    
    const result = await client.query(
      `INSERT INTO receipts (booking_id, payment_id, receipt_number, total_amount) 
       VALUES ($1, $2, $3, $4) 
       RETURNING *`,
      [bookingId, paymentId, receiptNumber, totalAmount]
    );

    logger.info('Receipt generated', { receiptId: result.rows[0].id, receiptNumber });
    return result.rows[0];
  }

  // NEW METHOD: Creates deadlock scenario when row locking is disabled
  private async updateBookingStatistics(client: PoolClient, roomId: number, guestId: number): Promise<void> {
    // First, update guest statistics (increment booking count)
    const lockClause = this.enableRowLocking ? 'FOR UPDATE' : '';
    
    // Access guest first, then room (order matters for deadlock)
    await client.query(
      `UPDATE guests SET booking_count = COALESCE(booking_count, 0) + 1, updated_at = CURRENT_TIMESTAMP 
       WHERE id = (SELECT id FROM guests WHERE id = $1 ${lockClause})`,
      [guestId]
    );

    // Add artificial delay to increase chance of deadlock
    await new Promise(resolve => setTimeout(resolve, 50));

    // Then update room statistics (increment booking count)
    await client.query(
      `UPDATE rooms SET booking_count = COALESCE(booking_count, 0) + 1, updated_at = CURRENT_TIMESTAMP 
       WHERE id = (SELECT id FROM rooms WHERE id = $1 ${lockClause})`,
      [roomId]
    );

    logger.info('Booking statistics updated', { roomId, guestId, lockingEnabled: this.enableRowLocking });
  }

  async cancelBooking(bookingId: number): Promise<void> {
    const client = await getClient();
    
    try {
      await client.query('BEGIN');
      
      // Get booking details with potential deadlock scenario
      const bookingResult = await client.query(
        'SELECT * FROM bookings WHERE id = $1',
        [bookingId]
      );

      if (bookingResult.rows.length === 0) {
        throw new Error('Booking not found');
      }

      const booking = bookingResult.rows[0];
      
      // Update booking status
      await client.query(
        'UPDATE bookings SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2',
        ['cancelled', bookingId]
      );

      // Make room available again
      await this.updateRoomAvailability(client, booking.room_id, true);

      // NEW: Revert statistics (potential deadlock scenario)
      await this.revertBookingStatistics(client, booking.room_id, booking.guest_id);

      await client.query('COMMIT');
      logger.info('Booking cancelled successfully', { bookingId });

    } catch (error) {
      await client.query('ROLLBACK');
      if (error instanceof Error) {
        logger.error('Failed to cancel booking', { bookingId, error: error.message });
      } else {
        logger.error('Failed to cancel booking', { bookingId, error: String(error) });
      }
      throw error;
    } finally {
      client.release();
    }
  }

  // NEW METHOD: Creates deadlock scenario when row locking is disabled
  private async revertBookingStatistics(client: PoolClient, roomId: number, guestId: number): Promise<void> {
    const lockClause = this.enableRowLocking ? 'FOR UPDATE' : '';
    
    // Access room first, then guest (opposite order from updateBookingStatistics)
    await client.query(
      `UPDATE rooms SET booking_count = GREATEST(COALESCE(booking_count, 0) - 1, 0), updated_at = CURRENT_TIMESTAMP 
       WHERE id = (SELECT id FROM rooms WHERE id = $1 ${lockClause})`,
      [roomId]
    );

    // Add artificial delay to increase chance of deadlock
    await new Promise(resolve => setTimeout(resolve, 50));

    // Then update guest statistics
    await client.query(
      `UPDATE guests SET booking_count = GREATEST(COALESCE(booking_count, 0) - 1, 0), updated_at = CURRENT_TIMESTAMP 
       WHERE id = (SELECT id FROM guests WHERE id = $1 ${lockClause})`,
      [guestId]
    );

    logger.info('Booking statistics reverted', { roomId, guestId, lockingEnabled: this.enableRowLocking });
  }

  async getBookingDetails(bookingId: number) {
    const client = await getClient();
    
    try {
      const result = await client.query(`
        SELECT 
          b.*,
          g.name as guest_name,
          g.email as guest_email,
          g.phone as guest_phone,
          r.room_number,
          r.room_type,
          r.price_per_night,
          p.transaction_id,
          p.payment_method,
          p.status as payment_status,
          rec.receipt_number
        FROM bookings b
        JOIN guests g ON b.guest_id = g.id
        JOIN rooms r ON b.room_id = r.id
        LEFT JOIN payments p ON b.id = p.booking_id
        LEFT JOIN receipts rec ON b.id = rec.booking_id
        WHERE b.id = $1
      `, [bookingId]);

      return result.rows[0] || null;
    } finally {
      client.release();
    }
  }

  // NEW METHOD: Bulk operation that can cause deadlocks
  async bulkUpdateRoomPricing(roomIds: number[], priceAdjustment: number): Promise<void> {
    const client = await getClient();
    
    try {
      await client.query('BEGIN');
      
      // Process rooms in different orders to create deadlock potential
      const shuffledRoomIds = this.enableRowLocking ? roomIds : this.shuffleArray([...roomIds]);
      
      for (const roomId of shuffledRoomIds) {
        const lockClause = this.enableRowLocking ? 'FOR UPDATE' : '';
        
        // Get current room data
        const roomResult = await client.query(
          `SELECT price_per_night FROM rooms WHERE id = $1 ${lockClause}`,
          [roomId]
        );
        
        if (roomResult.rows.length > 0) {
          const currentPrice = roomResult.rows[0].price_per_night;
          const newPrice = currentPrice + priceAdjustment;
          
          // Add delay to increase deadlock chance
          await new Promise(resolve => setTimeout(resolve, 25));
          
          await client.query(
            'UPDATE rooms SET price_per_night = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2',
            [newPrice, roomId]
          );
        }
      }
      
      await client.query('COMMIT');
      logger.info('Bulk room pricing updated', { roomIds: roomIds.length, priceAdjustment });
      
    } catch (error) {
      await client.query('ROLLBACK');
      if (error instanceof Error) {
        logger.error('Failed to update room pricing', { error: error.message });
      } else {
        logger.error('Failed to update room pricing', { error: String(error) });
      }
      throw error;
    } finally {
      client.release();
    }
  }

  // Helper method to shuffle array (creates non-deterministic access order)
  private shuffleArray<T>(array: T[]): T[] {
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [array[i], array[j]] = [array[j], array[i]];
    }
    return array;
  }
}