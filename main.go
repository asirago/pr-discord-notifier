package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
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
	Project     string `mapstructure:"project"`
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
	pflag.StringVar(&app.cfg.Project, "project", "", "Link to GitHub project")
	pflag.StringVar(&app.cfg.Environment, "environment", "dev", "dev | prod")

	configFileName := pflag.String("config", "", "config file name")

	pflag.Parse()

	viper.BindPFlag("port", pflag.Lookup("port"))
	viper.BindPFlag("token", pflag.Lookup("token"))
	viper.BindPFlag("channel", pflag.Lookup("channel"))
	viper.BindPFlag("guild", pflag.Lookup("guild"))
	viper.BindPFlag("repo", pflag.Lookup("repo"))
	viper.BindPFlag("project", pflag.Lookup("project"))
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
			Name:        "set_repo",
			Description: "Set link to GitHub repository",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "repo",

					Description: "Link to GitHub repository",
					Required:    true,
				},
			},
		},
		{
			Name:        "project",
			Description: "Link to project management e.g GitHub project kanban | trello",
		},
		{
			Name:        "set_project",
			Description: "Set link to project manager",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "set_project",

					Description: "Link to project",
					Required:    true,
				},
			},
		},
		{
			Name:        "source",
			Description: "Link to GitHub repository with source code",
		},
		{
			Name:        "channel",
			Description: "Pull request notifications will be sent to this channel",
		},
		{
			Name:        "set_channel",
			Description: "Set the channel id or #channel for PR notification",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "channel",

					Description: "#channel or channel id",
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
		"set_repo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			repo := i.ApplicationCommandData().Options[0].StringValue()

			app.cfg.Repo = repo
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Repo was set to %s", app.cfg.Repo),
				},
			})
		},
		"project": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: app.cfg.Project,
				},
			})
		},
		"set_project": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			project := i.ApplicationCommandData().Options[0].StringValue()
			app.cfg.Project = project
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Project was set to %s", app.cfg.Project),
				},
			})
		},
		"source": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "https://github.com/asirago/pr-discord-notifier",
				},
			})
		},
		"channel": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("<#%s>", app.cfg.ChannelID),
				},
			})
		},
		// TODO: Fix error handling when dealing with invalid inputs
		"set_channel": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			channelID := i.ApplicationCommandData().Options[0].StringValue()
			re := regexp.MustCompile(`(\d+)`)
			app.cfg.ChannelID = re.FindString(channelID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel was set to <#%s>", app.cfg.ChannelID),
				},
			})
		},
	}

	dgo, err := discordgo.New("Bot " + app.cfg.Token)
	if err != nil {
		app.log.Fatal().Err(err).Msg("Failed to create discord session")
	}

	dgo.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(dgo, i)
		}
	})

	err = dgo.Open()
	if err != nil {
		app.log.Fatal().Err(err).Msg("Error opening websocket connection")
	}
	defer dgo.Close()

	fmt.Println("Bot is now running. Press Ctrl+C to exit")

	app.dgo = dgo

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := app.dgo.ApplicationCommandCreate(app.dgo.State.User.ID, app.cfg.GuildID, v)
		if err != nil {
			app.log.Fatal().Err(err).Msg("Discord could not send command")
		}
		registeredCommands[i] = cmd
	}

	app.server()

	fmt.Println("Removing commands...")

	for _, v := range registeredCommands {
		err := dgo.ApplicationCommandDelete(dgo.State.User.ID, app.cfg.GuildID, v.ID)
		if err != nil {
			app.log.Panic().Err(err).Msg("Cannot delete commands")
		}
	}

	app.log.Info().Msg("Gracefully shutting down")
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

func (app *application) server() error {

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

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.log.Info().Str("signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		fmt.Println()
		app.log.Info().Msg("Closing server...")
		time.Sleep(1 * time.Second)

		shutdownError <- nil

	}()

	app.log.Info().
		Int("port", app.cfg.Port).
		Msg(fmt.Sprintf("Starting server on port %d", app.cfg.Port))

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	fmt.Println()
	app.log.Info().Str("port", fmt.Sprintf("%d", app.cfg.Port)).Msg("Stopped server")

	return nil
}
