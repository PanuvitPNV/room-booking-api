package main

import (
	"github.com/panuvitpnv/room-booking-api/internal/config"
	"github.com/panuvitpnv/room-booking-api/internal/server"
	"github.com/panuvitpnv/room-booking-api/pkg/databases"
)

// @Title         Hotel Booking System API
// @Version       1.0
// @Description   Hotel Booking System API
// @Host          localhost:8080
// @BasePath      /v1
func main() {
	conf := config.ConfigGetting()
	db := databases.NewPostgresDatabase(conf.Database)
	server := server.NewEchoServer(conf, db)

	server.Start()
}
