# Hotel Booking System

A Go-based hotel booking system that demonstrates transaction management and concurrency control using the Echo framework. This application showcases robust handling of concurrent booking operations with various locking strategies and transaction management techniques.

## Features

- Room management (types, facilities, and availability)
- Booking management (create, update, cancel)
- Payment processing (receipts, refunds)
- Transaction management with ACID compliance
- Concurrency control with both optimistic and pessimistic locking strategies
- Application-level and database-level locking mechanisms
- Deadlock prevention, detection, and testing
- Comprehensive concurrent transaction testing suite
- Transaction timeline visualization for test analysis

## Project Structure

```
hotel-booking-system/
├── cmd/
│   ├── api/
│   │   └── main.go                  # API server entry point
│   ├── deadlock_tests/
│   │   └── main.go                  # Dedicated deadlock testing tool
│   └── txtests/
│       └── main.go                  # Concurrency testing tool
├── docs/
│   ├── docs.go                      # Swagger documentation
│   ├── swagger.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── api/
│   │   ├── handlers/                # HTTP request handlers
│   │   │   ├── booking_handler.go
│   │   │   ├── receipt_handler.go
│   │   │   ├── room_handler.go
│   │   │   └── test_handler.go      # Test-specific endpoints
│   │   ├── middleware/              # HTTP middleware
│   │   │   ├── logger_middleware.go
│   │   │   └── middleware.go
│   │   └── routes/                  # API route definitions
│   │       └── routes.go
│   ├── config/                      # Application configuration
│   │   ├── config.go
│   │   └── config.yaml
│   ├── data/                        # Database seeding
│   │   └── seeds.go
│   ├── databases/                   # Database connections
│   │   ├── database.go
│   │   └── postgresDatabase.go
│   ├── models/                      # Data models
│   │   └── models.go
│   ├── repositories/                # Database access layer
│   │   ├── booking_repository.go
│   │   ├── receipt_repository.go
│   │   └── room_repository.go
│   ├── services/                    # Business logic layer
│   │   ├── booking_service.go
│   │   ├── receipt_service.go
│   │   └── room_service.go
│   └── utils/                       # Utility functions
│       ├── database.go              # Database utilities
│       ├── deadlock_util.go         # Deadlock testing utilities
│       ├── lock_manager.go          # Concurrency control
│       ├── logger.go                # Logging utilities
│       └── transaction_id.go        # Transaction tracking
├── logs/                            # Application logs
├── pkg/                             # Reusable packages
│   └── report/                      # Test reporting tools
│       ├── template_funcs.go        # Template helper functions
│       └── timeline.go              # Timeline visualization generator
├── test-results/                    # Output from concurrency tests
├── web/                             # Web assets and templates
│   ├── static/                      # Static files (CSS, JS, images)
│   │   ├── css/
│   │   ├── img/
│   │   └── js/
│   └── templates/                   # HTML templates
│       ├── reports/                 # Report templates
│       │   └── timeline.html        # Transaction timeline visualization
│       └── ... (other templates)
├── go.mod                           # Go module definition
└── README.md
```

## Technology Stack

- **Backend**: Go language with Echo framework
- **Database**: PostgreSQL with GORM ORM
- **Documentation**: Swagger
- **Testing**: Custom concurrency and deadlock testing frameworks
- **Visualization**: Transaction timeline visualization

## Transaction Management Features

### Database Transactions
- ACID compliance for all critical operations
- Automatic rollback on errors
- Transaction retry with exponential backoff for transient failures
- Transaction propagation across service layers

### Concurrency Control
- **Optimistic Locking**: Using version fields for non-critical operations
- **Pessimistic Locking**: Using row locks and advisory locks for critical operations
- **Application-Level Locking**: Custom LockManager implementation
- **Two-Phase Locking**: For complex operations involving multiple resources

### Deadlock Prevention and Testing
- Consistent lock acquisition order
- Lock timeouts to prevent indefinite waiting
- Lock acquisition retries with backoff
- Deadlock detection and resolution
- Dedicated deadlock testing framework
- Test mode for reproducing and analyzing deadlock scenarios

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

### Test Endpoints
- `GET /test/status` - Check the status of the test controller
- `GET /test/deadlock` - Trigger a deadlock test scenario
- `GET /test/concurrent-bookings/:roomNum/:count` - Run a concurrent booking test

## How to Run

### Prerequisites
- Go 1.22 or higher
- PostgreSQL database

### Setup and Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/PanuvitPNV/room-booking-api
   cd room-booking-api
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Configure the PostgreSQL database connection in `internal/config/config.yaml`

4. Run database migrations and seed data:
   ```bash
   go run cmd/api/main.go --migrate --seed
   ```

5. Start the API server:
   ```bash
   go run cmd/api/main.go
   ```

6. The API will be available at http://localhost:8080

## Testing

### Concurrency Testing
The system includes a comprehensive concurrency testing tool that simulates multiple clients interacting with the booking system simultaneously.

```bash
go run cmd/txtests/main.go [baseURL] [scenario]
```

Available scenarios:
- `peak_booking_rush` - Simulates many clients booking during peak periods
- `weekend_availability_race` - Tests race conditions for weekend availability
- `payment_processing_surge` - Tests concurrent payment processing
- `cancellation_and_rebooking` - Tests cancellation and immediate rebooking
- `booking_modification_conflicts` - Tests concurrent modifications
- `mixed_clients` - Runs clients with different behavior patterns
- `all` - Runs all scenarios in sequence

### Deadlock Testing
The system includes a dedicated deadlock testing tool that can trigger and analyze deadlock scenarios:

```bash
# Run dedicated deadlock tests
DEADLOCK_TEST_MODE=true ENABLE_DEADLOCK_MODE=true go run cmd/deadlock_tests/main.go

# Run the API server in deadlock test mode
DEADLOCK_TEST_MODE=true ENABLE_DEADLOCK_MODE=true go run cmd/api/main.go
```

The deadlock tests include:
- Guaranteed deadlock scenarios at the database level
- Cross-update deadlock scenarios with transactions in opposing orders
- Aggressive concurrent booking tests with high contention

### Test Visualization

After running the concurrency tests, a transaction timeline visualization is generated in the `test-results` directory. This HTML report provides an interactive timeline of all operations, showing:

- Timeline of all transactions with color-coding by type and status
- Concurrent operations and conflicts
- Details of individual transactions including timing and results
- Room-specific analysis showing contention patterns
- Client behavior analysis

To view the visualization:
1. Open the generated HTML file in `test-results/timeline_*.html`
2. Use the interactive filters to analyze different aspects of the test run
3. Click on events to see detailed information

## License

This project is licensed under the [MIT License](LICENSE) - see the [LICENSE](LICENSE) file for details.