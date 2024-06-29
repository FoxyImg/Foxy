package server

import (
	"foxy/internal/env"
	"foxy/internal/server/routes"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"log"
	"net/http"
	"time"
)

func StartServer() {
	store, err := memorystore.New(&memorystore.Config{
		// Number of tokens allowed per interval.
		Tokens: uint64(*env.FoxyEnvironment.TokensPerMinute),

		// Interval until tokens reset.
		Interval: time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}

	var rateLimiter *httplimit.Middleware
	if *env.FoxyEnvironment.UseRateLimiter {
		rateLimiter, err = httplimit.NewMiddleware(store, httplimit.IPKeyFunc("x-forwarded-for", "x-real-ip", "remote-addr"))
		if err != nil {
			log.Fatal(err)
		}
	}

	mux := http.NewServeMux()

	routes.RegisterSourceRoutes(mux, rateLimiter)
	routes.RegisterPresetRoutes(mux, rateLimiter)
	routes.RegisterImageRoutes(mux, rateLimiter)

	server := &http.Server{
		Addr:         ":" + env.FoxyEnvironment.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(*env.FoxyEnvironment.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(*env.FoxyEnvironment.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(*env.FoxyEnvironment.IdleTimeout) * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
