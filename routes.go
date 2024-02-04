package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Get("/healthcheck", app.healthcheck)
	router.Post("/github-webhook-receiver", app.githubWebhookReceiver)

	return router
}

type Payload struct {
	Action      string `json:"action"`
	Number      int64  `json:"number"`
	PullRequest struct {
		URL     string `json:"url"`
		ID      int64  `json:"id"`
		HTMLURL string `json:"htmlurl"`
		State   string `json:"state"`
		User    struct {
			Login string `json:"login"`
			Title string `json:"title"`
		} `json:"user"`
		Body string `json:"body"`
	} `json:"pull_request"`
}

func (app *application) githubWebhookReceiver(w http.ResponseWriter, r *http.Request) {
	app.log.Info().Msg("Received a webhook request")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
	}

	var payload Payload

	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Error unmarshalling json", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {

	data := map[string]any{
		"status":       "available",
		"port":         app.cfg.Port,
		"version":      "1.0.0",
		"environment:": app.cfg.Environment,
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
