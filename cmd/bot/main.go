package main

import (
	"context"
	"github.com/aliyasirnac/ezanBot/app"
	"github.com/aliyasirnac/ezanBot/config"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	parentCtx := context.Background()
	closeChan := make(chan os.Signal, 2)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)

	// logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Fatal("Failed to sync logger", zap.Error(err))
		}
	}(logger)
	zap.ReplaceGlobals(logger)

	cfg, err := config.LoadConfig(parentCtx)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	a := app.New(cfg)
	go func() {
		if err := a.Start(parentCtx); err != nil {
			logger.Fatal("Failed to start app", zap.Error(err))
		}
	}()

	sig := <-closeChan
	logger.Info("Caught signal %s: shutting down.", zap.String("signal", sig.String()))

	if err := a.Stop(parentCtx); err != nil {
		logger.Fatal("Failed to stop app", zap.Error(err))
	}
}
