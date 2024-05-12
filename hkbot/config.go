package hkbot

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var (
	hkConfig = viper.New()
	hkData   = viper.New()
	qunDb    = viper.New()
)

func initConfig() {
	if err := os.MkdirAll(filepath.Join("config", "net.cutereimu.hkbot"), 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join("data", "net.cutereimu.hkbot"), 0755); err != nil {
		panic(err)
	}

	hkConfig.AddConfigPath(filepath.Join("config", "net.cutereimu.hkbot"))
	hkConfig.SetConfigName("HKConfig")
	hkConfig.SetConfigType("yml")
	hkConfig.SetDefault("enable", true)
	hkConfig.SetDefault("speedrun_push_delay", int64(300))
	hkConfig.SetDefault("speedrun_push_qq_group", []int64{12345678})
	hkConfig.SetDefault("speedrun_api_key", "abcdefjhijk")
	_ = hkConfig.SafeWriteConfigAs(filepath.Join("config", "net.cutereimu.hkbot", "HKConfig.yml"))
	if err := hkConfig.ReadInConfig(); err != nil {
		panic(err)
	}

	hkData.AddConfigPath(filepath.Join("data", "net.cutereimu.hkbot"))
	hkData.SetConfigName("HKData")
	hkData.SetConfigType("yml")
	_ = hkData.SafeWriteConfigAs(filepath.Join("data", "net.cutereimu.hkbot", "HKData.yml"))
	if err := hkData.ReadInConfig(); err != nil {
		panic(err)
	}
}
