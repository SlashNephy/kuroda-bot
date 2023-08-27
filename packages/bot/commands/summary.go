package commands

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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

		type DebtData struct {
			Amount int
			Label  string
		}
		debts := map[string][]DebtData{}
		var messageRegex = regexp.MustCompile(`^(?:<@\d+>\s*)+([\d,]+)(?:\s*(.+))?$`)

		// コマンドが実行されたチャンネルでメッセージを全件取得する
		var page int
		var beforeID string
		for page < 5 {
			msgs, err := s.ChannelMessages(i.ChannelID, 100, beforeID, "", "")
			if err != nil {
				return err
			}

			if len(msgs) == 0 {
				break
			}

			for _, m := range msgs {
				if len(m.Mentions) == 0 {
					continue
				}

				// 改行ごとに借金フォーマットを探す
				for _, line := range strings.Split(m.Content, "\n") {
					match := messageRegex.FindStringSubmatch(line)
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
						label = match[2]
					}

					slog.Info("found message",
						slog.Int("amount", amount),
						slog.String("label", label),
						slog.Any("mentions", m.Mentions),
						slog.String("content", line),
					)

					for _, user := range m.Mentions {
						// 行にメンションが含まれている場合だけ集計
						if !strings.Contains(line, user.Mention()) {
							continue
						}

						debts[user.Mention()] = append(debts[user.Mention()], DebtData{
							Amount: amount,
							Label:  label,
						})
					}
				}
			}

			beforeID = msgs[len(msgs)-1].ID
			page++
		}

		if len(debts) == 0 {
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
		for mention, data := range debts {
			var sum int
			for _, d := range data {
				sum += d.Amount
			}

			fields = append(fields, fmt.Sprintf("%s\n%d", mention, sum))
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
