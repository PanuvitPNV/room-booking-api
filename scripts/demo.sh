#!/bin/bash

BASE_URL="http://localhost:3000/api"

echo "🏨 Hotel Booking API Demo"
echo "========================="
echo ""

# Check if PostgreSQL container is running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "❌ PostgreSQL container is not running. Start it with: docker-compose up -d postgres"
    exit 1
fi

# Check if server is running
if ! curl -s "$BASE_URL/../health" > /dev/null; then
    echo "❌ Server is not running. Please start the server first with 'npm run dev'"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Demo 1: Successful booking
echo "Demo 1: Creating a successful booking"
echo "------------------------------------"
BOOKING_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
    -H "Content-Type: application/json" \
    -d '{
        "guestName": "John Doe",
        "guestEmail": "john.doe@example.com",
        "guestPhone": "+1234567890",
        "roomId": 1,
        "checkInDate": "2024-12-01",
        "checkOutDate": "2024-12-05",
        "paymentMethod": "credit_card"
    }')

echo "Response:"
echo "$BOOKING_RESPONSE" | jq '.'
echo ""

# Extract booking ID
BOOKING_ID=$(echo "$BOOKING_RESPONSE" | jq -r '.data.booking.id // empty')

if [ ! -z "$BOOKING_ID" ]; then
    echo "✅ Booking created successfully with ID: $BOOKING_ID"
    echo ""
    
    # Demo 2: Get booking details
    echo "Demo 2: Retrieving booking details"
    echo "----------------------------------"
    curl -s "$BASE_URL/bookings/$BOOKING_ID" | jq '.'
    echo ""
    
    # Demo 3: Try to book the same room (should fail)
    echo "Demo 3: Attempting to book the same room (should fail)"
    echo "-----------------------------------------------------"
    curl -s -X POST "$BASE_URL/bookings" \
        -H "Content-Type: application/json" \
        -d '{
            "guestName": "Jane Smith",
            "guestEmail": "jane.smith@example.com",
            "guestPhone": "+1234567891",
            "roomId": 1,
            "checkInDate": "2024-12-02",
            "checkOutDate": "2024-12-06",
            "paymentMethod": "credit_card"
        }' | jq '.'
    echo ""
    
    # Demo 4: Cancel booking
    echo "Demo 4: Cancelling the booking"
    echo "------------------------------"
    curl -s -X DELETE "$BASE_URL/bookings/$BOOKING_ID" | jq '.'
    echo ""
    
    # Demo 5: Try to book the same room again (should succeed now)
    echo "Demo 5: Booking the same room after cancellation (should succeed)"
    echo "----------------------------------------------------------------"
    curl -s -X POST "$BASE_URL/bookings" \
        -H "Content-Type: application/json" \
        -d '{
            "guestName": "Jane Smith",
            "guestEmail": "jane.smith@example.com",
            "guestPhone": "+1234567891",
            "roomId": 1,
            "checkInDate": "2024-12-02",
            "checkOutDate": "2024-12-06",
            "paymentMethod": "credit_card"
        }' | jq '.'
    echo ""
    
else
    echo "❌ Failed to create booking"
fi

# Demo 6: Row locking demonstration
echo "Demo 6: Row locking demonstration"
echo "--------------------------------"
echo "Disabling row locking..."
curl -s -X POST "$BASE_URL/settings/row-locking" \
    -H "Content-Type: application/json" \
    -d '{"enabled": false}' | jq '.'
echo ""

echo "Enabling row locking..."
curl -s -X POST "$BASE_URL/settings/row-locking" \
    -H "Content-Type: application/json" \
    -d '{"enabled": true}' | jq '.'
echo ""

echo "✅ Demo completed!"
echo ""
echo "💡 Tips:"
echo "   - Check server logs to see transaction details"
echo "   - Use './scripts/load-test.sh' to test concurrent bookings"
echo "   - Use './scripts/monitor.sh' to monitor database activity"
echo "   - Visit http://localhost:8080 for Adminer database UI"
echo "   - Use 'docker-compose logs postgres' to see PostgreSQL logs"