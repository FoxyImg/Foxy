package server

import (
	"context"
	"errors"
	"foxy/internal/env"
	"foxy/internal/server/routes"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
	"golang.org/x/net/netutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartDebugServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		_ = http.ListenAndServe(":8086", mux)
	}()
}

func StartServer() {
	var httpRateLimiter *throttled.HTTPRateLimiterCtx
	if *env.FoxyEnvironment.Concurrency > 0 {
		store, err := memstore.NewCtx(65536)
		if err != nil {
			log.Fatal(err)
		}

		quota := throttled.RateQuota{
			MaxRate:  throttled.PerMin(*env.FoxyEnvironment.Concurrency),
			MaxBurst: *env.FoxyEnvironment.MaxBurst,
		}
		rateLimiter, err := throttled.NewGCRARateLimiterCtx(store, quota)
		if err != nil {
			log.Fatal(err)
		}

		httpRateLimiter = &throttled.HTTPRateLimiterCtx{
			RateLimiter: rateLimiter,
			VaryBy:      &throttled.VaryBy{Path: true},
		}
	}

	mux := http.NewServeMux()

	routes.RegisterSourceRoutes(mux)
	routes.RegisterPresetRoutes(mux)
	routes.RegisterImageRoutes(mux, httpRateLimiter)

	server := &http.Server{
		Addr:         ":" + env.FoxyEnvironment.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(*env.FoxyEnvironment.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(*env.FoxyEnvironment.WriteTimeout) * time.Second,
		//IdleTimeout:  time.Duration(*env.FoxyEnvironment.IdleTimeout) * time.Second,
	}

	listener, err := net.Listen("tcp", ":"+env.FoxyEnvironment.Port)
	if err != nil {
		panic(err)
	}

	if *env.FoxyEnvironment.MaxConnections > 0 {
		listener = netutil.LimitListener(listener, *env.FoxyEnvironment.MaxConnections)
	}

	defer listener.Close()

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Printf("interrupted, shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v\n", err)
	}
}
