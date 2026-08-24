package handler

import "net/http"

func NewRouter(health *HealthHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Live)
	mux.HandleFunc("GET /readyz", health.Ready)
	return mux
}
