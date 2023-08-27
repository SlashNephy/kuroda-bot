package handlers

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/SlashNephy/kuroda-bot/commands"
)

var onInteractionCreate = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, c := range commands.Commands {
		if c.Command.Name == i.ApplicationCommandData().Name {
			if err := c.Handler(s, i); err != nil {
				slog.Error("failed to handle command", slog.Any("err", err))
			}

			break
		}
	}
}
