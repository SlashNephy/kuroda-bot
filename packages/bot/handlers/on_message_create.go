package handlers

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/SlashNephy/kuroda-bot/commands"
	"github.com/SlashNephy/kuroda-bot/config"
)

var onMessageCreate = func(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 自分自身は無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// 対象のチャンネルではない
	if !slices.Contains(config.ApplicationConfig.ChannelIDs, m.ChannelID) {
		return
	}

	// スレッドは許可
	if m.Thread != nil {
		return
	}

	// 借金フォーマットである場合は許可
	if commands.MessageRegex.MatchString(m.Content) {
		return
	}

	slog.Info(
		"received disallowed message",
		slog.String("username", m.Author.Username),
		slog.String("content", m.Content),
	)

	content := fmt.Sprintf("%s\n⚠️ 借金フォーマット (`@メンション 金額 メモ`) に従わないメッセージは削除します。\n\n削除されたメッセージ\n```\n%s\n```", m.Author.Mention(), m.Content)
	sent, err := s.ChannelMessageSendReply(m.ChannelID, content, m.Reference())
	if err != nil {
		slog.Error("failed to send message", slog.Any("err", err))
		return
	}

	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		slog.Error("failed to delete message", slog.Any("err", err))
		return
	}

	time.Sleep(30 * time.Second)

	if err = s.ChannelMessageDelete(m.ChannelID, sent.ID); err != nil {
		slog.Error("failed to delete message", slog.Any("err", err))
	}
}
