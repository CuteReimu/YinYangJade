package main

import (
	"fmt"
	"github.com/CuteReimu/YinYangJade/fengsheng"
	"github.com/CuteReimu/YinYangJade/maplebot"
	"github.com/CuteReimu/YinYangJade/tfcc"
	"github.com/CuteReimu/mirai-sdk-http"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var mainConfig = viper.New()

func init() {
	writerError, err := rotatelogs.New(
		path.Join("logs", "log-%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		slog.Error("unable to write logs", "error", err)
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(writerError, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format("15:04:05.000"))
				}
			case slog.SourceKey:
				if s, ok := a.Value.Any().(*slog.Source); ok {
					if index := strings.LastIndex(s.File, "@"); index >= 0 {
						if index += strings.Index(s.File[index:], string(filepath.Separator)); index >= 0 {
							s.File = s.File[index+1:]
						}
					}
				}
			default:
				if e, ok := a.Value.Any().(error); ok {
					a.Value = slog.StringValue(fmt.Sprintf("%+v", e))
				}
			}
			return a
		}})))

	mainConfig.SetConfigName("config")
	mainConfig.SetConfigType("yml")
	mainConfig.AddConfigPath(".")
	mainConfig.SetDefault("host", "localhost")
	mainConfig.SetDefault("port", 8080)
	mainConfig.SetDefault("verifyKey", "ABCDEFGHIJK")
	mainConfig.SetDefault("qq", 123456789)
	mainConfig.SetDefault("check_qq_groups", []int64(nil))
	if err := mainConfig.SafeWriteConfig(); err == nil {
		fmt.Println("Already generated config.yaml. Please modify the config file and restart the program.")
		os.Exit(0)
	}
	if err := mainConfig.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	var err error
	host := mainConfig.GetString("host")
	port := mainConfig.GetInt("port")
	verifyKey := mainConfig.GetString("verifyKey")
	qq := mainConfig.GetInt64("qq")
	b, err := miraihttp.Connect(host, port, miraihttp.WsChannelAll, verifyKey, qq, false)
	if err != nil {
		slog.Error("connect failed", "error", err)
		os.Exit(1)
	}
	tfcc.Init(b)
	fengsheng.Init(b)
	maplebot.Init(b)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
