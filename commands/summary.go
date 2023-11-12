package commands

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
)

var MessageRegex = regexp.MustCompile(`^(?:<@\d+>\s*)+(-?[\d,]+)(?:\s*(.+))?$`)

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

		summaries, err := CalculateSummaries(s, i.ChannelID)
		if err != nil {
			return err
		}

		if len(summaries) == 0 {
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Description: "借金の履歴は見つかりませんでした。",
						Color:       0xD64B4B,
					},
				},
			})
			return err
		}

		// ユーザーごとに借金の合計を表示
		var fields []string
		for _, summary := range summaries {
			sum := lo.SumBy(summary.Debts, func(debt *Debt) int {
				return debt.Amount
			})

			var labels []string
			for _, debt := range summary.Debts {
				if debt.Label == "" {
					continue
				}

				labels = append(labels, debt.Label)
			}

			if len(labels) > 0 {
				fields = append(fields, fmt.Sprintf("%s\n%d (%s)", summary.User, sum, strings.Join(labels, ", ")))
			} else {
				fields = append(fields, fmt.Sprintf("%s\n%d", summary.User, sum))
			}
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					// Fields を使うとメンションが反映されない
					Description: strings.Join(fields, "\n"),
					Color:       0xD64B4B,
				},
			},
		})
		return err
	},
}

type Summary struct {
	User  *discordgo.User
	Debts []*Debt
}

type Debt struct {
	Amount  int
	Label   string
	Message *discordgo.Message
}

func CalculateSummaries(s *discordgo.Session, channelID string) ([]*Summary, error) {
	summaries := map[*discordgo.User][]*Debt{}

	// コマンドが実行されたチャンネルでメッセージを全件取得する
	var page int
	var beforeID string
	for page < 5 {
		messages, err := s.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			return nil, err
		}

		if len(messages) == 0 {
			break
		}

		for _, message := range messages {
			if len(message.Mentions) == 0 {
				continue
			}

			// 改行ごとに借金フォーマットを探す
			for _, line := range strings.Split(message.Content, "\n") {
				match := MessageRegex.FindStringSubmatch(line)
				if len(match) == 0 {
					continue
				}

				// カンマ消す
				s := strings.Replace(match[1], ",", "", -1)

				// 借金の金額
				amount, err := strconv.Atoi(s)
				if err != nil {
					slog.Warn("invalid amount", slog.String("content", line))
					continue
				}

				// 借金のラベル
				var label string
				if len(match) == 3 {
					label = strings.TrimSpace(match[2])
				}

				slog.Info("found message",
					slog.Int("amount", amount),
					slog.String("label", label),
					slog.Any("mentions", message.Mentions),
					slog.String("content", line),
				)

				for _, user := range message.Mentions {
					// 行にメンションが含まれている場合だけ集計
					if !strings.Contains(line, user.Mention()) {
						continue
					}

					summaries[user] = append(summaries[user], &Debt{
						Amount:  amount,
						Label:   label,
						Message: message,
					})
				}
			}
		}

		beforeID = messages[len(messages)-1].ID
		page++
	}

	return lo.MapToSlice(summaries, func(user *discordgo.User, debts []*Debt) *Summary {
		return &Summary{
			User:  user,
			Debts: debts,
		}
	}), nil
}
