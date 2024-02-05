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
	ChannelID   string `mapstructure:"channel"`
	GuildID     string `mapstructure:"guild"`
	Environment string `mapstructure:"env"`
	Repo        string `mapstructure:"repo"`
}

func main() {

	app := application{
		cfg: config{},
		log: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	pflag.IntVar(&app.cfg.Port, "port", 6666, "HTTP port")
	pflag.StringVar(&app.cfg.Token, "token", "", "Discord bot token")
	pflag.StringVar(&app.cfg.ChannelID, "channel", "", "Discord channel id")
	pflag.StringVar(&app.cfg.GuildID, "guild", "", "Discord guild (server) id")
	pflag.StringVar(&app.cfg.Repo, "repo", "", "Link to GitHub repo with the pull requests")
	pflag.StringVar(&app.cfg.Environment, "environment", "dev", "dev | prod")

	configFileName := pflag.String("config", "", "config file name")

	pflag.Parse()

	viper.BindPFlag("port", pflag.Lookup("port"))
	viper.BindPFlag("token", pflag.Lookup("token"))
	viper.BindPFlag("channel", pflag.Lookup("channel"))
	viper.BindPFlag("guild", pflag.Lookup("guild"))
	viper.BindPFlag("environment", pflag.Lookup("environment"))

	if *configFileName != "" {
		err := app.setupConfigFile(*configFileName)
		if err != nil {
			app.log.Fatal().Err(err).Msg("error unmarshalling config")
		}
	}

	fmt.Printf("%+v\n\n", app.cfg)

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "repo",
			Description: "Link to GitHub repository",
		},
		{
			Name:        "echo",
			Description: "Replies with your input",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "input",

					Description: "The input to echo back",
					Required:    true,
				},
			},
		},
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"repo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: app.cfg.Repo,
				},
			})
		},
		"echo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			input := i.ApplicationCommandData().Options[0].StringValue()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: input,
				},
			})
		},
	}

	dgo, err := discordgo.New("Bot " + app.cfg.Token)
	if err != nil {
		app.log.Fatal().Err(err).Msg("Failed to create discord session")
	}
	defer dgo.Close()

	dgo.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(dgo, i)
		}
	})

	err = dgo.Open()
	if err != nil {
		app.log.Fatal().Err(err).Msg("Error opening websocket connection")
	}

	fmt.Println("Bot is now running. Press Ctrl+C to exit")

	app.dgo = dgo

	for _, v := range commands {
		_, err := app.dgo.ApplicationCommandCreate(app.dgo.State.User.ID, app.cfg.GuildID, v)
		if err != nil {
			app.log.Fatal().Err(err).Msg("Discord could not send command")
		}
	}

	app.server()
}

func (app *application) setupConfigFile(filename string) error {
	viper.SetConfigName(filename)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
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
