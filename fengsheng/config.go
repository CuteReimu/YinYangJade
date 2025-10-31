package fengsheng

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/vicanso/go-charts/v2"
)

var (
	fengshengConfig = viper.New()
	permData        = viper.New()
	qunDb           = viper.New()
)

func initConfig() {
	if err := os.MkdirAll(filepath.Join("config", "com.fengsheng.bot"), 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join("data", "com.fengsheng.bot"), 0755); err != nil {
		panic(err)
	}

	fengshengConfig.AddConfigPath(filepath.Join("config", "com.fengsheng.bot"))
	fengshengConfig.SetConfigName("FengshengConfig")
	fengshengConfig.SetConfigType("yml")
	fengshengConfig.SetDefault("qq.super_admin_qq", int64(12345678))
	fengshengConfig.SetDefault("qq.qq_group", []int64{12345678})
	fengshengConfig.SetDefault("fengshengUrl", "http://127.0.0.1:9094")
	fengshengConfig.SetDefault("fengshengPageUrl", "http://127.0.0.1:12221")
	fengshengConfig.SetDefault("image_expire_hours", int64(24))
	_ = fengshengConfig.SafeWriteConfigAs(filepath.Join("config", "com.fengsheng.bot", "FengshengConfig.yml"))
	if err := fengshengConfig.ReadInConfig(); err != nil {
		panic(err)
	}

	permData.AddConfigPath(filepath.Join("data", "com.fengsheng.bot"))
	permData.SetConfigName("PermData")
	permData.SetConfigType("yml")
	_ = permData.SafeWriteConfigAs(filepath.Join("data", "com.fengsheng.bot", "PermData.yml"))
	if err := permData.ReadInConfig(); err != nil {
		panic(err)
	}

	qunDb.AddConfigPath(filepath.Join("data", "com.fengsheng.bot"))
	qunDb.SetConfigName("QunDb")
	qunDb.SetConfigType("yml")
	_ = qunDb.SafeWriteConfigAs(filepath.Join("data", "com.fengsheng.bot", "QunDb.yml"))
	if err := qunDb.ReadInConfig(); err != nil {
		panic(err)
	}
}

func init() {
	var ok bool
	_ = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if info != nil && info.Name() == "simhei.ttf" {
			buf, err := os.ReadFile(path)
			if err != nil {
				slog.Error("读取字体文件失败", "error", err)
				return err
			}
			err = charts.InstallFont("simhei", buf)
			if err != nil {
				slog.Error("安装字体失败", "error", err)
				return err
			}
			ok = true
		}
		return nil
	})
	if ok {
		return
	}
	_ = filepath.Walk(filepath.Join("/usr", "share", "fonts"), func(path string, info fs.FileInfo, err error) error {
		if info != nil && info.Name() == "simhei.ttf" {
			buf, err := os.ReadFile(path)
			if err != nil {
				slog.Error("读取字体文件失败", "error", err)
				return err
			}
			err = charts.InstallFont("simhei", buf)
			if err != nil {
				slog.Error("安装字体失败", "error", err)
				return err
			}
		}
		return nil
	})
}
