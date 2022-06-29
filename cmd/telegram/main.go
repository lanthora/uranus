// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/background"
	"uranus/internal/telegram"
	"uranus/pkg/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger.InitLogrusFormat()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	config := viper.New()
	config.SetConfigName("telegram")
	config.SetConfigType("yaml")
	config.AddConfigPath("/etc/hackernel")
	if err := config.ReadInConfig(); err != nil {
		logrus.Fatal(err)
	}

	token := config.GetString("token")
	ownerID := config.GetInt64("id")
	dataSourceName := config.GetString("dsn")

	telegramWorker := telegram.NewWorker(token, ownerID)
	processWorker := background.NewProcessWorker(dataSourceName)

	if err := telegramWorker.Start(); err != nil {
		logrus.Fatal(err)
	}
	if err := processWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	sig := <-sigchan
	logrus.Info(sig)

	telegramWorker.Stop()
	processWorker.Stop()
}
