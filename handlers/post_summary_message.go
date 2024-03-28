package handlers

import (
	"log/slog"
	"slices"

	"github.com/bwmarrin/discordgo"

	"github.com/SlashNephy/kuroda-bot/commands"
	"github.com/SlashNephy/kuroda-bot/config"
)

type PostSummaryMessage struct{}

func (p *PostSummaryMessage) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// メッセージが投稿されたときにサマリーを投稿する
	if err := p.PostSummaryMessage(s, m.Message); err != nil {
		slog.Error("failed to post summary message", slog.Any("err", err))
	}
}

func (p *PostSummaryMessage) OnMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	// メッセージが更新されたときにサマリーを投稿する
	if err := p.PostSummaryMessage(s, m.Message); err != nil {
		slog.Error("failed to post summary message", slog.Any("err", err))
	}
}

func (p *PostSummaryMessage) OnMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	// ここでは ID しか参照できないのでキャッシュを参照する
	message, ok := messageCache.Pop(m.ID)
	if !ok {
		return
	}

	// メッセージが削除されたときにサマリーを投稿する
	if err := p.PostSummaryMessage(s, message); err != nil {
		slog.Error("failed to post summary message", slog.Any("err", err))
	}
}

func (p *PostSummaryMessage) shouldPostSummaryMessage(s *discordgo.Session, message *discordgo.Message) bool {
	// Webhook / bot は無視
	if message.Author == nil || message.Author.Bot {
		return false
	}

	// 自分自身は無視
	if message.Author.ID == s.State.User.ID {
		return false
	}

	// 対象のチャンネルではない
	if !slices.Contains(config.ApplicationConfig.WatchChannelIDs, message.ChannelID) {
		return false
	}

	// 借金フォーマットではないのは無視
	return commands.MessageRegex.MatchString(message.Content)
}

func (p *PostSummaryMessage) PostSummaryMessage(s *discordgo.Session, message *discordgo.Message) error {
	if !p.shouldPostSummaryMessage(s, message) {
		return nil
	}

	messages, err := commands.FetchMessages(s, message.ChannelID, true)
	if err != nil {
		return err
	}

	embed, err := commands.RenderSummaryMessageEmbed(messages)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSendEmbed(message.ChannelID, embed)
	return err
}

var postSummaryMessage = &PostSummaryMessage{}
