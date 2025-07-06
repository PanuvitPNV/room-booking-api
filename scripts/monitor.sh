#!/bin/bash

echo "🔍 Database Activity Monitor"
echo "==========================="
echo ""

# Check if Docker container is running
if ! docker-compose ps postgres | grep -q "Up"; then
    echo "❌ PostgreSQL container is not running. Start it with: docker-compose up -d postgres"
    exit 1
fi

# Check if PostgreSQL is accessible
if ! docker-compose exec postgres psql -U postgres -d hotel_booking -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ Cannot connect to PostgreSQL database"
    exit 1
fi

echo "✅ Connected to database"
echo ""

# Function to show current bookings
show_bookings() {
    echo "📋 Current Bookings:"
    echo "-------------------"
    docker-compose exec postgres psql -U postgres -d hotel_booking -c "
        SELECT 
            b.id,
            g.name as guest_name,
            r.room_number,
            b.check_in_date,
            b.check_out_date,
            b.total_amount,
            b.status
        FROM bookings b
        JOIN guests g ON b.guest_id = g.id
        JOIN rooms r ON b.room_id = r.id
        ORDER BY b.created_at DESC;
    " -t -A -F'|'
    echo ""
}

# Function to show room availability
show_room_availability() {
    echo "🏠 Room Availability:"
    echo "--------------------"
    docker-compose exec postgres psql -U postgres -d hotel_booking -c "
        SELECT 
            id,
            room_number,
            room_type,
            price_per_night,
            CASE WHEN is_available THEN 'Available' ELSE 'Booked' END as status
        FROM rooms
        ORDER BY room_number;
    " -t -A -F'|'
    echo ""
}

# Function to show recent transactions
show_recent_transactions() {
    echo "💳 Recent Transactions:"
    echo "----------------------"
    docker-compose exec postgres psql -U postgres -d hotel_booking -c "
        SELECT 
            p.id,
            p.transaction_id,
            p.amount,
            p.payment_method,
            p.status,
            p.created_at
        FROM payments p
        ORDER BY p.created_at DESC
        LIMIT 10;
    " -t -A -F'|'
    echo ""
}

# Function to show database locks
show_locks() {
    echo "🔒 Current Database Locks:"
    echo "-------------------------"
    docker-compose exec postgres psql -U postgres -d hotel_booking -c "
        SELECT 
            l.locktype,
            l.database,
            l.relation::regclass,
            l.page,
            l.tuple,
            l.virtualxid,
            l.transactionid,
            l.mode,
            l.granted,
            a.application_name,
            a.client_addr,
            a.state,
            a.query
        FROM pg_locks l
        LEFT JOIN pg_stat_activity a ON l.pid = a.pid
        WHERE l.database = (SELECT oid FROM pg_database WHERE datname = 'hotel_booking');
    " -t -A -F'|'
    echo ""
}

# Main monitoring loop
while true; do
    clear
    echo "🔍 Database Activity Monitor - $(date)"
    echo "====================================="
    echo ""
    
    show_bookings
    show_room_availability
    show_recent_transactions
    show_locks
    
    echo "Press Ctrl+C to exit, or wait 5 seconds for refresh..."
    sleep 5
done