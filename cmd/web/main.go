package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/web"
	"uranus/pkg/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogrusFormat()

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	webWorker := web.NewWorker("0.0.0.0:80")
	if err := webWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	sig := <-sigchan
	logrus.Info(sig)

	if err := webWorker.Stop(); err != nil {
		logrus.Fatal(err)
	}
}
