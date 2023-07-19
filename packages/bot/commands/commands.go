package commands

import "github.com/bwmarrin/discordgo"

type DiscordCommand struct {
	Command *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

var Commands = []*DiscordCommand{
	summary,
}
