package main

import (
	"foxy/internal/db"
	"foxy/internal/env"
	"foxy/internal/onnx"
	"foxy/internal/params"
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
		log.Panic("Error booting db: ", err)
	}

	if env.FoxyEnvironment.UseML != nil && *env.FoxyEnvironment.UseML {
		err = onnx.Setup()
		if err != nil {
			log.Panic("Error setting up onnx: ", err)
		}

		err = onnx.InitBackgroundRemoval()
		if err != nil {
			log.Panic("Error initializing background removal: ", err)
		}
	}

	err = params.Boot()
	if err != nil {
		log.Panic("Error booting params: ", err)
	}

	server.StartServer()
}
