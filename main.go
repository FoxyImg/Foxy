package main

import (
	"foxy/internal/cli"
	"foxy/internal/db"
	"foxy/internal/env"
	"foxy/internal/onnx"
	"foxy/internal/params"
	"github.com/alecthomas/kong"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"os"
)

func main() {
	log.Println(os.Executable())
	log.Println(os.Getwd())

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

	cmd := kong.Parse(&cli.CLI)
	if cmd.Error != nil {
		log.Fatal(cmd.Error)
	}

	err = cmd.Run(&cli.Context{})
	cmd.FatalIfErrorf(err)
	//server.StartServer()
}
