package main

import (
	"foxy/internal/env"
	"foxy/internal/server"
	"github.com/davidbyttow/govips/v2/vips"
)

func main() {
	vips.Startup(nil)
	defer vips.Shutdown()

	env.Boot()
	server.StartServer()
}
