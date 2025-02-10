package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/handlers"
	"github.com/panuvitpnv/room-booking-api/internal/repository"
	"github.com/panuvitpnv/room-booking-api/internal/routes"
	"github.com/panuvitpnv/room-booking-api/internal/service"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"

	_ "github.com/panuvitpnv/room-booking-api/docs" // This imports the generated Swagger docs
	echoSwagger "github.com/swaggo/echo-swagger"
)

type echoServer struct {
	app    *echo.Echo
	db     databases.Database
	conf   *config.Config
	router *routes.Router
}

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

var (
	once   sync.Once
	server *echoServer
)

func NewEchoServer(conf *config.Config, db databases.Database) *echoServer {
	echoApp := echo.New()
	echoApp.Logger.SetLevel(log.DEBUG)

	echoApp.Validator = NewCustomValidator()

	once.Do(func() {
		// Initialize repositories
		bookingRepo := repository.NewBookingRepository(db.Connect())
		roomRepo := repository.NewRoomRepository(db.Connect())
		guestRepo := repository.NewGuestRepository(db.Connect())
		roomStatusRepo := repository.NewRoomStatusRepository(db.Connect())
		roomTypeRepo := repository.NewRoomTypeRepository(db.Connect())

		// Initialize services
		bookingService := service.NewBookingService(bookingRepo, roomRepo, guestRepo, roomStatusRepo)
		roomService := service.NewRoomService(roomRepo, roomTypeRepo)
		guestService := service.NewGuestService(guestRepo, bookingRepo)
		roomTypeService := service.NewRoomTypeService(roomTypeRepo)
		roomStatusService := service.NewRoomStatusService(roomStatusRepo, roomRepo)

		// Initialize handlers
		bookingHandler := handlers.NewBookingHandler(bookingService)
		roomHandler := handlers.NewRoomHandler(roomService, roomTypeService, roomStatusService)
		guestHandler := handlers.NewGuestHandler(guestService)

		// Initialize router
		router := routes.NewRouter(bookingHandler, roomHandler, guestHandler)

		server = &echoServer{
			app:    echoApp,
			db:     db,
			conf:   conf,
			router: router,
		}
	})

	return server
}

func (s *echoServer) Start() {
	corsMiddleware := getCORSMiddleware(s.conf.Server.AllowOrigins)
	bodyLimitMiddleware := getBodyLimitMiddleware(s.conf.Server.BodyLimit)
	timeOutMiddleware := getTimeOutMiddleware(s.conf.Server.Timeout)

	s.app.Use(middleware.Recover())
	s.app.Use(middleware.Logger())
	s.app.Use(corsMiddleware)
	s.app.Use(bodyLimitMiddleware)
	s.app.Use(timeOutMiddleware)

	// Setup health check
	s.app.GET("/v1/health", s.healthCheck)

	// Setup routes using the router
	s.router.SetupRoutes(s.app)

	// Swagger route
	s.app.GET("/swagger/*", echoSwagger.WrapHandler)

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	go s.gracefullyShutdown(quitCh)

	s.httpListening()
}

// Keep your existing methods
func (s *echoServer) httpListening() {
	url := fmt.Sprintf(":%d", s.conf.Server.Port)

	if err := s.app.Start(url); err != nil && err != http.ErrServerClosed {
		s.app.Logger.Fatalf("Error: %s", err.Error())
	}
}

func (s *echoServer) gracefullyShutdown(quitCh chan os.Signal) {
	ctx := context.Background()

	<-quitCh
	s.app.Logger.Info("Shutting down server...")

	if err := s.app.Shutdown(ctx); err != nil {
		s.app.Logger.Fatalf("Error: %s", err.Error())
	}
}

func (s *echoServer) healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// Keep your existing middleware functions
func getTimeOutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Request Timeout",
		Timeout:      timeout * time.Second,
	})
}

func getCORSMiddleware(allowOrigins []string) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: allowOrigins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	})
}

func getBodyLimitMiddleware(bodyLimit string) echo.MiddlewareFunc {
	return middleware.BodyLimit(bodyLimit)
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
