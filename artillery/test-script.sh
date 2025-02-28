#!/bin/bash

# Base URL of your API
BASE_URL="http://localhost:8080"

# Function to make a booking request
make_booking() {
    local room_num=$1
    local guest_id=$2
    
    curl -X POST "${BASE_URL}/bookings" \
    -H "Content-Type: application/json" \
    -d "{
        \"room_num\": $room_num,
        \"guest_id\": $guest_id,
        \"check_in_date\": \"2024-03-01\",
        \"check_out_date\": \"2024-03-03\"
    }" &
}

# Test 1: Concurrent bookings for the same room
echo "Test 1: Attempting concurrent bookings for the same room"
for i in {1..5}; do
    make_booking 101 1
done
wait
echo "Test 1 completed"

# Test 2: Check room availability
echo -e "\nTest 2: Checking room availability"
curl -X POST "${BASE_URL}/bookings/check-availability" \
-H "Content-Type: application/json" \
-d '{
    "room_num": 101,
    "check_in_date": "2024-03-01",
    "check_out_date": "2024-03-03"
}'

# Test 3: Transaction rollback test
echo -e "\nTest 3: Testing transaction rollback with invalid guest"
curl -X POST "${BASE_URL}/bookings" \
-H "Content-Type: application/json" \
-d '{
    "room_num": 102,
    "guest_id": 999,
    "check_in_date": "2024-03-01",
    "check_out_date": "2024-03-03"
}'

# Test 4: Multiple rooms booking test
echo -e "\nTest 4: Testing multiple room bookings simultaneously"
make_booking 201 1
make_booking 202 2
make_booking 203 3
wait
echo "Test 4 completed"