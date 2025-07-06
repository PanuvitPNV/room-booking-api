#!/bin/bash

BASE_URL="http://localhost:3000/api"
CONCURRENT_REQUESTS=5
ROOM_ID=1

echo "🧪 Hotel Booking API Load Tests"
echo "================================"
echo ""

# Function to make a booking request
make_booking() {
    local guest_suffix=$1
    local delay=$2
    
    if [ ! -z "$delay" ]; then
        sleep $delay
    fi
    
    curl -s -X POST "$BASE_URL/bookings" \
        -H "Content-Type: application/json" \
        -d "{
            \"guestName\": \"Guest $guest_suffix\",
            \"guestEmail\": \"guest$guest_suffix@example.com\",
            \"guestPhone\": \"+123456789$guest_suffix\",
            \"roomId\": $ROOM_ID,
            \"checkInDate\": \"2024-12-01\",
            \"checkOutDate\": \"2024-12-05\",
            \"paymentMethod\": \"credit_card\"
        }" | jq -r '.success // false'
}

# Function to cancel a booking
cancel_booking() {
    local booking_id=$1
    curl -s -X DELETE "$BASE_URL/bookings/$booking_id" | jq -r '.success // false'
}

# Function to set row locking
set_row_locking() {
    local enabled=$1
    curl -s -X POST "$BASE_URL/settings/row-locking" \
        -H "Content-Type: application/json" \
        -d "{\"enabled\": $enabled}" | jq -r '.success // false'
}

# Test 1: Concurrent bookings with row locking enabled
echo "Test 1: Concurrent bookings WITH row locking"
echo "--------------------------------------------"
set_row_locking true
echo "Row locking enabled"

# Cancel any existing bookings to free up rooms
curl -s "$BASE_URL/bookings/1" > /dev/null 2>&1 && cancel_booking 1

echo "Making $CONCURRENT_REQUESTS concurrent booking requests..."
for i in $(seq 1 $CONCURRENT_REQUESTS); do
    make_booking $i &
done

wait
echo ""

# Test 2: Concurrent bookings with row locking disabled
echo "Test 2: Concurrent bookings WITHOUT row locking"
echo "-----------------------------------------------"
set_row_locking false
echo "Row locking disabled"

# Cancel any existing bookings to free up rooms
curl -s "$BASE_URL/bookings/1" > /dev/null 2>&1 && cancel_booking 1

echo "Making $CONCURRENT_REQUESTS concurrent booking requests..."
for i in $(seq 6 $((5 + CONCURRENT_REQUESTS))); do
    make_booking $i &
done

wait
echo ""

# Test 3: Deadlock simulation
echo "Test 3: Deadlock simulation"
echo "---------------------------"
set_row_locking true
echo "Row locking enabled"

echo "Simulating potential deadlock scenario..."
echo "Making overlapping bookings with different rooms..."

# Try to book multiple rooms simultaneously
curl -s -X POST "$BASE_URL/bookings" \
    -H "Content-Type: application/json" \
    -d "{
        \"guestName\": \"Deadlock Test 1\",
        \"guestEmail\": \"deadlock1@example.com\",
        \"guestPhone\": \"+1234567890\",
        \"roomId\": 1,
        \"checkInDate\": \"2024-12-01\",
        \"checkOutDate\": \"2024-12-05\",
        \"paymentMethod\": \"credit_card\"
    }" &

curl -s -X POST "$BASE_URL/bookings" \
    -H "Content-Type: application/json" \
    -d "{
        \"guestName\": \"Deadlock Test 2\",
        \"guestEmail\": \"deadlock2@example.com\",
        \"guestPhone\": \"+1234567891\",
        \"roomId\": 2,
        \"checkInDate\": \"2024-12-01\",
        \"checkOutDate\": \"2024-12-05\",
        \"paymentMethod\": \"credit_card\"
    }" &

wait
echo ""

echo "✅ Load tests completed!"
echo ""
echo "💡 Check the server logs to see transaction details and any potential issues."