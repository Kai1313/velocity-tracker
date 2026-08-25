package handler

import "net/http"

type Handlers struct {
	Health        *HealthHandler
	Users         *UserHandler
	Projects      *ProjectHandler
	Tickets       *TicketHandler
	Sprints       *SprintHandler
	SprintEntries *SprintEntryHandler
	Dashboard     *DashboardHandler
}

func NewRouter(h Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.Health.Live)
	mux.HandleFunc("GET /readyz", h.Health.Ready)

	mux.HandleFunc("POST /users", h.Users.Create)
	mux.HandleFunc("GET /users", h.Users.List)
	mux.HandleFunc("GET /users/{id}", h.Users.Get)
	mux.HandleFunc("PUT /users/{id}", h.Users.Update)
	mux.HandleFunc("DELETE /users/{id}", h.Users.Delete)

	mux.HandleFunc("POST /projects", h.Projects.Create)
	mux.HandleFunc("GET /projects", h.Projects.List)
	mux.HandleFunc("GET /projects/{id}", h.Projects.Get)
	mux.HandleFunc("PUT /projects/{id}", h.Projects.Update)
	mux.HandleFunc("DELETE /projects/{id}", h.Projects.Delete)

	mux.HandleFunc("POST /tickets", h.Tickets.Create)
	mux.HandleFunc("GET /tickets", h.Tickets.List)
	mux.HandleFunc("GET /tickets/{id}", h.Tickets.Get)
	mux.HandleFunc("PUT /tickets/{id}", h.Tickets.Update)
	mux.HandleFunc("DELETE /tickets/{id}", h.Tickets.Delete)

	mux.HandleFunc("POST /sprints", h.Sprints.Create)
	mux.HandleFunc("GET /sprints", h.Sprints.List)
	mux.HandleFunc("GET /sprints/{id}", h.Sprints.Get)
	mux.HandleFunc("PUT /sprints/{id}", h.Sprints.Update)
	mux.HandleFunc("DELETE /sprints/{id}", h.Sprints.Delete)

	mux.HandleFunc("POST /sprint-entries", h.SprintEntries.Create)
	mux.HandleFunc("GET /sprint-entries", h.SprintEntries.List)
	mux.HandleFunc("GET /sprint-entries/{id}", h.SprintEntries.Get)
	mux.HandleFunc("PUT /sprint-entries/{id}", h.SprintEntries.Update)
	mux.HandleFunc("DELETE /sprint-entries/{id}", h.SprintEntries.Delete)

	mux.HandleFunc("GET /dashboard/sprints", h.Dashboard.SprintSummaries)
	mux.HandleFunc("GET /dashboard/sprints/{id}", h.Dashboard.SprintDeveloperBreakdown)

	return mux
}
