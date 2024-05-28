package server

import (
	"foxy/internal/env"
	"foxy/internal/server/routes"
	"net/http"
)

func StartServer() {
	if env.FoxyEnvironment.AllowPresetManagement {
		routes.RegisterSourceRoutes()
		routes.RegisterPresetRoutes()
	}

	routes.RegisterImageRoutes()

	err := http.ListenAndServe(":"+env.FoxyEnvironment.Port, nil)
	if err != nil {
		panic(err)
	}
}
