package main

import (
	"github.com/panuvitpnv/room-booking-api/docs"
	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/server"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
)

// @title Hotel Booking API
// @version 1.0
// @description This is a hotel room booking server with transaction management and concurrency control.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
func main() {
	conf := config.ConfigGetting()
	db := databases.NewPostgresDatabase(conf.Database)
	server := server.NewEchoServer(conf, db)

	// Setup Swagger documentation
	docs.SetupSwagger(server.GetEcho())

	server.Start()
}
