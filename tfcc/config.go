package tfcc

import (
	"github.com/spf13/viper"
	"path/filepath"
)

var (
	tfccConfig   = viper.New()
	bilibiliData = viper.New()
	permData     = viper.New()
)

func init() {
	tfccConfig.AddConfigPath(filepath.Join("config", "org.tfcc.bot"))
	tfccConfig.SetConfigName("TFCCConfig")
	tfccConfig.SetConfigType("yml")
	tfccConfig.SetDefault("bilibili.area_v2", 236)
	tfccConfig.SetDefault("bilibili.mid", 12345678)
	tfccConfig.SetDefault("bilibili.room_id", 12345678)
	tfccConfig.SetDefault("qq.super_admin_qq", int64(12345678))
	tfccConfig.SetDefault("qq.qq_group", []int64{12345678})
	_ = tfccConfig.SafeWriteConfigAs(filepath.Join("config", "org.tfcc.bot", "TFCCConfig.yml"))
	if err := tfccConfig.ReadInConfig(); err != nil {
		panic(err)
	}

	bilibiliData.AddConfigPath(filepath.Join("data", "org.tfcc.bot"))
	bilibiliData.SetConfigName("BilibiliData")
	bilibiliData.SetConfigType("yml")
	bilibiliData.SetDefault("cookies", []string(nil))
	bilibiliData.SetDefault("live", int64(0))
	_ = bilibiliData.SafeWriteConfigAs(filepath.Join("data", "org.tfcc.bot", "BilibiliData.yml"))
	if err := bilibiliData.ReadInConfig(); err != nil {
		panic(err)
	}

	permData.AddConfigPath(filepath.Join("data", "org.tfcc.bot"))
	permData.SetConfigName("PermData")
	permData.SetConfigType("yml")
	permData.SetDefault("admin", []int64(nil))
	permData.SetDefault("white_list", []int64(nil))
	_ = permData.SafeWriteConfigAs(filepath.Join("data", "org.tfcc.bot", "PermData.yml"))
	if err := permData.ReadInConfig(); err != nil {
		panic(err)
	}
}
