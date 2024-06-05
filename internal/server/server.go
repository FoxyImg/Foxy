package server

import (
	"foxy/internal/env"
	"foxy/internal/server/routes"
	"net/http"
)

func StartServer() {
	mux := http.NewServeMux()

	if env.FoxyEnvironment.AllowPresetManagement {
		routes.RegisterSourceRoutes(mux)
		routes.RegisterPresetRoutes(mux)
	}

	routes.RegisterImageRoutes(mux)

	server := &http.Server{
		Addr:    ":" + env.FoxyEnvironment.Port,
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
