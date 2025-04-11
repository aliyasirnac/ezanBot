package bot

import (
	"context"

	"github.com/aliyasirnac/ezanBot/config"
	"github.com/aliyasirnac/ezanBot/thirdparty"
	botApi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Bot struct {
	Config  *config.Telegram
	EzanAPI *thirdparty.ThirdParty[thirdparty.EzanRequest, thirdparty.EzanResponse]
	bot     *botApi.Bot
}

func New(config *config.Telegram, api *thirdparty.ThirdParty[thirdparty.EzanRequest, thirdparty.EzanResponse]) *Bot {
	return &Bot{Config: config, EzanAPI: api}
}

func (b *Bot) Start(ctx context.Context) error {
	opts := []botApi.Option{
		botApi.WithDefaultHandler(handler),
	}

	bot, err := botApi.New(b.Config.ApiKey, opts...)
	if err != nil {
		zap.L().Error("Bot could not started", zap.Error(err))
		return err
	}

	b.bot = bot
	b.bot.Start(ctx)

	return nil
}

func (b *Bot) SendMessage(ctx context.Context, message string) {
	if b.bot == nil {
		zap.L().Error("Bot is not initialized")
		return
	}

	_, err := b.bot.SendMessage(ctx, &botApi.SendMessageParams{
		ChatID: b.Config.ChatID,
		Text:   message,
	})
	if err != nil {
		zap.L().Error("Failed to send message", zap.Error(err))
	}
}

func handler(ctx context.Context, b *botApi.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	switch update.Message.Text {
	case "/ping":
		_, err := b.SendMessage(ctx, &botApi.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "pong 🏓",
		})
		if err != nil {
			zap.L().Error("Failed to handle /ping command", zap.Error(err))
		}
	default:
		_, err := b.SendMessage(ctx, &botApi.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Komutu anlayamadım. 🤖",
		})
		if err != nil {
			zap.L().Error("Failed to handle unknown command", zap.Error(err))
		}
	}
}
