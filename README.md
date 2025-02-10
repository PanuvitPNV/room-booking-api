
# Hotel Room Booking API

A Go-based API for managing hotel room bookings using Echo framework and GORM.

## Project Structure
```
├── README.md
├── artillery
│   ├── booking-test.yml
│   ├── complex-test.yml
│   ├── logger_test.sh
│   ├── realistic-booking-test.yml
│   └── test-data.csv
├── cmd
│   └── main.go
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   ├── config.go
│   │   └── config.yaml
│   ├── dto
│   │   ├── request
│   │   │   ├── booking_request.go
│   │   │   ├── guest_request.go
│   │   │   ├── request_parser.go
│   │   │   └── room_request.go
│   │   └── response
│   │       ├── booking_response.go
│   │       ├── common_response.go
│   │       ├── guest_response.go
│   │       └── room_response.go
│   ├── errors
│   │   ├── errors.go
│   │   ├── handler.go
│   │   └── logger.go
│   ├── handlers
│   │   ├── booking_handler.go
│   │   ├── converter.go
│   │   ├── guest_handler.go
│   │   └── room_handler.go
│   ├── models
│   │   └── models.go
│   ├── repository
│   │   ├── booking_repository.go
│   │   ├── guest_repository.go
│   │   ├── interfaces.go
│   │   ├── room_repository.go
│   │   ├── room_status_repository.go
│   │   └── room_type_repository.go
│   ├── routes
│   │   └── routes.go
│   ├── server
│   │   └── server.go
│   ├── service
│   │   ├── booking_service.go
│   │   ├── guest_service.go
│   │   ├── interfaces.go
│   │   ├── room_service.go
│   │   ├── room_status_service.go
│   │   └── room_type_service.go
│   ├── test
│   │   └── helper.go
│   ├── utils
│   └── validators
│       ├── request_validators.go
│       ├── validation_rules.go
│       └── validator.go
├── logs
│   └── {logs_file}.log
└── pkg
    ├── databases
    │   ├── database.go
    │   └── postgresDatabase.go
    ├── migration
    │   └── migration.go
    ├── seeder
    │   └── seeder.go
    └── test
        └── test.go
```

## Setup and Installation

To set up and run the Hotel Room Booking API, follow the instructions below: