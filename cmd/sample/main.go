// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/sample"
	"uranus/pkg/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogrusFormat()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	sampleWorker := sample.NewWorker()
	sampleWorker.Start()

	sig := <-sigchan
	logrus.Info(sig)

	sampleWorker.Stop()
}
