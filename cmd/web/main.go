// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"database/sql"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"uranus/internal/background"
	"uranus/internal/web"
	"uranus/pkg/logger"

	_ "github.com/mattn/go-sqlite3"
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
	dbFile := config.GetString("db")
	dbOptions := "?cache=shared&mode=rwc&_journal_mode=WAL"
	dataSourceName := dbFile + dbOptions
	os.MkdirAll(filepath.Dir(dataSourceName), os.ModeDir)
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	processWorker := background.NewProcessWorker(db)
	fileWorker := background.NewFileWorker(db)
	netWorker := background.NewNetWorker(db)
	webWorker := web.NewWorker(listen, db)

	if err := processWorker.Init(); err != nil {
		logrus.Fatal(err)
	}

	if err := fileWorker.Init(); err != nil {
		logrus.Fatal(err)
	}

	if err := netWorker.Init(); err != nil {
		logrus.Fatal(err)
	}

	if err := webWorker.Init(); err != nil {
		logrus.Fatal(err)
	}

	if err := processWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	if err := fileWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

	if err := netWorker.Start(); err != nil {
		logrus.Fatal(err)
	}

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

	if err := netWorker.Stop(); err != nil {
		logrus.Error(err)
	}
}
