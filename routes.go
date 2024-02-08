package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
		HTMLURL string `json:"html_url"`
		State   string `json:"state"`
		Title   string `json:"title"`
		User    struct {
			Login     string `json:"login"`
			ID        int64  `json:"id"`
			HTMLURL   string `json:"htm_url"`
			Title     string `json:"title"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		Body     string     `json:"body"`
		ClosedAt *time.Time `json:"closed_at"`
		MergedAt *time.Time `json:"merged_at"`
	} `json:"pull_request"`
	MergedBy struct {
		Login string `json:"login"`
	} `json:"merged_by"`
	Changes struct {
		Title struct {
			From string `json:"from"`
		} `json:"title"`
		Body struct {
			From string `json:"from"`
		} `json:"body"`
	} `json:"changes"`
	Repo struct {
		HTMLURL string `json:"html_url"`
	} `json:"repository"`
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

	switch payload.Action {

	case "opened":
		err = app.sendOpenedPullRequestMessage(payload)
		if err != nil {
			app.log.Error().Err(err).Msg("could not send pull request message")
			http.Error(
				w,
				"Error sending opened pull request message",
				http.StatusInternalServerError,
			)
		}
		return
	case "edited":
		err = app.sendEditedPullRequestMessage(payload)
		if err != nil {
			app.log.Error().Err(err).Msg("could not send pull request message")
			http.Error(
				w,
				"Error sending opened pull request message",
				http.StatusInternalServerError,
			)
		}
		return
	case "closed":
		err = app.sendClosedPullRequestMessage(payload)
		if err != nil {
			app.log.Error().Err(err).Msg("could not send pull request message")
			http.Error(
				w,
				"Error sending closed pull request message",
				http.StatusInternalServerError,
			)
		}
		return
	default:
		fmt.Println("any other pull request event")

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
