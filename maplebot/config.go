package maplebot

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var (
	config       = viper.New()
	qunDb        = viper.New()
	findRoleData = viper.New()
)

func initConfig() {
	if err := os.MkdirAll(filepath.Join("config", "net.cutereimu.maplebots"), 0755); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Join("data", "net.cutereimu.maplebots"), 0755); err != nil {
		panic(err)
	}

	config.AddConfigPath(filepath.Join("config", "net.cutereimu.maplebots"))
	config.SetConfigName("Config")
	config.SetConfigType("yml")
	config.SetDefault("admin", int64(12345678))
	config.SetDefault("qq_groups", []int64{12345678})
	config.SetDefault("image_expire_hours", int64(24))
	_ = config.SafeWriteConfigAs(filepath.Join("config", "net.cutereimu.maplebots", "Config.yml"))
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}

	qunDb.AddConfigPath(filepath.Join("data", "net.cutereimu.maplebots"))
	qunDb.SetConfigName("QunDb")
	qunDb.SetConfigType("yml")
	_ = qunDb.SafeWriteConfigAs(filepath.Join("data", "net.cutereimu.maplebots", "QunDb.yml"))
	if err := qunDb.ReadInConfig(); err != nil {
		panic(err)
	}

	findRoleData.AddConfigPath(filepath.Join("data", "net.cutereimu.maplebots"))
	findRoleData.SetConfigName("FindRoleData")
	findRoleData.SetConfigType("yml")
	_ = findRoleData.SafeWriteConfigAs(filepath.Join("data", "net.cutereimu.maplebots", "FindRoleData.yml"))
	if err := findRoleData.ReadInConfig(); err != nil {
		panic(err)
	}
}
