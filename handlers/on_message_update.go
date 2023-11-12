package handlers

import (
	"log/slog"
	"slices"

	"github.com/bwmarrin/discordgo"

	"github.com/SlashNephy/kuroda-bot/commands"
	"github.com/SlashNephy/kuroda-bot/config"
)

var onMessageUpdate = func(s *discordgo.Session, m *discordgo.MessageUpdate) {
	// 自分自身は無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// 対象のチャンネルではない
	if !slices.Contains(config.ApplicationConfig.WatchChannelIDs, m.ChannelID) {
		return
	}

	// 借金フォーマットではないのは無視
	if !commands.MessageRegex.MatchString(m.Content) {
		return
	}

	// メッセージが更新されたときにサマリーを投稿する
	if err := PostSummaryMessage(s, m.ChannelID); err != nil {
		slog.Error("failed to post summary message", slog.Any("err", err))
		return
	}
}
