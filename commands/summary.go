package commands

import (
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

var (
	// 「@メンション 金額 メモ」形式
	amountFirstLineRegex = regexp.MustCompile(`^(?:<@\d+>\s*)+([\d,]+)円?(?:\s*(.+))?$`)
	// 「@メンション メモ 金額」形式 (金額とメモを書き間違えた場合の救済)
	labelFirstLineRegex = regexp.MustCompile(`^(?:<@\d+>\s*)+(.+?)\s+([\d,]+)円?$`)
)

// ParseDebtLine は 1 行から借金の金額とメモを取り出す。
// 金額とメモはどちらが先に書かれていてもよいが、金額を含まない行は借金フォーマットとみなさない。
func ParseDebtLine(line string) (*Debt, bool) {
	// 「@メンション 100 200」のような曖昧な行では、金額先行として解釈する
	if match := amountFirstLineRegex.FindStringSubmatch(line); match != nil {
		return newDebt(match[1], match[2])
	}

	if match := labelFirstLineRegex.FindStringSubmatch(line); match != nil {
		return newDebt(match[2], match[1])
	}

	return nil, false
}

func newDebt(rawAmount, rawLabel string) (*Debt, bool) {
	// カンマ消す
	amount, err := strconv.ParseUint(strings.ReplaceAll(rawAmount, ",", ""), 10, 64)
	if err != nil {
		slog.Warn("invalid amount", slog.String("amount", rawAmount))
		return nil, false
	}

	return &Debt{
		Amount: amount,
		Label:  strings.TrimSpace(rawLabel),
	}, true
}

var summary = &DiscordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "summary",
		Description: "Prints summary of current debts.",
		DescriptionLocalizations: &map[discordgo.Locale]string{
			discordgo.Japanese: "現在の借金のサマリーを出力します。",
		},
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		// 「考え中」を出す
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			return err
		}

		messages, err := FetchMessages(s, i.ChannelID, false)
		if err != nil {
			return err
		}

		embed, err := RenderSummaryMessageEmbed(messages)
		if err != nil {
			return err
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		return err
	},
}

type Summary struct {
	User  *discordgo.User
	Debts []*Debt
	Sum   uint64
}

type Debt struct {
	Amount  uint64
	Label   string
	Message *discordgo.Message
}

const maxPage = 5

func FetchMessages(s *discordgo.Session, channelID string, cleanup bool) ([]*discordgo.Message, error) {
	var messages []*discordgo.Message
	var page int
	var beforeID string
	for page < maxPage {
		msgs, err := s.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			return nil, err
		}

		if len(msgs) == 0 {
			break
		}

		messages = append(messages, msgs...)
		beforeID = messages[len(messages)-1].ID
		page++
	}

	defer func() {
		if !cleanup {
			return
		}

		slog.Info("clean up messages", slog.Int("count", len(messages)))

		if err := cleanupMessages(s, messages); err != nil {
			slog.Error("failed to clean up messages", slog.Any("err", err))
		}
	}()

	return messages, nil
}

func cleanupMessages(s *discordgo.Session, messages []*discordgo.Message) error {
	var messageIDs []string
	for _, message := range messages {
		if message.Author.ID != s.State.User.ID {
			continue
		}

		messageIDs = append(messageIDs, message.ID)
	}

	if len(messageIDs) == 0 {
		return nil
	}

	var eg errgroup.Group
	for _, messageID := range messageIDs {
		messageID := messageID
		eg.Go(func() error {
			// すべて同じチャンネルから送信されたメッセージであると仮定している
			return s.ChannelMessageDelete(messages[0].ChannelID, messageID)
		})
	}

	return eg.Wait()
}

func calculateSummaries(messages []*discordgo.Message) ([]*Summary, error) {
	usersByID := map[string]*discordgo.User{}
	summariesByUserID := map[string][]*Debt{}

	for _, message := range messages {
		if message.Author.Bot {
			continue
		}

		if len(message.Mentions) == 0 {
			continue
		}

		// 改行ごとに借金フォーマットを探す
		for _, line := range strings.Split(message.Content, "\n") {
			debt, ok := ParseDebtLine(line)
			if !ok {
				continue
			}

			debt.Message = message

			slog.Info("found message",
				slog.Uint64("amount", debt.Amount),
				slog.String("label", debt.Label),
				slog.Any("mentions", message.Mentions),
				slog.String("content", line),
			)

			for _, user := range message.Mentions {
				// 行にメンションが含まれている場合だけ集計
				if !strings.Contains(line, user.Mention()) {
					continue
				}

				usersByID[user.ID] = user
				summariesByUserID[user.ID] = append(summariesByUserID[user.ID], debt)
			}
		}
	}

	summaries := lo.MapToSlice(summariesByUserID, func(userID string, debts []*Debt) *Summary {
		return &Summary{
			User:  usersByID[userID],
			Debts: debts,
			Sum: lo.SumBy(debts, func(debt *Debt) uint64 {
				return debt.Amount
			}),
		}
	})
	// Sum の降順でソートする
	sort.SliceStable(summaries, func(i, j int) bool {
		return summaries[i].Sum > summaries[j].Sum
	})
	return summaries, nil
}

func RenderSummaryMessageEmbed(messages []*discordgo.Message) (*discordgo.MessageEmbed, error) {
	summaries, err := calculateSummaries(messages)
	if err != nil {
		return nil, err
	}

	if len(summaries) == 0 {
		return &discordgo.MessageEmbed{
			Description: "借金の履歴は見つかりませんでした。",
			Color:       0xD64B4B,
		}, nil
	}

	// ユーザーごとに借金の合計を表示
	var fields []string
	for _, summary := range summaries {
		var labels []string
		for _, debt := range summary.Debts {
			if debt.Label == "" {
				continue
			}

			labels = append(labels, debt.Label)
		}

		if len(labels) > 0 {
			fields = append(fields, fmt.Sprintf("%s\n%d (%s)", summary.User.Mention(), summary.Sum, strings.Join(labels, ", ")))
		} else {
			fields = append(fields, fmt.Sprintf("%s\n%d", summary.User.Mention(), summary.Sum))
		}
	}

	return &discordgo.MessageEmbed{
		// Fields を使うとメンションが反映されない
		Description: strings.Join(fields, "\n"),
		Color:       0xD64B4B,
	}, nil
}
