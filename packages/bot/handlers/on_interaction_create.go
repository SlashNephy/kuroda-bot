package handlers

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"

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
