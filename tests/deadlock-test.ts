// deadlock-test.ts
import { BookingService } from '../src/services/bookingService';
import { Booking, Payment, Receipt } from '../src/types/index';

interface BookingResponse {
  booking: Booking;
  payment: Payment;
  receipt: Receipt;
}

const bookingService = new BookingService();

// Test 1: Concurrent bookings that can cause deadlocks
async function testConcurrentBookings() {
  console.log('Testing concurrent bookings...');
  
  // Disable row locking to enable deadlock scenarios
  bookingService.setRowLocking(false);
  
  const booking1 = {
    guestName: 'John Doe',
    guestEmail: 'john@example.com',
    guestPhone: '123-456-7890',
    roomId: 1,
    checkInDate: '2024-12-01',
    checkOutDate: '2024-12-03',
    paymentMethod: 'credit_card'
  };
  
  const booking2 = {
    guestName: 'Jane Smith',
    guestEmail: 'jane@example.com',
    guestPhone: '987-654-3210',
    roomId: 2,
    checkInDate: '2024-12-02',
    checkOutDate: '2024-12-04',
    paymentMethod: 'debit_card'
  };
  
  try {
    // Run multiple concurrent bookings
    const promises: Promise<BookingResponse>[] = [];
    for (let i = 0; i < 10; i++) {
      promises.push(bookingService.createBooking({
        ...booking1,
        guestEmail: `john${i}@example.com`,
        roomId: (i % 2) + 1 // Alternate between room 1 and 2
      }));
    }
    
    const results = await Promise.allSettled(promises);
    
    const successes = results.filter(r => r.status === 'fulfilled').length;
    const failures = results.filter(r => r.status === 'rejected').length;
    
    console.log(`Concurrent bookings: ${successes} successful, ${failures} failed`);
    
    // Print failure reasons
    results.forEach((result, index) => {
      if (result.status === 'rejected') {
        console.log(`Booking ${index + 1} failed:`, result.reason.message);
      }
    });
    
  } catch (error) {
    console.error('Error in concurrent booking test:', error);
  }
}

// Test 2: Concurrent booking and cancellation
async function testConcurrentBookingAndCancellation() {
  console.log('\nTesting concurrent booking and cancellation...');
  
  bookingService.setRowLocking(false);
  
  try {
    // Create a booking first
    const booking = await bookingService.createBooking({
      guestName: 'Test User',
      guestEmail: 'test@example.com',
      guestPhone: '555-0123',
      roomId: 1,
      checkInDate: '2024-12-10',
      checkOutDate: '2024-12-12',
      paymentMethod: 'credit_card'
    });
    
    console.log('Initial booking created:', booking.booking.id);
    
    // Now run concurrent operations
    const promises = [
      // Multiple cancellations of the same booking
      bookingService.cancelBooking(booking.booking.id),
      bookingService.cancelBooking(booking.booking.id),
      // New bookings for same room
      bookingService.createBooking({
        guestName: 'Another User',
        guestEmail: 'another@example.com',
        guestPhone: '555-0124',
        roomId: 1,
        checkInDate: '2024-12-11',
        checkOutDate: '2024-12-13',
        paymentMethod: 'credit_card'
      }),
      bookingService.createBooking({
        guestName: 'Third User',
        guestEmail: 'third@example.com',
        guestPhone: '555-0125',
        roomId: 1,
        checkInDate: '2024-12-12',
        checkOutDate: '2024-12-14',
        paymentMethod: 'credit_card'
      })
    ];
    
    const results = await Promise.allSettled(promises);
    
    const successes = results.filter(r => r.status === 'fulfilled').length;
    const failures = results.filter(r => r.status === 'rejected').length;
    
    console.log(`Concurrent operations: ${successes} successful, ${failures} failed`);
    
    // Print failure reasons
    results.forEach((result, index) => {
      if (result.status === 'rejected') {
        console.log(`Operation ${index + 1} failed:`, result.reason.message);
      }
    });
    
  } catch (error) {
    console.error('Error in concurrent booking/cancellation test:', error);
  }
}

// Test 3: Bulk price updates causing deadlocks
async function testBulkPriceUpdates() {
  console.log('\nTesting bulk price updates...');
  
  bookingService.setRowLocking(false);
  
  try {
    const roomIds = [1, 2, 3, 4, 5];
    
    // Run multiple concurrent bulk updates
    const promises = [
      bookingService.bulkUpdateRoomPricing(roomIds, 10),
      bookingService.bulkUpdateRoomPricing([5, 4, 3, 2, 1], -5), // Reverse order
      bookingService.bulkUpdateRoomPricing([2, 4, 1, 5, 3], 15), // Random order
      bookingService.bulkUpdateRoomPricing(roomIds, -10)
    ];
    
    const results = await Promise.allSettled(promises);
    
    const successes = results.filter(r => r.status === 'fulfilled').length;
    const failures = results.filter(r => r.status === 'rejected').length;
    
    console.log(`Bulk updates: ${successes} successful, ${failures} failed`);
    
    // Print failure reasons
    results.forEach((result, index) => {
      if (result.status === 'rejected') {
        console.log(`Bulk update ${index + 1} failed:`, result.reason.message);
      }
    });
    
  } catch (error) {
    console.error('Error in bulk price update test:', error);
  }
}

// Test 4: High concurrency stress test
async function stressTest() {
  console.log('\nRunning stress test...');
  
  bookingService.setRowLocking(false);
  
  const promises: Promise<BookingResponse>[] = [];
  const roomIds = [1, 2, 3];
  
  // Create many concurrent booking requests
  for (let i = 0; i < 50; i++) {
    promises.push(bookingService.createBooking({
      guestName: `User ${i}`,
      guestEmail: `user${i}@example.com`,
      guestPhone: `555-${String(i).padStart(4, '0')}`,
      roomId: roomIds[i % roomIds.length],
      checkInDate: '2024-12-20',
      checkOutDate: '2024-12-22',
      paymentMethod: 'credit_card'
    }));
  }
  
  try {
    const results = await Promise.allSettled(promises);
    
    const successes = results.filter(r => r.status === 'fulfilled').length;
    const failures = results.filter(r => r.status === 'rejected').length;
    const deadlocks = results.filter(r => 
      r.status === 'rejected' && 
      r.reason.message.includes('deadlock')
    ).length;
    
    console.log(`Stress test results: ${successes} successful, ${failures} failed, ${deadlocks} deadlocks`);
    
  } catch (error) {
    console.error('Error in stress test:', error);
  }
}

// Run all tests
async function runAllTests() {
  console.log('=== DEADLOCK TESTING WITH ROW LOCKING DISABLED ===\n');
  
  await testConcurrentBookings();
  await testConcurrentBookingAndCancellation();
  await testBulkPriceUpdates();
  await stressTest();
  
  console.log('\n=== TESTING COMPLETE ===');
  console.log('To avoid deadlocks, enable row locking: bookingService.setRowLocking(true)');
}

// Export for use in your application
export { runAllTests };

// Run tests if this file is executed directly
if (require.main === module) {
  runAllTests().catch(console.error);
}