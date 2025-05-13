// Package main for server.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/avointsev/yp7m-go/internal/flags"
	"github.com/avointsev/yp7m-go/internal/server/handlers"
	"github.com/avointsev/yp7m-go/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

func main() {
	config, err := flags.ParseServerConfig()
	if err != nil {
		log.Fatalf("%s: %v", "Failed to parse arguments", err)
	}

	store := storage.NewMemStorage()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", handlers.RootHandler(store))
	r.Get("/value/{type}/{name}", handlers.GetMetricHandler(store))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricHandler(store))

	srv := &http.Server{
		Addr:         config.Address,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Printf("%s on http://%s", "Server started", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%s: %v", "Server can't be started", err)
	}
	// log.Printf("%s on http://%s", "Server started", config.Address)
	// if err := http.ListenAndServe(config.Address, r); err != nil {
	// 	log.Fatalf("%s: %v", "Server can't be started", err)
	// }
}
