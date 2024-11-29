package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/avointsev/yp7m-go/internal/flags"
	"github.com/avointsev/yp7m-go/internal/logger"
	"github.com/avointsev/yp7m-go/internal/server/handlers"
	"github.com/avointsev/yp7m-go/internal/server/storage"
)

func main() {
	config, err := flags.ParseServerConfig()
	if err != nil {
		log.Fatalf("%s: %v", logger.ErrFlagsParse, err)
	}

	store := storage.NewMemStorage()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", handlers.RootHandler(store))
	r.Get("/value/{type}/{name}", handlers.GetMetricHandler(store))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricHandler(store))

	log.Printf("%s on http://%s", logger.OkServerStarted, config.Address)
	if err := http.ListenAndServe(config.Address, r); err != nil {
		log.Fatalf("%s: %v", logger.ErrServerNotStarted, err)
	}
}
