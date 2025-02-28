# Room Booking API Development Guidelines

## Build & Run Commands
- `make docker-build` - Build Docker images
- `make docker-up` - Start application with Docker Compose
- `make docker-down` - Stop and remove Docker containers
- `make logs` - View container logs
- `make swagger` - Generate Swagger API documentation
- `make migrate` - Run database migrations
- `make seed` - Seed database with initial data

## Testing
- Use `artillery/test-script.sh` for manual API testing
- Load testing: `artillery/logger_test.sh`
- Test helpers in `internal/test/helper.go`

## Code Style Guidelines
- **Error Handling**: Use custom `AppError` with HTTP status codes and contextual messages
- **Naming**: packages (lowercase), interfaces (PascalCase), structs (PascalCase), methods (CamelCase)
- **Imports**: Standard lib first, third-party second, internal packages last; alphabetical sorting
- **Architecture**: Follow handler → service → repository layered pattern
- **Documentation**: Add Swagger annotations for all API endpoints
- **Validation**: Use custom validators with predefined validation rules
- **Concurrency**: Use sync.Map and Mutex for resource locking when needed
- **DI**: Constructor-based dependency injection throughout the application