package main

import (
	"context"
	"log"
	"net/http"

	"velocity-tracker/backend/internal/config"
	"velocity-tracker/backend/internal/handler"
	"velocity-tracker/backend/internal/repository"
	"velocity-tracker/backend/internal/service"
)

func main() {
	cfg := config.Load()

	if err := repository.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	pg, err := repository.NewPostgres(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to postgres failed: %v", err)
	}
	defer pg.Close()

	healthService := service.NewHealthService(pg)
	healthHandler := handler.NewHealthHandler(healthService)
	router := handler.NewRouter(healthHandler)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
