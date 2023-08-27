package handlers

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

var onReady = func(s *discordgo.Session, r *discordgo.Ready) {
	slog.Info("successfully logged in",
		slog.String("username", s.State.User.Username),
		slog.String("discriminator", s.State.User.Discriminator),
	)
}
