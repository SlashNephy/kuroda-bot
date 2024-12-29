package handlers

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"

	"github.com/SlashNephy/kuroda-bot/commands"
	"github.com/SlashNephy/kuroda-bot/config"
)

var onMessageCreate = func(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 自分自身は無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// 対象のチャンネルではない
	if !slices.Contains(config.ApplicationConfig.WatchChannelIDs, m.ChannelID) {
		return
	}

	// スレッドは許可
	if m.Thread != nil {
		return
	}

	// 行ごとに検査して、すべて正しい借金フォーマットである場合は許可
	ok := lo.EveryBy(strings.Split(m.Content, "\n"), func(line string) bool {
		return commands.MessageRegex.MatchString(line)
	})
	if ok {
		return
	}

	slog.Info(
		"received disallowed message",
		slog.String("username", m.Author.Username),
		slog.String("content", m.Content),
	)

	var dest *discordgo.Channel
	if config.ApplicationConfig.ConversationChannelID != "" {
		var err error
		dest, err = s.Channel(config.ApplicationConfig.ConversationChannelID)
		if err != nil {
			slog.Error("failed to get destination channel", slog.Any("err", err))
			return
		}
	}

	var content string
	if dest == nil {
		content = fmt.Sprintf("%s\n⚠️ 借金フォーマット (`@メンション 金額 メモ`) に従わないメッセージは削除します。\n\n削除されたメッセージ\n```\n%s\n```", m.Author.Mention(), m.Content)
	} else {
		content = fmt.Sprintf("%s\n⚠️ 借金フォーマット (`@メンション 金額 メモ`) に従わないメッセージは削除します。 %s に転送します。", m.Author.Mention(), dest.Mention())
	}

	sent, err := s.ChannelMessageSendReply(m.ChannelID, content, m.Reference())
	if err != nil {
		slog.Error("failed to send message", slog.Any("err", err))
		return
	}

	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		slog.Error("failed to delete message", slog.Any("err", err))
		return
	}

	if dest != nil {
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			slog.Error("failed to get member", slog.Any("err", err))
			return
		}

		channel, err := s.Channel(m.ChannelID)
		if err != nil {
			slog.Error("failed to get channel", slog.Any("err", err))
			return
		}

		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    effectiveName(m.Author, member),
				IconURL: member.AvatarURL("128"),
			},
			Description: m.Content,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("「%s」から転送", channel.Name),
			},
		}

		if len(m.Mentions) > 0 {
			var mentions []string
			for _, user := range m.Mentions {
				mentions = append(mentions, user.Mention())
			}

			data := &discordgo.MessageSend{
				Content: strings.Join(mentions, " "),
				Embed:   embed,
			}
			if _, err = s.ChannelMessageSendComplex(dest.ID, data); err != nil {
				slog.Error("failed to transfer message with reply", slog.Any("err", err))
				return
			}
		} else {
			if _, err = s.ChannelMessageSendEmbed(dest.ID, embed); err != nil {
				slog.Error("failed to transfer message", slog.Any("err", err))
				return
			}
		}
	}

	time.Sleep(30 * time.Second)

	if err = s.ChannelMessageDelete(m.ChannelID, sent.ID); err != nil {
		slog.Error("failed to delete message", slog.Any("err", err))
	}
}

func effectiveName(user *discordgo.User, member *discordgo.Member) string {
	if member == nil || member.Nick == "" {
		return user.Username
	}

	return fmt.Sprintf("%s (%s)", member.Nick, user.Username)
}
