package server

import (
	"foxy/internal/env"
	"foxy/internal/server/routes"
	"net/http"
)

func StartServer() {
	mux := http.NewServeMux()

	routes.RegisterSourceRoutes(mux)
	routes.RegisterPresetRoutes(mux)
	routes.RegisterImageRoutes(mux)

	server := &http.Server{
		Addr:         ":" + env.FoxyEnvironment.Port,
		Handler:      mux,
		ReadTimeout:  env.FoxyEnvironment.ReadTimeout,
		WriteTimeout: env.FoxyEnvironment.WriteTimeout,
		IdleTimeout:  env.FoxyEnvironment.IdleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
