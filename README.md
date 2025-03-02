# Hotel Booking System

This is a Go application that demonstrates transaction management and concurrency control in a hotel room booking system. The application uses the Echo framework for the API and GORM for database operations.

## Features

- Room management (types, facilities, and availability)
- Booking management (create, update, cancel)
- Payment processing (receipts, refunds)
- Robust transaction management
- Concurrency control with application-level locking and database-level locking
- Optimistic and pessimistic concurrency strategies

## Project Structure

```
hotel-booking-system/
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point
├── config/
│   └── config.go                    # Application configuration
├── internal/
│   ├── api/
│   │   ├── handlers/                # HTTP request handlers
│   │   ├── middleware/              # HTTP middleware
│   │   └── routes/                  # API route definitions
│   ├── data/
│   │   └── seeds.go                 # Database seeding
│   ├── repositories/                # Database access layer
│   ├── services/                    # Business logic layer
│   └── utils/
│       ├── database.go              # Database utilities
│       └── lock_manager.go          # Concurrency control
├── models/
│   └── models.go                    # Data models
├── go.mod
└── README.md
```

## Technology Stack

- Go language
- Echo web framework
- GORM ORM library
- PostgreSQL database

## Transaction Management Features

1. **Database Transactions**:
   - ACID compliance for critical operations
   - Automatic rollback on errors
   - Transaction retry with exponential backoff

2. **Concurrency Control**:
   - Optimistic locking for non-critical operations
   - Pessimistic locking for critical operations
   - Application-level locking via LockManager

3. **Deadlock Prevention**:
   - Consistent lock acquisition order
   - Lock timeouts
   - Lock acquisition retries

## API Endpoints

### Rooms
- `GET /api/v1/rooms` - Get all rooms
- `GET /api/v1/rooms/{roomNum}` - Get room by number
- `POST /api/v1/rooms/{roomNum}/status` - Get room status for date range
- `GET /api/v1/rooms/types` - Get all room types
- `GET /api/v1/rooms/type/{typeId}` - Get rooms by type
- `POST /api/v1/rooms/availability` - Get room availability summary
- `POST /api/v1/rooms/available` - Get available rooms for booking

### Bookings
- `POST /api/v1/bookings` - Create a booking
- `GET /api/v1/bookings/{id}` - Get booking by ID
- `DELETE /api/v1/bookings/{id}` - Cancel a booking
- `PUT /api/v1/bookings/{id}` - Update a booking
- `POST /api/v1/bookings/by-date` - Get bookings by date range

### Receipts
- `POST /api/v1/receipts` - Create a receipt (process payment)
- `GET /api/v1/receipts` - Get all receipts with pagination
- `GET /api/v1/receipts/{id}` - Get receipt by ID
- `GET /api/v1/receipts/booking/{bookingId}` - Get receipt by booking ID
- `POST /api/v1/receipts/refund` - Process a refund
- `POST /api/v1/receipts/by-date` - Get receipts by date range

## How to Run

1. Set up PostgreSQL database
2. Configure connection in `config.yaml`
3. Run the application:

```bash
go run cmd/api/main.go
```

## Concurrency Testing

To test concurrency control, you can use a tool like Apache Bench or hey:

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test concurrent booking attempts
hey -n 100 -c 10 -m POST -H "Content-Type: application/json" -d '{"booking_name":"Test","room_num":101,"check_in_date":"2023-04-01T00:00:00Z","check_out_date":"2023-04-03T00:00:00Z"}' http://localhost:8080/api/v1/bookings
```