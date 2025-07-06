# Hotel Booking API

A simple hotel booking API built with Node.js, TypeScript, Express, and PostgreSQL for learning database transactions, locking mechanisms, and concurrency control.

## Features

- **Hotel Room Booking**: Create bookings with guest information
- **Payment Processing**: Simulate payment processing with transaction IDs
- **Receipt Generation**: Generate receipts for successful bookings
- **Database Transactions**: ACID compliance with rollback on failures
- **Row Locking**: Configurable row locking to prevent race conditions
- **Concurrency Testing**: Tools to test deadlock scenarios
- **Comprehensive Logging**: Built-in logger for debugging
- **Docker Support**: Containerized PostgreSQL with Docker Compose

## Prerequisites

- **Node.js** (v16 or higher)
- **Docker** and **Docker Compose**
- **curl** and **jq** (for demo scripts)

## Quick Start

1. **Setup the project:**
   ```bash
   make setup
   ```

2. **Start the development server:**
   ```bash
   make dev
   ```

3. **Run the demo:**
   ```bash
   make demo
   ```

## Docker Services

The project includes these Docker services:

- **PostgreSQL**: Database server on port 5432
- **Adminer**: Web-based database management UI on port 8080

### Docker Commands

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs

# Access database shell
make db-shell

# Open Adminer web UI
make adminer
```

## API Endpoints

### Bookings
- `POST /api/bookings` - Create a new booking
- `GET /api/bookings/:id` - Get booking details
- `DELETE /api/bookings/:id` - Cancel a booking

### Settings
- `POST /api/settings/row-locking` - Enable/disable row locking

### Health Check
- `GET /health` - Server health status

## Database Schema

The system uses 5 main tables:
- `guests` - Guest information
- `rooms` - Room details and availability
- `bookings` - Booking records
- `payments` - Payment transactions
- `receipts` - Generated receipts

## Learning Scenarios

### 1. Normal Transaction Flow
```bash
make demo
```

### 2. Concurrent Booking Tests
```bash
make load-test
```

### 3. Stress Testing
```bash
make stress-test
```

### 4. Database Monitoring
```bash
make monitor
```

## Row Locking Demonstration

The API allows you to enable/disable row locking to observe different behaviors:

**With Row Locking (Enabled):**
- Prevents race conditions
- Ensures data consistency
- One booking succeeds, others wait or fail

**Without Row Locking (Disabled):**
- Potential race conditions
- May allow double bookings
- Demonstrates need for proper locking

## Example Usage

### Create a Booking
```bash
curl -X POST http://localhost:3000/api/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "guestName": "John Doe",
    "guestEmail": "john@example.com",
    "guestPhone": "+1234567890",
    "roomId": 1,
    "checkInDate": "2024-12-01",
    "checkOutDate": "2024-12-05",
    "paymentMethod": "credit_card"
  }'
```

### Toggle Row Locking
```bash
# Enable row locking
curl -X POST http://localhost:3000/api/settings/row-locking \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'

# Disable row locking
curl -X POST http://localhost:3000/api/settings/row-locking \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

## Testing Scenarios

### 1. Successful Booking Flow
- Guest creation/lookup
- Room availability check
- Booking creation
- Payment processing
- Receipt generation

### 2. Transaction Rollback
- Payment failure scenarios
- Room availability restoration
- Data consistency verification

### 3. Concurrency Control
- Multiple simultaneous bookings
- Deadlock detection
- Lock timeout handling

### 4. Race Conditions
- Demonstrates what happens without proper locking
- Shows data inconsistency issues
- Emphasizes importance of transaction isolation

## Available Commands

### Development
- `make setup` - Complete project setup with Docker
- `make dev` - Start development server
- `make test` - Run unit tests
- `make build` - Build the project

### Docker Management
- `make docker-up` - Start PostgreSQL and Adminer
- `make docker-down` - Stop all Docker services
- `make docker-logs` - View Docker container logs
- `make db-shell` - Open PostgreSQL shell
- `make adminer` - Instructions for Adminer access

### Testing & Monitoring
- `make demo` - Run interactive demo
- `make load-test` - Test concurrent bookings
- `make stress-test` - High-volume testing
- `make monitor` - Real-time database monitoring

## Database Management

### Adminer Web UI
Visit http://localhost:8080 and use these credentials:
- **System**: PostgreSQL
- **Server**: postgres
- **Username**: postgres
- **Password**: password
- **Database**: hotel_booking

### Direct Database Access
```bash
# Open PostgreSQL shell
make db-shell

# Or use docker-compose directly
docker-compose exec postgres psql -U postgres -d hotel_booking
```

## Environment Variables

The project uses these environment variables:

```bash
# Database connection
DATABASE_URL=postgresql://postgres:password@localhost:5432/hotel_booking
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hotel_booking
DB_USER=postgres
DB_PASSWORD=password

# Server configuration
PORT=3000
```

## Learning Objectives

This project demonstrates:
1. **ACID Properties** - Atomicity, Consistency, Isolation, Durability
2. **Transaction Management** - BEGIN, COMMIT, ROLLBACK
3. **Concurrency Control** - Row locking, deadlock prevention
4. **Race Conditions** - What happens without proper locking
5. **Database Design** - Proper foreign key relationships
6. **Error Handling** - Graceful failure and recovery
7. **Docker Integration** - Containerized database services

Perfect for understanding database fundamentals and transaction processing with modern development practices!