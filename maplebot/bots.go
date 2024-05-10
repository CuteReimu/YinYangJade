package maplebot

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var restyClient = resty.New()

func init() {
	restyClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	restyClient.SetTimeout(20 * time.Second)
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69",
		"connection":   "close",
	})
}

var B *Bot

func Init(b *Bot) {
	initConfig()
	B = b
	B.ListenGroupMessage(handleGroupMessage)
}

var addDbQQList = make(map[int64]string)

func handleGroupMessage(message *GroupMessage) bool {
	if len(message.MessageChain) != 2 {
		return true
	}
	if !slices.Contains(config.GetIntSlice("qq_groups"), int(message.Sender.Group.Id)) {
		return true
	}
	plain, ok := message.MessageChain[1].(*Plain)
	if !ok {
		return true
	}
	perm := message.Sender.Permission == PermAdministrator || message.Sender.Permission == PermOwner
	if plain.Text == "ping" {
		sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "pong"})
	} else if plain.Text == "roll" {
		sendGroupMessage(message.Sender.Group.Id, &Plain{Text: message.Sender.MemberName + " roll: " + strconv.Itoa(rand.IntN(100))})
	} else if strings.HasPrefix(plain.Text, "roll ") {
		upperLimit, _ := strconv.Atoi(strings.TrimSpace(plain.Text[4:]))
		if upperLimit > 0 {
			sendGroupMessage(message.Sender.Group.Id, &Plain{Text: message.Sender.MemberName + " roll: " + strconv.Itoa(rand.IntN(upperLimit)+1)})
		}
	} else if perm && strings.HasPrefix(plain.Text, "添加词条 ") {
		key := dealKey(plain.Text[4:])
		if len(key) > 0 {
			if qunDb.IsSet("data." + key) {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "词条已存在"})
			} else {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "请输入要添加的内容"})
				addDbQQList[message.Sender.Id] = key
			}
		}
	} else if perm && strings.HasPrefix(plain.Text, "修改词条 ") {
		key := dealKey(plain.Text[4:])
		if len(key) > 0 {
			if !qunDb.IsSet("data." + key) {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "词条不存在"})
			} else {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "请输入要添加的内容"})
				addDbQQList[message.Sender.Id] = key
			}
		}
	} else if perm && strings.HasPrefix(plain.Text, "删除词条 ") {
		key := dealKey(plain.Text[4:])
		if len(key) > 0 {
			m := qunDb.GetStringMapString("data")
			if _, ok = m[key]; !ok {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "词条不存在"})
			} else {
				delete(m, key)
				qunDb.Set("data", m)
				if err := qunDb.WriteConfig(); err != nil {
					slog.Error("write data failed", "error", err)
				}
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "删除词条成功"})
			}
		}
	} else if strings.HasPrefix(plain.Text, "查询词条 ") || strings.HasPrefix(plain.Text, "搜索词条 ") {
		key := dealKey(plain.Text[4:])
		if len(key) > 0 {
			var res []string
			m := qunDb.GetStringMapString("data")
			for k := range m {
				if strings.Contains(k, key) {
					res = append(res, k)
				}
			}
			if len(res) > 0 {
				slices.Sort(res)
				num := len(res)
				if num > 10 {
					res = res[:10]
					res[9] += fmt.Sprintf("\n等%d个词条", num)
				}
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "搜索到以下词条：\n" + strings.Join(res, "\n")})
			} else {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "搜索不到词条(" + key + ")"})
			}
		}
	} else {
		if key, ok := addDbQQList[message.Sender.Id]; ok { // 添加词条
			var ms []SingleMessage
			for _, m := range message.MessageChain {
				if _, ok = m.(*Source); !ok {
					ms = append(ms, m)
				}
			}
			buf, err := json.Marshal(ms)
			if err != nil {
				slog.Error("json marshal failed", "error", err)
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "编辑词条失败"})
				return true
			}
			if err = saveImage(ms); err != nil {
				sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "编辑词条失败，" + err.Error()})
				return true
			}
			qunDb.Set("data."+key, string(buf))
			if err = qunDb.WriteConfig(); err != nil {
				slog.Error("write data failed", "error", err)
			}
			sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "编辑词条成功"})
		} else { // 调用词条
			qunDb.GetString("data." + dealKey(message.MessageChain[1].(*Plain).Text))
		}
	}
	return true
}

func saveImage(message []SingleMessage) error {
	for _, m := range message {
		if img, ok := m.(*Image); ok && len(img.Url) > 0 {
			resp, err := restyClient.R().Get(img.Url)
			if err != nil {
				slog.Error("get image failed", "error", err)
				return errors.New("保存图片失败")
			}
			if resp.StatusCode() != 200 {
				slog.Error("get image failed", "status", resp.Status(), "body", resp.String())
				return errors.New("保存图片失败")
			}
			if err = os.MkdirAll("chat-images", 0755); err != nil {
				slog.Error("mkdir failed", "error", err)
				return errors.New("保存图片失败")
			}
			if err = os.WriteFile(filepath.Join("dictionary-images", img.ImageId), resp.Body(), 0644); err != nil {
				slog.Error("write image failed", "error", err)
				return errors.New("保存图片失败")
			}
			img.Path = filepath.Join("..", "YinYangJade", "dictionary-images", img.ImageId)
			img.Url = ""
		}
	}
	return nil
}

func sendGroupMessage(group int64, messages ...SingleMessage) {
	_, err := B.SendGroupMessage(group, 0, messages)
	if err != nil {
		slog.Error("send group message failed", "error", err)
	}
}

func dealKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "零", "0")
	s = strings.ReplaceAll(s, "一", "1")
	s = strings.ReplaceAll(s, "二", "2")
	s = strings.ReplaceAll(s, "三", "3")
	s = strings.ReplaceAll(s, "四", "4")
	s = strings.ReplaceAll(s, "五", "5")
	s = strings.ReplaceAll(s, "六", "6")
	s = strings.ReplaceAll(s, "七", "7")
	s = strings.ReplaceAll(s, "八", "8")
	s = strings.ReplaceAll(s, "九", "9")
	return strings.ToLower(s)
}
