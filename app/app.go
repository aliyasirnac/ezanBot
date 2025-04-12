package app

import (
	"context"
	"fmt"
	bot2 "github.com/aliyasirnac/ezanBot/bot"
	"github.com/aliyasirnac/ezanBot/config"
	"github.com/aliyasirnac/ezanBot/thirdparty"
	"go.uber.org/zap"
	"time"

	"github.com/robfig/cron/v3"
)

type App struct {
	Config *config.Config
	Cron   *cron.Cron
	Bot    *bot2.Bot
}

func New(config *config.Config) *App {
	return &App{Config: config}
}

func (a *App) Start(ctx context.Context) error {
	zap.L().Info("Bot starting")
	ezanService := thirdparty.NewEzan("https://ezanvakti.emushaf.net")
	bot := bot2.New(&a.Config.Telegram, &ezanService)
	a.Bot = bot

	errCh := make(chan error, 1)
	go func() {
		err := bot.Start(ctx)
		if err != nil {
			zap.L().Error("Error while starting bot", zap.Error(err))
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Second):
		zap.L().Info("Bot started successfully")
	}

	zap.L().Info("Cron starting")
	a.Cron = cron.New()

	_, err := a.Cron.AddFunc("0 0 1 * *", func() {
		zap.L().Info("Monthly fetch by cron job")
		a.sendEzanTimes(ctx, ezanService)
	})
	if err != nil {
		zap.L().Error("Failed to start monthly cronjob", zap.Error(err))
		return err
	}

	_, err = a.Cron.AddFunc("0 1 * * *", func() {
		zap.L().Info("Daily fetch by cron job")
		a.sendEzanTimes(ctx, ezanService)
	})

	//a.sendEzanTimes(ctx, ezanService)

	if err != nil {
		zap.L().Error("Failed to start daily cronjob", zap.Error(err))
		return err
	}

	a.Cron.Start()
	return nil
}

func (a *App) sendEzanTimes(ctx context.Context, ezanService thirdparty.ThirdParty[thirdparty.EzanRequest, thirdparty.EzanResponse]) {
	today := time.Now().Format("02.01.2006")
	zap.L().Info("Today", zap.String("today", today))
	res, err := ezanService.Handler(ctx, &thirdparty.EzanRequest{IlceId: 9541})
	if err != nil {
		zap.L().Error("Could not get ezan response", zap.Error(err))
		return
	}

	ezanData := a.filterEzanByDate(res, today)
	if ezanData == nil {
		zap.L().Error("No ezan data found for today")
		return
	}

	message := fmt.Sprintf(
		"📅 %s İstanbul için Ezan Saatleri\n\n🌅 İmsak: %s\n☀️ Güneş: %s\n🕌 Öğle: %s\n🕒 İkindi: %s\n🌇 Akşam: %s\n🌙 Yatsı: %s",
		ezanData.MiladiTarihKisa,
		ezanData.Imsak,
		ezanData.Gunes,
		ezanData.Ogle,
		ezanData.Ikindi,
		ezanData.Aksam,
		ezanData.Yatsi,
	)
	a.Bot.SendMessage(ctx, message)
}

func (a *App) filterEzanByDate(response *thirdparty.EzanResponse, date string) *thirdparty.EzanBody {
	for _, ezan := range response.EzanBody {
		if ezan.MiladiTarihKisa == date {
			return &ezan
		}
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	zap.L().Info("Stopping cron jobs...")
	a.Cron.Stop()

	select {
	case <-ctx.Done():
		zap.L().Warn("Shutdown took too long, forcing stop")
		return ctx.Err()
	default:
		zap.L().Info("Shutdown completed successfully")
		return nil
	}
}
