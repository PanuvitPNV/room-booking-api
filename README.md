
# Hotel Room Booking API

A Go-based API for managing hotel room bookings using Echo framework and GORM.

## Project Structure
```
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   └── models.go
│   ├── handlers/
│   │   └── booking.go
│   ├── repository/
│   │   └── booking.go
│   ├── service/
│   │   └── booking.go
│   └── utils/
│       └── errors.go
├── pkg/
│   ├── database/
│   │   └── db.go
│   ├── migration/
│   │   └── migration.go
│   ├── seeder/
│   │   └── seeder.go
│   └── test/
│       └── test.go
├── go.mod
├── go.sum
└── README.md
```

## Setup and Installation

1. Clone the repository
```bash
git clone <your-repository-url>
```

2. Install dependencies
```bash
go mod tidy
```

3. Run database migration
```bash
go run pkg/migration/migration.go
```

4. Seed sample data
```bash
go run pkg/seeder/seeder.go
```

5. Run tests
```bash
go run pkg/test/test.go
```