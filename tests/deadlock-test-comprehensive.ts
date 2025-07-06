// deadlock-test-comprehensive.ts
import { BookingService } from '../src/services/bookingService';
import { logger } from '../src/utils/logger';
import { resetBookingCounters } from '../src/scripts/initDb';
import { Booking, Payment, Receipt } from '../src/types/index';

const bookingService = new BookingService();

// Helper function to create random delay
const randomDelay = (min: number, max: number) => {
  return new Promise(resolve => 
    setTimeout(resolve, Math.floor(Math.random() * (max - min + 1)) + min)
  );
};

// Test 1: Concurrent bookings for same rooms (high deadlock probability)
async function testConcurrentSameRoomBookings() {
  console.log('\n=== Test 1: Concurrent Same Room Bookings ===');
  
  bookingService.setRowLocking(false);
  
  const promises: Promise<any>[] = [];
  
  // Create 20 concurrent requests for the same 2 rooms
  for (let i = 0; i < 20; i++) {
    const roomId = (i % 2) + 1; // Alternate between room 1 and 2
    
    promises.push(
      bookingService.createBooking({
        guestName: `Guest ${i}`,
        guestEmail: `guest${i}@test.com`,
        guestPhone: `555-${String(i).padStart(4, '0')}`,
        roomId: roomId,
        checkInDate: '2024-12-01',
        checkOutDate: '2024-12-03',
        paymentMethod: 'credit_card'
      })
    );
  }
  
  try {
    const results = await Promise.allSettled(promises);
    
    const fulfilled = results.filter(r => r.status === 'fulfilled');
    const rejected = results.filter(r => r.status === 'rejected');
    const deadlocks = rejected.filter(r => 
      r.reason?.message?.toLowerCase().includes('deadlock')
    );
    
    console.log(`Results: ${fulfilled.length} successful, ${rejected.length} failed`);
    console.log(`Deadlocks detected: ${deadlocks.length}`);
    
    if (deadlocks.length > 0) {
      console.log('✅ DEADLOCK SUCCESSFULLY REPRODUCED!');
      deadlocks.slice(0, 3).forEach((result, index) => {
        console.log(`  Deadlock ${index + 1}:`, result.reason.message);
      });
    } else {
      console.log('❌ No deadlocks detected in this test');
    }
    
    return { fulfilled: fulfilled.length, rejected: rejected.length, deadlocks: deadlocks.length };
    
  } catch (error) {
    console.error('Test 1 error:', error);
    return { fulfilled: 0, rejected: 0, deadlocks: 0 };
  }
}

// Test 2: Mixed operations (booking + cancellation)
async function testMixedOperations() {
  console.log('\n=== Test 2: Mixed Operations (Booking + Cancellation) ===');
  
  bookingService.setRowLocking(false);
  
  try {
    // First, create some bookings to cancel
    const initialBookings: number[] = [];
    for (let i = 0; i < 5; i++) {
      const booking = await bookingService.createBooking({
        guestName: `Initial Guest ${i}`,
        guestEmail: `initial${i}@test.com`,
        guestPhone: `555-100${i}`,
        roomId: (i % 3) + 1,
        checkInDate: '2024-12-10',
        checkOutDate: '2024-12-12',
        paymentMethod: 'credit_card'
      });
      initialBookings.push(booking.booking.id);
    }
    
    console.log(`Created ${initialBookings.length} initial bookings`);
    
    // Now run mixed operations concurrently
    const promises: Promise<any>[] = [];
    
    // Add cancellations
    initialBookings.forEach(bookingId => {
      promises.push(bookingService.cancelBooking(bookingId));
    });
    
    // Add new bookings for same rooms
    for (let i = 0; i < 15; i++) {
      promises.push(
        bookingService.createBooking({
          guestName: `New Guest ${i}`,
          guestEmail: `new${i}@test.com`,
          guestPhone: `555-200${i}`,
          roomId: (i % 3) + 1,
          checkInDate: '2024-12-15',
          checkOutDate: '2024-12-17',
          paymentMethod: 'credit_card'
        })
      );
    }
    
    // Add bulk price updates
    promises.push(bookingService.bulkUpdateRoomPricing([1, 2, 3], 25));
    promises.push(bookingService.bulkUpdateRoomPricing([3, 2, 1], -10));
    
    const results = await Promise.allSettled(promises);
    
    const fulfilled = results.filter(r => r.status === 'fulfilled');
    const rejected = results.filter(r => r.status === 'rejected');
    const deadlocks = rejected.filter(r => 
      r.reason?.message?.toLowerCase().includes('deadlock')
    );
    
    console.log(`Results: ${fulfilled.length} successful, ${rejected.length} failed`);
    console.log(`Deadlocks detected: ${deadlocks.length}`);
    
    if (deadlocks.length > 0) {
      console.log('✅ DEADLOCK SUCCESSFULLY REPRODUCED!');
      deadlocks.slice(0, 3).forEach((result, index) => {
        console.log(`  Deadlock ${index + 1}:`, result.reason.message);
      });
    } else {
      console.log('❌ No deadlocks detected in this test');
    }
    
    return { fulfilled: fulfilled.length, rejected: rejected.length, deadlocks: deadlocks.length };
    
  } catch (error) {
    console.error('Test 2 error:', error);
    return { fulfilled: 0, rejected: 0, deadlocks: 0 };
  }
}

// Test 3: High-frequency operations
async function testHighFrequencyOperations() {
  console.log('\n=== Test 3: High-Frequency Operations ===');
  
  bookingService.setRowLocking(false);
  
  const promises: Promise<any>[] = [];
  
  // Create rapid-fire requests
  for (let batch = 0; batch < 5; batch++) {
    for (let i = 0; i < 10; i++) {
      const requestIndex = batch * 10 + i;
      
      promises.push(
        (async () => {
          await randomDelay(0, 20); // Random small delay
          return bookingService.createBooking({
            guestName: `Rapid Guest ${requestIndex}`,
            guestEmail: `rapid${requestIndex}@test.com`,
            guestPhone: `555-300${requestIndex}`,
            roomId: (requestIndex % 4) + 1,
            checkInDate: '2024-12-20',
            checkOutDate: '2024-12-22',
            paymentMethod: 'credit_card'
          });
        })()
      );
    }
  }
  
  try {
    const results = await Promise.allSettled(promises);
    
    const fulfilled = results.filter(r => r.status === 'fulfilled');
    const rejected = results.filter(r => r.status === 'rejected');
    const deadlocks = rejected.filter(r => 
      r.reason?.message?.toLowerCase().includes('deadlock')
    );
    
    console.log(`Results: ${fulfilled.length} successful, ${rejected.length} failed`);
    console.log(`Deadlocks detected: ${deadlocks.length}`);
    
    if (deadlocks.length > 0) {
      console.log('✅ DEADLOCK SUCCESSFULLY REPRODUCED!');
      deadlocks.slice(0, 3).forEach((result, index) => {
        console.log(`  Deadlock ${index + 1}:`, result.reason.message);
      });
    } else {
      console.log('❌ No deadlocks detected in this test');
    }
    
    return { fulfilled: fulfilled.length, rejected: rejected.length, deadlocks: deadlocks.length };
    
  } catch (error) {
    console.error('Test 3 error:', error);
    return { fulfilled: 0, rejected: 0, deadlocks: 0 };
  }
}

// Test 4: Bulk operations causing deadlocks
async function testBulkOperationDeadlocks() {
  console.log('\n=== Test 4: Bulk Operations Deadlocks ===');
  
  bookingService.setRowLocking(false);
  
  const promises: Promise<any>[] = [];
  
  // Multiple bulk operations with different room orders
  const roomSets = [
    [1, 2, 3, 4, 5],
    [5, 4, 3, 2, 1],
    [2, 4, 1, 5, 3],
    [3, 1, 5, 2, 4],
    [4, 5, 2, 1, 3]
  ];
  
  roomSets.forEach((rooms, index) => {
    promises.push(bookingService.bulkUpdateRoomPricing(rooms, (index + 1) * 5));
  });
  
  // Add concurrent bookings during bulk operations
  for (let i = 0; i < 10; i++) {
    promises.push(
      bookingService.createBooking({
        guestName: `Bulk Test Guest ${i}`,
        guestEmail: `bulk${i}@test.com`,
        guestPhone: `555-400${i}`,
        roomId: (i % 5) + 1,
        checkInDate: '2024-12-25',
        checkOutDate: '2024-12-27',
        paymentMethod: 'credit_card'
      })
    );
  }
  
  try {
    const results = await Promise.allSettled(promises);
    
    const fulfilled = results.filter(r => r.status === 'fulfilled');
    const rejected = results.filter(r => r.status === 'rejected');
    const deadlocks = rejected.filter(r => 
      r.reason?.message?.toLowerCase().includes('deadlock')
    );
    
    console.log(`Results: ${fulfilled.length} successful, ${rejected.length} failed`);
    console.log(`Deadlocks detected: ${deadlocks.length}`);
    
    if (deadlocks.length > 0) {
      console.log('✅ DEADLOCK SUCCESSFULLY REPRODUCED!');
      deadlocks.slice(0, 3).forEach((result, index) => {
        console.log(`  Deadlock ${index + 1}:`, result.reason.message);
      });
    } else {
      console.log('❌ No deadlocks detected in this test');
    }
    
    return { fulfilled: fulfilled.length, rejected: rejected.length, deadlocks: deadlocks.length };
    
  } catch (error) {
    console.error('Test 4 error:', error);
    return { fulfilled: 0, rejected: 0, deadlocks: 0 };
  }
}

// Main test runner
async function runDeadlockTests() {
  console.log('🚀 Starting Comprehensive Deadlock Tests');
  console.log('⚠️  Row locking is DISABLED - deadlocks should occur');
  
  try {
    // Reset counters before testing
    await resetBookingCounters();
    
    const test1Results = await testConcurrentSameRoomBookings();
    const test2Results = await testMixedOperations();
    const test3Results = await testHighFrequencyOperations();
    const test4Results = await testBulkOperationDeadlocks();
    
    const totalDeadlocks = test1Results.deadlocks + test2Results.deadlocks + 
                          test3Results.deadlocks + test4Results.deadlocks;
    
    console.log('\n=== SUMMARY ===');
    console.log(`Total deadlocks across all tests: ${totalDeadlocks}`);
    
    if (totalDeadlocks > 0) {
      console.log('✅ SUCCESS: Deadlocks were reproduced!');
      console.log('💡 To prevent deadlocks, enable row locking: bookingService.setRowLocking(true)');
    } else {
      console.log('❌ No deadlocks detected. Try running the tests multiple times.');
      console.log('💡 You may need to increase concurrency or reduce delays.');
    }
    
    // Test with row locking enabled for comparison
    console.log('\n=== Testing with Row Locking Enabled ===');
    bookingService.setRowLocking(true);
    
    const safeTest = await testConcurrentSameRoomBookings();
    console.log(`With row locking: ${safeTest.deadlocks} deadlocks (should be 0)`);
    
  } catch (error) {
    console.error('Error in deadlock tests:', error);
  }
}

// Export for use in other modules
export { 
  runDeadlockTests,
  testConcurrentSameRoomBookings,
  testMixedOperations,
  testHighFrequencyOperations,
  testBulkOperationDeadlocks
};

// Run tests if this file is executed directly
if (require.main === module) {
  runDeadlockTests().catch(console.error);
}