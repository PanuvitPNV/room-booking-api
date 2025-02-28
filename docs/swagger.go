package docs

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title           Hotel Booking API
// @version         1.0
// @description     API for a hotel room booking system with transaction management and concurrency control.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.basic  BasicAuth

// SetupSwagger initializes and registers swagger routes
func SetupSwagger(e *echo.Echo) {
	// Serve the API documentation UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
