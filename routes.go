package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Get("/healthcheck", app.healthcheck)

	return router
}

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {

	data := map[string]any{
		"status":       "available",
		"port":         app.cfg.port,
		"version":      "1.0.0",
		"environment:": app.cfg.environment,
	}

	json, err := json.Marshal(data)
	if err != nil {
		app.log.Error().Err(err).Msg("Failed to marshal healthcheck data")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
