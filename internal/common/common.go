package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetDataSourceNameFromConfig(config *viper.Viper) (dataSourceName string) {
	dbFile := strings.TrimPrefix(config.GetString("db"), "file:")

	if dbFile == ":memory:" {
		logrus.Info("Currently using an in-memory database, data will be lost when the process exits")
		dataSourceName = "file:" + dbFile
		return
	}

	if strings.ContainsAny(dbFile, "?=&") {
		logrus.Fatal("Path contains invalid characters")
		return
	}

	os.MkdirAll(filepath.Dir(dbFile), os.ModeDir)
	dbOptions := "?cache=shared&mode=rwc&_journal_mode=WAL"
	dataSourceName = "file:" + dbFile + dbOptions
	return
}
