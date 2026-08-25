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
	userService := service.NewUserService(repository.NewUserRepository(pg.Pool))
	projectService := service.NewProjectService(repository.NewProjectRepository(pg.Pool))
	ticketService := service.NewTicketService(repository.NewTicketRepository(pg.Pool))
	sprintRepo := repository.NewSprintRepository(pg.Pool)
	sprintService := service.NewSprintService(sprintRepo)
	sprintEntryService := service.NewSprintEntryService(repository.NewSprintEntryRepository(pg.Pool), sprintRepo)
	dashboardService := service.NewDashboardService(repository.NewDashboardRepository(pg.Pool), sprintRepo)

	router := handler.NewRouter(handler.Handlers{
		Health:        handler.NewHealthHandler(healthService),
		Users:         handler.NewUserHandler(userService),
		Projects:      handler.NewProjectHandler(projectService),
		Tickets:       handler.NewTicketHandler(ticketService),
		Sprints:       handler.NewSprintHandler(sprintService),
		SprintEntries: handler.NewSprintEntryHandler(sprintEntryService),
		Dashboard:     handler.NewDashboardHandler(dashboardService),
	})

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
