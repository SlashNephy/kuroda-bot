package handlers

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

var onReady = func(s *discordgo.Session, r *discordgo.Ready) {
	slog.Info("successfully logged in",
		slog.String("username", s.State.User.Username),
		slog.String("discriminator", s.State.User.Discriminator),
	)
}
