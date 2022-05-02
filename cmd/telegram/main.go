package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/judge"
	"uranus/internal/telegram"
	"uranus/pkg/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger.InitLogrusFormat()

	config := viper.New()
	config.SetConfigName("telegram")
	config.SetConfigType("yaml")
	config.AddConfigPath("/etc/hackernel")
	config.ReadInConfig()
	token := config.GetString("token")
	ownerID := config.GetInt64("id")
	dbName := config.GetString("db")

	telegramWorker := telegram.NewWorker(token, ownerID)
	judgeWorker := judge.NewProcessWorker(dbName)

	if err := telegramWorker.Start(); err != nil {
		logrus.Fatal(err)
	}
	if err := judgeWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigchan
	logrus.Info(sig)

	telegramWorker.Stop()
	judgeWorker.Stop()
}
