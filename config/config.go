package config

import (
	"context"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	koanfyaml "github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Config struct {
	Telegram Telegram
}

type Telegram struct {
	ApiKey string
	ChatID string
}

func LoadConfig(ctx context.Context) (*Config, error) {
	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	} else {
		zap.L().Info(".env file not found, using system environment variables")
	}

	k := koanfyaml.New(".")
	parser := yaml.Parser()
	configFile := "config.yaml"

	yamlContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "while reading config file")
	}

	yamlString := os.ExpandEnv(string(yamlContent))

	if err := k.Load(rawbytes.Provider([]byte(yamlString)), parser); err != nil {
		return nil, errors.Wrap(err, "while loading config file with replaced env vars")
	}

	var cfg Config
	err = k.Unmarshal("", &cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling config file")
	}

	if cfg.Telegram.ApiKey == "" {
		return nil, errors.New("Telegram bot api key can not be empty!")
	}

	if cfg.Telegram.ChatID == "" {
		return nil, errors.New("Telegram chat id can not be empty!")
	}
	return &cfg, nil
}
