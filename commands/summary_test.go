package commands_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SlashNephy/kuroda-bot/commands"
)

func TestParseDebtLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		line       string
		wantOK     bool
		wantAmount uint64
		wantLabel  string
	}{
		{
			name:       "金額のみ",
			line:       "<@1> 3400",
			wantOK:     true,
			wantAmount: 3400,
		},
		{
			name:       "金額先行",
			line:       "<@1> 3400 焼肉",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "メモ先行",
			line:       "<@1> 焼肉 3400",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "メモ先行 数字を含むメモ",
			line:       "<@1> ライブ2日目のチケット 2815",
			wantOK:     true,
			wantAmount: 2815,
			wantLabel:  "ライブ2日目のチケット",
		},
		{
			name:       "メモ先行 カンマ区切りの金額",
			line:       "<@1> 焼肉 3,400",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "メモ先行 円付きの金額",
			line:       "<@1> 焼肉 3400円",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "メモ先行 行末に空白がある",
			line:       "<@1> 焼肉 3400 ",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "金額先行 行末に空白がある",
			line:       "<@1> 3400 焼肉 ",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "金額のみ 行末に空白がある",
			line:       "<@1> 3400 ",
			wantOK:     true,
			wantAmount: 3400,
		},
		{
			name:       "金額先行 数字を含むメモ",
			line:       "<@1> 30566 (焼肉, ブック, ベスト10, よもだ, スキー)",
			wantOK:     true,
			wantAmount: 30566,
			wantLabel:  "(焼肉, ブック, ベスト10, よもだ, スキー)",
		},
		{
			name:       "複数メンション メモ先行",
			line:       "<@1> <@2> 焼肉 3400",
			wantOK:     true,
			wantAmount: 3400,
			wantLabel:  "焼肉",
		},
		{
			name:       "数値が 2 つ並ぶ場合は金額先行を優先する",
			line:       "<@1> 100 200",
			wantOK:     true,
			wantAmount: 100,
			wantLabel:  "200",
		},
		{
			name:   "メモのみで金額がない",
			line:   "<@1> 焼肉",
			wantOK: false,
		},
		{
			name:   "数字を含むメモのみで金額がない",
			line:   "<@1> ベスト10",
			wantOK: false,
		},
		{
			name:   "メンションがない",
			line:   "3400 焼肉",
			wantOK: false,
		},
		{
			name:   "空行",
			line:   "",
			wantOK: false,
		},
		{
			name:   "金額が uint64 に収まらない",
			line:   "<@1> 99999999999999999999999 焼肉",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			debt, ok := commands.ParseDebtLine(tt.line)
			if !assert.Equal(t, tt.wantOK, ok) || !tt.wantOK {
				return
			}

			assert.Equal(t, tt.wantAmount, debt.Amount)
			assert.Equal(t, tt.wantLabel, debt.Label)
		})
	}
}

func TestRenderSummaryMessageEmbed(t *testing.T) {
	t.Parallel()

	alice := &discordgo.User{ID: "1", Username: "alice"}
	bob := &discordgo.User{ID: "2", Username: "bob"}

	tests := []struct {
		name            string
		messages        []*discordgo.Message
		wantDescription string
	}{
		{
			name: "金額先行とメモ先行が混在していても集計する",
			messages: []*discordgo.Message{
				{
					Author:   &discordgo.User{ID: "99"},
					Mentions: []*discordgo.User{alice},
					Content:  "<@1> 3400 焼肉",
				},
				{
					Author:   &discordgo.User{ID: "99"},
					Mentions: []*discordgo.User{alice},
					Content:  "<@1> ビール 500",
				},
			},
			wantDescription: "<@1>\n3900 (焼肉, ビール)",
		},
		{
			name: "複数行のメッセージを行ごとに集計する",
			messages: []*discordgo.Message{
				{
					Author:   &discordgo.User{ID: "99"},
					Mentions: []*discordgo.User{alice, bob},
					Content:  "<@1> 映画 1500\n<@2> 900 (ビール, ラーメン)",
				},
			},
			wantDescription: "<@1>\n1500 (映画)\n<@2>\n900 ((ビール, ラーメン))",
		},
		{
			name: "bot の投稿は集計しない",
			messages: []*discordgo.Message{
				{
					Author:   &discordgo.User{ID: "99", Bot: true},
					Mentions: []*discordgo.User{alice},
					Content:  "<@1> 焼肉 3400",
				},
			},
			wantDescription: "借金の履歴は見つかりませんでした。",
		},
		{
			name: "金額のないメモのみの行は集計しない",
			messages: []*discordgo.Message{
				{
					Author:   &discordgo.User{ID: "99"},
					Mentions: []*discordgo.User{alice},
					Content:  "<@1> 焼肉",
				},
			},
			wantDescription: "借金の履歴は見つかりませんでした。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			embed, err := commands.RenderSummaryMessageEmbed(tt.messages)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDescription, embed.Description)
		})
	}
}
