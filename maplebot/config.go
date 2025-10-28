package maplebot

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/CuteReimu/YinYangJade/maplebot/scripts"
	"github.com/spf13/viper"
)

var (
	config         = viper.New()
	qunDb          = viper.New()
	findRoleData   = viper.New()
	levelExpData   = viper.New()
	classImageData = viper.New()
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
	config.SetDefault("admin_groups", []int64{12345678})
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

	levelExpData.AddConfigPath(filepath.Join("data", "net.cutereimu.maplebots"))
	levelExpData.SetConfigName("LevelExpData")
	levelExpData.SetConfigType("yml")
	_ = levelExpData.SafeWriteConfigAs(filepath.Join("data", "net.cutereimu.maplebots", "LevelExpData.yml"))
	if err := levelExpData.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := scripts.BuildLvlData(levelExpData); err != nil {
		slog.Error("BuildLvlData失败", "error", err.Error())
	}

	classImageData.AddConfigPath(filepath.Join("data", "net.cutereimu.maplebots"))
	classImageData.SetConfigName("ClassImageData")
	classImageData.SetConfigType("yml")
	for k := range ClassNameMap {
		if len(k) > 0 {
			classImageData.SetDefault(k, "")
		}
	}
	_ = classImageData.SafeWriteConfigAs(filepath.Join("data", "net.cutereimu.maplebots", "ClassImageData.yml"))
	if err := classImageData.ReadInConfig(); err != nil {
		panic(err)
	}
}
