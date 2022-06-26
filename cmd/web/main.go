// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/background"
	"uranus/internal/web"
	"uranus/pkg/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger.InitLogrusFormat()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	config := viper.New()
	config.SetConfigName("web")
	config.SetConfigType("yaml")
	config.AddConfigPath("/etc/hackernel")
	if err := config.ReadInConfig(); err != nil {
		logrus.Fatal(err)
	}

	listen := config.GetString("listen")
	dbName := config.GetString("db")

	processWorker := background.NewProcessWorker(dbName)
	if err := processWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	fileWorker := background.NewFileWorker(dbName)
	if err := fileWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	webWorker := web.NewWorker(listen, dbName)
	if err := webWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("listen: ", listen)

	sig := <-sigchan
	logrus.Info(sig)

	if err := webWorker.Stop(); err != nil {
		logrus.Error(err)
	}

	if err := processWorker.Stop(); err != nil {
		logrus.Error(err)
	}

	if err := fileWorker.Stop(); err != nil {
		logrus.Error(err)
	}
}
