package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"

	"github.com/SlashNephy/kuroda-bot/commands"
	"github.com/SlashNephy/kuroda-bot/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	session, err := discordgo.New(fmt.Sprintf("Bot %s", cfg.DiscordBotToken))
	if err != nil {
		panic(err)
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info("successfully logged in",
			slog.String("username", s.State.User.Username),
			slog.String("discriminator", s.State.User.Discriminator),
		)
	})

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		for _, c := range commands.Commands {
			if c.Command.Name == i.ApplicationCommandData().Name {
				if err = c.Handler(s, i); err != nil {
					slog.Error("failed to handle command", slog.Any("err", err))
				}

				return
			}
		}
	})

	if err = session.Open(); err != nil {
		panic(err)
	}
	defer session.Close()

	for _, c := range commands.Commands {
		if _, err = session.ApplicationCommandCreate(session.State.User.ID, "", c.Command); err != nil {
			slog.Error("failed to create command",
				slog.String("command", c.Command.Name),
				slog.Any("err", err),
			)
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	slog.Info("press Ctrl+C to exit")
	<-stop

	slog.Info("gracefully shutting down...")
}
