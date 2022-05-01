package main

import (
	"os"
	"os/signal"
	"syscall"
	"uranus/internal/sample"
	"uranus/pkg/logger"
	"uranus/pkg/worker"

	"github.com/sirupsen/logrus"
)

func main() {
	var scheduler worker.Scheduler
	var sampleWorker sample.SampleWorker

	logger.InitLogrusFormat()

	scheduler.Init()
	scheduler.Register("sample", &sampleWorker)
	scheduler.StartWorker()

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigchan
	logrus.Debug(sig.String())

	scheduler.StopWorker()
	scheduler.Unregister("sample")
}
