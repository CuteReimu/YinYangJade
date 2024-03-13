package main

import (
	"fmt"
	"github.com/CuteReimu/YinYangJade/bots"
	"github.com/CuteReimu/YinYangJade/db"
	"github.com/CuteReimu/mirai-sdk-http/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	utils.InitLogger("./logs", slog.LevelDebug)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 8080)
	viper.SetDefault("verifyKey", "ABCDEFGHIJK")
	viper.SetDefault("qq", 123456789)
	viper.SetDefault("qq.super_admin_qq", 987654321)
	viper.SetDefault("tfcc.qq_groups", []int64{123456789})
	viper.SetDefault("tfcc.bilibili.mid", 123456789)
	viper.SetDefault("tfcc.bilibili.room_id", 123456789)
	viper.SetDefault("tfcc.bilibili.area_v2", 624)
	if err := viper.SafeWriteConfig(); err != nil {
		fmt.Println("Already generated config.yaml. Please modify the config file and restart the program.")
		os.Exit(0)
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	db.Init()
	defer db.Stop()
	bots.Init()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
