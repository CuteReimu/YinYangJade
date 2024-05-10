package fengsheng

import (
	"encoding/json"
	"fmt"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

	checkQunDb()
}

func checkQunDb() {
	m := qunDb.GetStringMapString("data")
loop:
	for k, v := range m {
		if strings.Contains(k, ".") {
			panic("词条名称中不能包含 . 符号: " + k)
		}
		var a []any
		err := json.Unmarshal([]byte(v), &a)
		if err != nil {
			panic(err)
		}
		for i, b0 := range a {
			b := b0.(map[string]any)
			switch b["type"].(string) {
			case "Image":
				if b["path"] != nil && len(b["path"].(string)) > 0 {
					continue
				}
				imageId, ok := b["imageId"].(string)
				if !ok {
					panic("imageId 错误: " + k)
				}
				a[i] = &Image{
					Type: "Image",
					Path: filepath.Join("..", "YinYangJade", "chat-images", imageId),
				}
			case "Face":
				if b["faceId"] != nil && b["faceId"].(float64) != 0 {
					continue
				}
				faceId, ok := b["id"].(float64)
				if !ok {
					panic("faceId 错误: " + k)
				}
				a[i] = &Face{
					Type:   "Face",
					FaceId: int32(faceId),
				}
			case "PlainText":
				content, ok := b["content"].(string)
				if !ok {
					panic("content 错误: " + k)
				}
				a[i] = &Plain{
					Type: "Plain",
					Text: content,
				}
			case "Plain":
			default:
				fmt.Println("Unknown type: ", b["type"], ", in: ", k)
				continue loop
			}
		}
		buf, err := json.Marshal(a)
		if err != nil {
			panic(err)
		}
		m[k] = string(buf)
	}
	qunDb.Set("data", m)
	if err := qunDb.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
}
