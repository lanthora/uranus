package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/lanthora/uranus/internal/notify"
	"github.com/lanthora/uranus/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logger.InitLogrusFormat()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	config := viper.New()
	config.SetConfigName("notify")
	config.SetConfigType("yaml")
	config.AddConfigPath("$HOME/.uranus")

	if err := config.ReadInConfig(); err != nil {
		logrus.Fatal(err)
	}

	server := config.GetString("server")
	username := config.GetString("username")
	password := config.GetString("password")
	processEventOffset := config.GetInt64("process-event-offset")

	notifier := notify.NewWorker(server, username, password, processEventOffset)
	notifier.Start()

	sig := <-sigchan
	logrus.Info(sig)

	notifier.Stop()

	// 每次正常退出的时候记录已经获取过通知的偏移量,避免重启后重复获取
	config.Set("process-event-offset", notifier.ProcessEventOffset)
	if err := config.WriteConfig(); err != nil {
		logrus.Error(err)
	}
}
