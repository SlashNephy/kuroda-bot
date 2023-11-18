package handlers

import (
	"slices"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/SlashNephy/kuroda-bot/config"
)

type MessageCache struct {
	cache map[string]*discordgo.Message
	lock  sync.Mutex
}

func (c *MessageCache) OnMessageCreate(_ *discordgo.Session, m *discordgo.MessageCreate) {
	// 対象のチャンネルではない
	if !slices.Contains(config.ApplicationConfig.WatchChannelIDs, m.ChannelID) {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache[m.ID] = m.Message
}

func (c *MessageCache) OnMessageUpdate(_ *discordgo.Session, m *discordgo.MessageUpdate) {
	// 対象のチャンネルではない
	if !slices.Contains(config.ApplicationConfig.WatchChannelIDs, m.ChannelID) {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache[m.ID] = m.Message
}

func (c *MessageCache) Pop(messageID string) (*discordgo.Message, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	message, ok := c.cache[messageID]
	delete(c.cache, messageID)
	return message, ok
}

var messageCache = &MessageCache{
	cache: map[string]*discordgo.Message{},
}
