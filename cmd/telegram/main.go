// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"database/sql"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"uranus/internal/background"
	"uranus/internal/telegram"
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
	config.SetConfigName("telegram")
	config.SetConfigType("yaml")
	config.AddConfigPath("/etc/hackernel")
	if err := config.ReadInConfig(); err != nil {
		logrus.Fatal(err)
	}

	token := config.GetString("token")
	ownerID := config.GetInt64("id")
	dbFile := config.GetString("db")
	dbOptions := "?cache=shared&mode=rwc&_journal_mode=WAL"
	dataSourceName := dbFile + dbOptions

	os.MkdirAll(filepath.Dir(dataSourceName), os.ModeDir)
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	telegramWorker := telegram.NewWorker(token, ownerID)
	processWorker := background.NewProcessWorker(db)

	if err := telegram.SetStandaloneMode(db); err != nil {
		logrus.Fatal(err)
	}

	if err := processWorker.Init(); err != nil {
		logrus.Fatal(err)
	}

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
