package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/panuvitpnv/room-booking-api/internal/api/handlers"
	appMiddleware "github.com/panuvitpnv/room-booking-api/internal/api/middleware"
	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/models"
	"github.com/panuvitpnv/room-booking-api/internal/repositories"
	"github.com/panuvitpnv/room-booking-api/internal/services"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
	"gorm.io/gorm"
)

// EchoServer represents the web server
type EchoServer struct {
	config *config.Config
	echo   *echo.Echo
	db     *gorm.DB
}

// NewEchoServer creates a new server instance
func NewEchoServer(config *config.Config, db databases.Database) *EchoServer {
	e := echo.New()

	// Configure middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.Server.AllowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.BodyLimit(config.Server.BodyLimit))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: config.Server.Timeout * time.Second,
	}))

	// Get DB connection
	conn := db.Connect()

	// Add transaction middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Store the db instance in context for repositories
			c.Set("db", conn)
			return next(c)
		}
	})
	e.Use(appMiddleware.TransactionMiddleware(conn))

	return &EchoServer{
		config: config,
		echo:   e,
		db:     conn,
	}
}

// GetEcho returns the Echo instance for configuring routes and middleware
func (s *EchoServer) GetEcho() *echo.Echo {
	return s.echo
}

// Start starts the server
func (s *EchoServer) Start() {
	// Auto-migrate database models
	err := s.autoMigrate()
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	bookingRepo := repositories.NewBookingRepository(s.db)
	receiptRepo := repositories.NewReceiptRepository(s.db)
	roomRepo := repositories.NewRoomRepository(s.db)
	concurrentScenarios := repositories.NewConcurrentScenarios(s.db)

	// Initialize services
	bookingService := services.NewBookingService(s.db, bookingRepo)
	receiptService := services.NewReceiptService(s.db, receiptRepo)
	roomService := services.NewRoomService(s.db, roomRepo)
	concurrentService := services.NewConcurrentService(s.db, concurrentScenarios)

	// Initialize handlers
	bookingHandler := handlers.NewBookingHandler(bookingService)
	receiptHandler := handlers.NewReceiptHandler(receiptService)
	roomHandler := handlers.NewRoomHandler(roomService)
	concurrentHandler := handlers.NewConcurrentHandler(concurrentService)

	// Register routes
	bookingHandler.RegisterRoutes(s.echo)
	receiptHandler.RegisterRoutes(s.echo)
	roomHandler.RegisterRoutes(s.echo)
	concurrentHandler.RegisterRoutes(s.echo)

	// Add health check endpoint
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"name":   "room-booking-api",
		})
	})

	// Start server in a goroutine so we can handle graceful shutdown
	go func() {
		address := fmt.Sprintf(":%d", s.config.Server.Port)
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.echo.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.echo.Shutdown(ctx); err != nil {
		s.echo.Logger.Fatal(err)
	}
}

// autoMigrate automatically migrates the database
func (s *EchoServer) autoMigrate() error {
	return s.db.AutoMigrate(
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
