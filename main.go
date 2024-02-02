package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type application struct {
	cfg config
	dgo *discordgo.Session
	log zerolog.Logger
}

type config struct {
	port        int
	token       string
	channelID   string
	environment string
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "p", 8080, "Listen to port...")
	flag.StringVar(&cfg.token, "token", "", "Discord bot token")
	flag.StringVar(&cfg.channelID, "chanID", "", "Discord channel id")
	flag.StringVar(&cfg.environment, "env", "dev", "dev | prod")

	flag.Parse()

	app := application{
		cfg: cfg,
		log: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	dgo, err := discordgo.New("Bot " + app.cfg.token)
	if err != nil {
		app.log.Fatal().Err(err).Msg("Failed to create discord session")
	}

	app.dgo = dgo

	app.server()

}

func (app *application) server() {

	var localhost string
	if app.cfg.environment == "dev" {
		localhost = "localhost"
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", localhost, app.cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.log.Info().
		Int("port", app.cfg.port).
		Msg(fmt.Sprintf("Starting server on port %d", app.cfg.port))

	err := server.ListenAndServe()
	if err != nil {
		app.log.Fatal().Err(err).Msg("Failed to start server")
	}
}
