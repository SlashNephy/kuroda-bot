package handlers

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type MessageCache struct {
	cache map[string]*discordgo.Message
	lock  sync.Mutex
}

func (c *MessageCache) OnMessageCreate(_ *discordgo.Session, m *discordgo.MessageCreate) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache[m.ID] = m.Message
}

func (c *MessageCache) OnMessageUpdate(_ *discordgo.Session, m *discordgo.MessageUpdate) {
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
