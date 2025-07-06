#!/bin/bash

BASE_URL="http://localhost:3000/api"
TOTAL_REQUESTS=50
CONCURRENT_BATCHES=10
ROOM_ID=1

echo "🚀 Stress Testing Hotel Booking API"
echo "==================================="
echo ""

# Check if PostgreSQL container is running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "❌ PostgreSQL container is not running. Start it with: docker-compose up -d postgres"
    exit 1
fi

# Function to make booking requests in batches
run_batch() {
    local batch_num=$1
    local start_id=$((batch_num * 10))
    
    echo "Running batch $batch_num (requests $start_id - $((start_id + 9)))..."
    
    for i in $(seq $start_id $((start_id + 9))); do
        curl -s -X POST "$BASE_URL/bookings" \
            -H "Content-Type: application/json" \
            -d "{
                \"guestName\": \"StressTest Guest $i\",
                \"guestEmail\": \"stress$i@example.com\",
                \"guestPhone\": \"+12345678$(printf "%02d" $i)\",
                \"roomId\": $((ROOM_ID + (i % 5))),
                \"checkInDate\": \"2024-12-0$((1 + (i % 9)))\",
                \"checkOutDate\": \"2024-12-0$((5 + (i % 9)))\",
                \"paymentMethod\": \"credit_card\"
            }" > /dev/null 2>&1 &
    done
    
    wait
}

# Test with row locking enabled
echo "Test 1: Stress test WITH row locking"
echo "------------------------------------"
curl -s -X POST "$BASE_URL/settings/row-locking" \
    -H "Content-Type: application/json" \
    -d '{"enabled": true}' > /dev/null

echo "Running $TOTAL_REQUESTS requests in $CONCURRENT_BATCHES concurrent batches..."
start_time=$(date +%s)

for batch in $(seq 1 $CONCURRENT_BATCHES); do
    run_batch $batch &
done

wait
end_time=$(date +%s)
duration=$((end_time - start_time))

echo "✅ Completed in $duration seconds"
echo ""

# Test with row locking disabled
echo "Test 2: Stress test WITHOUT row locking"
echo "---------------------------------------"
curl -s -X POST "$BASE_URL/settings/row-locking" \
    -H "Content-Type: application/json" \
    -d '{"enabled": false}' > /dev/null

echo "Running $TOTAL_REQUESTS requests in $CONCURRENT_BATCHES concurrent batches..."
start_time=$(date +%s)

for batch in $(seq 1 $CONCURRENT_BATCHES); do
    run_batch $batch &
done

wait
end_time=$(date +%s)
duration=$((end_time - start_time))

echo "✅ Completed in $duration seconds"
echo ""

echo "🔍 Checking final database state..."
docker-compose exec postgres psql -U postgres -d hotel_booking -c "
    SELECT 
        'Bookings' as table_name,
        COUNT(*) as record_count
    FROM bookings
    UNION ALL
    SELECT 
        'Payments' as table_name,
        COUNT(*) as record_count
    FROM payments
    UNION ALL
    SELECT 
        'Receipts' as table_name,
        COUNT(*) as record_count
    FROM receipts
    UNION ALL
    SELECT 
        'Available Rooms' as table_name,
        COUNT(*) as record_count
    FROM rooms
    WHERE is_available = true;
"

echo ""
echo "✅ Stress test completed!"
echo ""
echo "💡 Check the server logs to analyze:"
echo "   - Transaction success/failure rates"
echo "   - Lock contention issues"
echo "   - Database deadlocks"
echo "   - Performance metrics"
echo ""
echo "🐳 Docker logs:"
echo "   - Server logs: Check your terminal where 'npm run dev' is running"
echo "   - PostgreSQL logs: docker-compose logs postgres"