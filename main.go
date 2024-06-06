package main

import (
	"foxy/internal/db"
	"foxy/internal/env"
	"foxy/internal/server"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
)

func main() {
	vips.Startup(nil)
	defer vips.Shutdown()

	env.Boot()
	err := db.Boot()
	if err != nil {
		log.Panic("Error booting db", err)
	}

	server.StartServer()
}
