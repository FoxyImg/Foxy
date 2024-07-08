package server

import (
	"context"
	"errors"
	"foxy/internal/env"
	"foxy/internal/server/routes"
	"github.com/go-chi/chi/v5"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
	"golang.org/x/net/netutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

func StartDebugServer() {
	router := http.NewServeMux()

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		_ = http.ListenAndServe(":8086", router)
	}()
}

func originMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refererUrl := r.Header.Get("Referer")
		if refererUrl == "" {
			http.Error(w, "Not Allowed", http.StatusForbidden)
			return
		}

		parsedUrl, err := url.Parse(refererUrl)
		if err != nil {
			http.Error(w, "Not Allowed", http.StatusForbidden)
			return
		}

		if !slices.Contains(*env.FoxyEnvironment.AllowedOrigins, parsedUrl.Host) {
			http.Error(w, "Not Allowed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
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
			VaryBy: &throttled.VaryBy{
				Path: true,
			},
		}
	}

	router := chi.NewRouter()

	if len(*env.FoxyEnvironment.AllowedOrigins) > 0 {
		router.Use(originMiddleware)
	}

	if httpRateLimiter != nil {
		router.Use(httpRateLimiter.RateLimit)
	}

	routes.RegisterSourceRoutes(router)
	routes.RegisterPresetRoutes(router)
	routes.RegisterImageRoutes(router)

	server := &http.Server{
		Addr:         ":" + env.FoxyEnvironment.Port,
		Handler:      router,
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
