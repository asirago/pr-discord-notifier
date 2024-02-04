package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type application struct {
	cfg config
	dgo *discordgo.Session
	log zerolog.Logger
}

type config struct {
	Port        int    `mapstructure:"port"`
	Token       string `mapstructure:"token"`
	ChannelID   string `mapstructure:"channelid"`
	Environment string `mapstructure:"environment"`
	Repo        string `mapstructure:"repo"`
}

func main() {

	app := application{
		cfg: config{},
		log: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	pflag.IntVar(&app.cfg.Port, "port", 6666, "HTTP port")
	pflag.StringVar(&app.cfg.Token, "token", "", "Discord bot token")
	pflag.StringVar(&app.cfg.ChannelID, "channelID", "", "Discord channel id")
	pflag.StringVar(&app.cfg.Environment, "env", "dev", "dev | prod")

	configFileName := pflag.String("config", "", "config file name")

	pflag.Parse()

	viper.BindPFlag("port", pflag.Lookup("port"))
	viper.BindPFlag("token", pflag.Lookup("token"))
	viper.BindPFlag("channelID", pflag.Lookup("channelID"))
	viper.BindPFlag("environment", pflag.Lookup("env"))

	if *configFileName != "" {
		err := app.setupConfigFile(*configFileName)
		if err != nil {
			app.log.Fatal().Err(err).Msg("error unmarshalling config")
		}
	}

	dgo, err := discordgo.New("Bot " + app.cfg.Token)
	if err != nil {
		app.log.Fatal().Err(err).Msg("Failed to create discord session")
	}

	app.dgo = dgo

	app.server()
}

func (app *application) setupConfigFile(filename string) error {
	viper.SetConfigName(filename)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		app.log.Warn().Msg("config file could not be found")
		return err
	}

	err = viper.Unmarshal(&app.cfg)
	if err != nil {
		app.log.Fatal().Err(err).Msg("error unmarshalling to config struct")
		return err
	}

	return nil
}

func (app *application) server() {

	var localhost string
	if app.cfg.Environment == "dev" {
		localhost = "localhost"
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", localhost, app.cfg.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.log.Info().
		Int("port", app.cfg.Port).
		Msg(fmt.Sprintf("Starting server on port %d", app.cfg.Port))

	err := server.ListenAndServe()
	if err != nil {
		app.log.Fatal().Err(err).Msg("Failed to start server")
	}
}
