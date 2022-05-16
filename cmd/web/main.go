package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/judge"
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
	dbName := config.GetString("db")
	judgeWorker := judge.NewProcessWorker(dbName)
	if err := judgeWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	listen := config.GetString("listen")
	webWorker := web.NewWorker(listen)
	if err := webWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	sig := <-sigchan
	logrus.Info(sig)

	if err := webWorker.Stop(); err != nil {
		logrus.Error(err)
	}
	if err := judgeWorker.Stop(); err != nil {
		logrus.Error(err)
	}
}
