package config

import (
	"github.com/caarlos0/env/v9"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DiscordBotToken string   `env:"DISCORD_BOT_TOKEN,required"`
	ChannelIDs      []string `env:"CHANNEL_IDS" envSeparator:","`
}

var ApplicationConfig *Config

func init() {
	var err error
	ApplicationConfig, err = LoadConfig()
	if err != nil {
		panic(err)
	}
}

func LoadConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
