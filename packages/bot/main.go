package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"

	"github.com/SlashNephy/kuroda-bot/commands"
	"github.com/SlashNephy/kuroda-bot/config"
	"github.com/SlashNephy/kuroda-bot/handlers"
)

func main() {
	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.ApplicationConfig.DiscordBotToken))
	if err != nil {
		panic(err)
	}

	for _, handler := range handlers.Handlers {
		session.AddHandler(handler)
	}

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
