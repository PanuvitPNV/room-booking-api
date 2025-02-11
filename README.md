# Hotel Room Booking API

A Go-based API for managing hotel room bookings using Echo framework and GORM.

## Setup and Installation

To set up and run the Hotel Room Booking API, follow the instructions below:

### 1. **Clone the Repository**

First, clone the repository to your local machine:

```bash
git clone https://github.com/PanuvitPNV/room-booking-api
cd room-booking-api
```

### 2. **Install Dependencies**

Ensure you have the required dependencies installed:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### 3. **Build Docker Containers**

Before starting the app, build the Docker containers:

```bash
make docker-build
```

This will build the Docker images based on the `Dockerfile` configurations in your `.docker` directory.

### 4. **Start the Application**

To start the application, migrate the database, and seed the initial data, simply run:

```bash
make docker-up
```

This command will:

- Start the application and its services in detached mode.
- Automatically run database migrations.
- Automatically seed the database with test data.

### 5. **Set Up Configuration (Optional)**

You can set up a `.env` file to configure your environment variables (like database credentials, etc.). This is optional — if the `.env` file is not set, the application will use default values. To create the `.env` file, copy the example:

```bash
cp .env.example .env
```

Update the `.env` file as per your environment.

### 6. **View Logs**

To view the logs of the running containers in real-time, run:

```bash
make logs
```

This will stream the logs from all containers, helping you monitor the application.

### 7. **Optional Commands**

You can also run the following commands for additional functionality:

- **Generate Swagger Documentation**:  
  To generate the Swagger API documentation, run:
  ```bash
  make swagger
  ```

- **Run Database Migrations Manually**:  
  If you need to apply database migrations manually, run:
  ```bash
  make migrate
  ```

- **Seed Database Manually**:  
  To seed the database with initial data manually, run:
  ```bash
  make seed
  ```

- **Stop the Containers**:  
  To stop and remove the containers, run:
  ```bash
  make docker-down
  ```

---

## Project Structure

Below is the general structure of the project:

```
.
├── .air.toml
├── .docker
│   ├── Dockerfile
│   ├── Dockerfile.migration
│   ├── Dockerfile.seeder
│   ├── docker-compose.yml
│   ├── docs
│   ├── logs
│   └── wait-for-postgres.sh
├── .env
├── .gitignore
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
├── makefile
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

---

## Contributing

We welcome contributions to this project! To contribute, please fork the repository and create a pull request with your proposed changes.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
