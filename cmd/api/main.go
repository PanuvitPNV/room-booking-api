package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/panuvitpnv/room-booking-api/internal/api/handlers"
	"github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	"github.com/panuvitpnv/room-booking-api/internal/api/routes"
	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/data"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"github.com/panuvitpnv/room-booking-api/internal/services"
	"github.com/panuvitpnv/room-booking-api/internal/utils"

	// Import your database package
	"github.com/panuvitpnv/room-booking-api/internal/databases"

	_ "github.com/panuvitpnv/room-booking-api/docs" // Import the generated docs
)

// @title Hotel Booking System API
// @version 1.0
// @description API for hotel booking system with transaction management and concurrency control
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg := config.ConfigGetting()

	// Initialize database
	log.Println("Connecting to database...")
	db := databases.NewPostgresDatabase(cfg.Database).Connect()

	// AutoMigrate database schema
	log.Println("Running database migrations...")
	if err := migrateDatabase(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed database with initial data
	if err := data.SeedDatabase(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize DB reference
	utils.SetDB(db)

	// Create lock manager for concurrency control
	lockManager := utils.NewLockManager(10 * time.Second)

	// Create repositories
	bookingRepo := repositories.NewBookingRepository(db)
	receiptRepo := repositories.NewReceiptRepository(db)
	roomRepo := repositories.NewRoomRepository(db)

	// Create services
	bookingService := services.NewBookingService(bookingRepo, roomRepo, lockManager)
	receiptService := services.NewReceiptService(receiptRepo, bookingRepo, lockManager)
	roomService := services.NewRoomService(roomRepo, lockManager)

	// Create handlers
	bookingHandler := handlers.NewBookingHandler(bookingService)
	receiptHandler := handlers.NewReceiptHandler(receiptService)
	roomHandler := handlers.NewRoomHandler(roomService)

	// Create Echo instance
	e := echo.New()

	// Setup middleware
	middleware.SetupMiddleware(e, cfg)

	// Setup routes
	routes.SetupRoutes(e, bookingHandler, receiptHandler, roomHandler)

	// Start server with graceful shutdown
	startServer(e, cfg)
}

// migrateDatabase runs database migrations
func migrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.RoomType{},
		&models.Facility{},
		&models.RoomFacility{},
		&models.Room{},
		&models.Booking{},
		&models.Receipt{},
		&models.RoomStatus{},
		&models.LastRunning{},
	)
}

// startServer starts the HTTP server with graceful shutdown
func startServer(e *echo.Echo, cfg *config.Config) {
	// Start server in a goroutine
	go func() {
		address := fmt.Sprintf(":%d", cfg.Server.Port)
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
