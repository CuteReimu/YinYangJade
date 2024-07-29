package maplebot

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CuteReimu/YinYangJade/slicegame"
	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"os"
	"os/exec"
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
	go func() {
		B.Run(clearExpiredImages)
		B.Run(clearExpiredImages2)
		for range time.Tick(24 * time.Hour) {
			B.Run(clearExpiredImages)
			B.Run(clearExpiredImages2)
		}
	}()
	B.ListenGroupMessage(handleGroupMessage)
}

var (
	addDbQQList          = make(map[int64]string)
	updateClassImageList = make(map[int64]string)
)

func handleGroupMessage(message *GroupMessage) bool {
	if len(message.Message) <= 0 {
		return true
	}
	if !slices.Contains(config.GetIntSlice("qq_groups"), int(message.GroupId)) {
		return true
	}
	if len(message.Message) >= 2 {
		if text, ok := message.Message[0].(*Text); ok && strings.TrimSpace(text.Text) == "查询" {
			if at, ok := message.Message[1].(*At); ok {
				data := findRoleData.GetStringMapString("data")
				name := data[at.QQ]
				if len(name) == 0 {
					sendGroupMessage(message, &Text{Text: "该玩家还未绑定"})
				} else {
					go func() {
						defer func() {
							if err := recover(); err != nil {
								slog.Error("panic recovered", "error", err)
							}
						}()
						sendGroupMessage(message, findRole(name)...)
					}()
				}
			}
			return true
		}
	}
	if len(message.Message) == 1 {
		if text, ok := message.Message[0].(*Text); ok {
			perm := message.Sender.Role == RoleAdmin || message.Sender.Role == RoleOwner ||
				config.GetInt64("admin") == message.Sender.UserId
			if perm {
				perm = slices.Contains(config.GetIntSlice("admin_groups"), int(message.GroupId))
			}
			if text.Text == "ping" {
				sendGroupMessage(message, &Text{Text: "pong"})
				return true
			} else if text.Text == "roll" {
				replyGroupMessage(true, message, &Text{Text: "roll: " + strconv.Itoa(rand.IntN(100))})
				return true
			} else if strings.HasPrefix(text.Text, "roll ") {
				upperLimit, _ := strconv.Atoi(strings.TrimSpace(text.Text[len("roll"):]))
				if upperLimit > 0 {
					replyGroupMessage(true, message, &Text{Text: "roll: " + strconv.Itoa(rand.IntN(upperLimit)+1)})
				}
				return true
			} else if text.Text == "滑块" {
				sendGroupMessage(message, slicegame.DoStuff()...)
				return true
			} else if text.Text == "8421" {
				sendGroupMessage(message, calculatePotion()...)
				return true
			} else if text.Text == "升级经验" {
				sendGroupMessage(message, calculateLevelExp()...)
				return true
			} else if strings.HasPrefix(text.Text, "等级压制 ") {
				content := strings.TrimSpace(text.Text[len("等级压制"):])
				sendGroupMessage(message, calculateExpDamage(content)...)
				return true
			} else if strings.HasPrefix(text.Text, "生成表格 ") {
				content := strings.TrimSpace(text.Text[len("升级经验"):])
				sendGroupMessage(message, genTable(content)...)
				return true
			} else if strings.HasPrefix(text.Text, "升级经验 ") {
				content := strings.TrimSpace(text.Text[len("升级经验"):])
				parts := strings.SplitN(content, " ", 2)
				if len(parts) == 2 {
					start, _ := strconv.Atoi(parts[0])
					end, _ := strconv.Atoi(parts[1])
					sendGroupMessage(message, calculateExpBetweenLevel(int64(start), int64(end))...)
				}
				return true
			} else if text.Text == "爆炸次数" || strings.HasPrefix(text.Text, "爆炸次数 ") {
				sendGroupMessage(message, calculateBoomCount(text.Text[len("爆炸次数"):])...)
				return true
			} else if text.Text == "查询我" {
				data := findRoleData.GetStringMapString("data")
				name := data[strconv.FormatInt(message.Sender.UserId, 10)]
				if len(name) == 0 {
					sendGroupMessage(message, &Text{Text: "你还未绑定"})
				} else {
					go func() {
						defer func() {
							if err := recover(); err != nil {
								slog.Error("panic recovered", "error", err)
							}
						}()
						sendGroupMessage(message, findRole(name)...)
					}()
				}
				return true
			} else if strings.HasPrefix(text.Text, "查询 ") {
				name := strings.TrimSpace(text.Text[len("查询"):])
				parts := strings.Split(name, " ")
				if len(parts) <= 1 {
					if !slices.ContainsFunc([]byte(name), func(b byte) bool { return (b < '0' || b > '9') && (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') }) {
						go func() {
							defer func() {
								if err := recover(); err != nil {
									slog.Error("panic recovered", "error", err)
								}
							}()
							sendGroupMessage(message, findRole(name)...)
						}()
					}
				} else {
					name1, name2 := parts[0], parts[1]
					if !slices.ContainsFunc([]byte(name1), func(b byte) bool { return (b < '0' || b > '9') && (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') }) &&
						!slices.ContainsFunc([]byte(name2), func(b byte) bool { return (b < '0' || b > '9') && (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') }) {
						sendGroupMessage(message, &Text{Text: "查询两个角色功能暂不可用"})
						//go func() {
						//	defer func() {
						//		if err := recover(); err != nil {
						//			slog.Error("panic recovered", "error", err, "stack", debug.Stack())
						//		}
						//	}()
						//	sendGroupMessage(message, findRole2(name1, name2)...)
						//}()
					}
				}
				return true
			} else if strings.HasPrefix(text.Text, "绑定 ") {
				data := findRoleData.GetStringMapString("data")
				if data[strconv.FormatInt(message.Sender.UserId, 10)] != "" {
					sendGroupMessage(message, &Text{Text: "你已经绑定过了，如需更换请先解绑"})
				} else {
					name := strings.TrimSpace(text.Text[len("绑定"):])
					if !slices.ContainsFunc([]byte(name), func(b byte) bool { return (b < '0' || b > '9') && (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') }) {
						data[strconv.FormatInt(message.Sender.UserId, 10)] = name
						findRoleData.Set("data", data)
						if err := findRoleData.WriteConfig(); err != nil {
							slog.Error("write config failed", "error", err)
						}
						sendGroupMessage(message, &Text{Text: "绑定成功"})
					}
				}
				return true
			} else if text.Text == "解绑" {
				data := findRoleData.GetStringMapString("data")
				if data[strconv.FormatInt(message.Sender.UserId, 10)] != "" {
					delete(data, strconv.FormatInt(message.Sender.UserId, 10))
					findRoleData.Set("data", data)
					if err := findRoleData.WriteConfig(); err != nil {
						slog.Error("write config failed", "error", err)
					}
					sendGroupMessage(message, &Text{Text: "解绑成功"})
				} else {
					sendGroupMessage(message, &Text{Text: "你还未绑定"})
				}
				return true
			} else if text.Text == "升星性价比" {
				sendGroupMessage(message, calStarForceCostPerformance()...)
			} else if strings.HasPrefix(text.Text, "模拟升星 ") || strings.HasPrefix(text.Text, "模拟上星 ") ||
				strings.HasPrefix(text.Text, "升星期望 ") || strings.HasPrefix(text.Text, "上星期望 ") {
				content := strings.TrimSpace(text.Text[len("模拟升星"):])
				result1 := calculateStarForce1(content)
				if len(result1) > 0 {
					sendGroupMessage(message, result1...)
				} else if itemLevel, err := strconv.Atoi(content); err == nil {
					sendGroupMessage(message, calculateStarForce2(itemLevel, false, false)...)
				}
				return true
			} else if strings.HasPrefix(text.Text, "模拟升星必成活动 ") || strings.HasPrefix(text.Text, "模拟上星必成活动 ") ||
				strings.HasPrefix(text.Text, "升星期望必成活动 ") || strings.HasPrefix(text.Text, "上星期望必成活动 ") {
				content := strings.TrimSpace(text.Text[len("模拟升星必成活动"):])
				if itemLevel, err := strconv.Atoi(content); err == nil {
					sendGroupMessage(message, calculateStarForce2(itemLevel, false, true)...)
				}
				return true
			} else if strings.HasPrefix(text.Text, "模拟升星七折活动 ") || strings.HasPrefix(text.Text, "模拟上星七折活动 ") ||
				strings.HasPrefix(text.Text, "升星期望七折活动 ") || strings.HasPrefix(text.Text, "上星期望七折活动 ") {
				content := strings.TrimSpace(text.Text[len("模拟升星七折活动"):])
				if itemLevel, err := strconv.Atoi(content); err == nil {
					sendGroupMessage(message, calculateStarForce2(itemLevel, true, false)...)
				}
				return true
			} else if strings.HasPrefix(text.Text, "模拟升星超必活动 ") || strings.HasPrefix(text.Text, "模拟升星超级必成 ") ||
				strings.HasPrefix(text.Text, "模拟上星超必活动 ") || strings.HasPrefix(text.Text, "模拟上星超必活动 ") ||
				strings.HasPrefix(text.Text, "升星期望超必活动 ") || strings.HasPrefix(text.Text, "升星期望超级必成 ") ||
				strings.HasPrefix(text.Text, "上星期望超必活动 ") || strings.HasPrefix(text.Text, "上星期望超级必成 ") {
				content := strings.TrimSpace(text.Text[len("模拟升星超必活动"):])
				if itemLevel, err := strconv.Atoi(content); err == nil {
					sendGroupMessage(message, calculateStarForce2(itemLevel, true, true)...)
				}
				return true
			} else if strings.HasPrefix(text.Text, "模拟升星超级必成活动 ") || strings.HasPrefix(text.Text, "模拟上星超级必成活动 ") ||
				strings.HasPrefix(text.Text, "升星期望超级必成活动 ") || strings.HasPrefix(text.Text, "上星期望超级必成活动 ") {
				content := strings.TrimSpace(text.Text[len("模拟升星超级必成活动"):])
				if itemLevel, err := strconv.Atoi(content); err == nil {
					sendGroupMessage(message, calculateStarForce2(itemLevel, true, true)...)
				}
				return true
			} else if text.Text == "洗魔方" {
				sendGroupMessage(message, calculateCubeAll()...)
				return true
			} else if strings.HasPrefix(text.Text, "洗魔方 ") {
				sendGroupMessage(message, calculateCube(strings.TrimSpace(text.Text[len("洗魔方"):]))...)
				return true
			} else if perm && strings.HasPrefix(text.Text, "修改职业图片 ") {
				key := strings.TrimSpace(text.Text[len("修改职业图片"):])
				if len(key) > 0 {
					if _, ok := ClassNameMap[key]; !ok {
						sendGroupMessage(message, &Text{Text: "不存在的职业"})
					} else {
						sendGroupMessage(message, &Text{Text: "请输入要修改的内容"})
						updateClassImageList[message.Sender.UserId] = key
					}
				}
				return true
			} else if strings.HasPrefix(text.Text, "查询职业图片 ") {
				key := strings.TrimSpace(text.Text[len("查询职业图片"):])
				if len(key) > 0 {
					if img, err := GetClassOriginImageBuff(key); err != nil {
						slog.Error("找不到职业图片", "error", err)
						sendGroupMessage(message, &Text{Text: "找不到职业图片"})
					} else {
						sendGroupMessage(message, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(img)})
					}
				}
				return true
			} else if perm && strings.HasPrefix(text.Text, "添加词条 ") {
				key := dealKey(text.Text[len("添加词条"):])
				if strings.Contains(key, ".") {
					sendGroupMessage(message, &Text{Text: "词条名称中不能包含 . 符号"})
					return true
				}
				if len(key) > 0 {
					m := qunDb.GetStringMapString("data")
					if _, ok = m[key]; ok {
						sendGroupMessage(message, &Text{Text: "词条已存在"})
					} else {
						sendGroupMessage(message, &Text{Text: "请输入要添加的内容"})
						addDbQQList[message.Sender.UserId] = key
					}
				}
				return true
			} else if perm && strings.HasPrefix(text.Text, "修改词条 ") {
				key := dealKey(text.Text[len("修改词条"):])
				if len(key) > 0 {
					m := qunDb.GetStringMapString("data")
					if _, ok = m[key]; !ok {
						sendGroupMessage(message, &Text{Text: "词条不存在"})
					} else {
						sendGroupMessage(message, &Text{Text: "请输入要修改的内容"})
						addDbQQList[message.Sender.UserId] = key
					}
				}
				return true
			} else if perm && strings.HasPrefix(text.Text, "删除词条 ") {
				key := dealKey(text.Text[len("删除词条"):])
				if len(key) > 0 {
					m := qunDb.GetStringMapString("data")
					if _, ok = m[key]; !ok {
						sendGroupMessage(message, &Text{Text: "词条不存在"})
					} else {
						delete(m, key)
						qunDb.Set("data", m)
						if err := qunDb.WriteConfig(); err != nil {
							slog.Error("write data failed", "error", err)
						}
						sendGroupMessage(message, &Text{Text: "删除词条成功"})
					}
				}
				return true
			} else if strings.HasPrefix(text.Text, "查询词条 ") || strings.HasPrefix(text.Text, "搜索词条 ") {
				key := dealKey(text.Text[len("搜索词条"):])
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
						for i := range res {
							res[i] = fmt.Sprintf("%d. %s", i+1, res[i])
						}
						sendGroupMessage(message, &Text{Text: "搜索到以下词条：\n" + strings.Join(res, "\n")})
					} else {
						sendGroupMessage(message, &Text{Text: "搜索不到词条(" + key + ")"})
					}
				}
				return true
			}
		}
	}
	if key, ok := updateClassImageList[message.Sender.UserId]; ok { // 修改职业图片
		delete(updateClassImageList, message.Sender.UserId)
		if len(message.Message) != 1 {
			sendGroupMessage(message, &Text{Text: "提供的不是一张图片，修改失败"})
		} else if img, ok := message.Message[0].(*Image); !ok {
			sendGroupMessage(message, &Text{Text: "提供的不是一张图片，修改失败"})
		} else if len(img.Url) == 0 {
			sendGroupMessage(message, &Text{Text: "无法识别图片，修改失败"})
		} else {
			sendGroupMessage(message, SetClassImage(key, img)...)
		}
		return true
	} else if key, ok = addDbQQList[message.Sender.UserId]; ok { // 添加词条
		delete(addDbQQList, message.Sender.UserId)
		if err := saveImage(message.Message); err != nil {
			sendGroupMessage(message, &Text{Text: "编辑词条失败，" + err.Error()})
			return true
		}
		buf, err := json.Marshal(&message.Message)
		if err != nil {
			slog.Error("json marshal failed", "error", err)
			sendGroupMessage(message, &Text{Text: "编辑词条失败"})
			return true
		}
		m := qunDb.GetStringMapString("data")
		m[key] = string(buf)
		qunDb.Set("data", m)
		if err = qunDb.WriteConfig(); err != nil {
			slog.Error("write data failed", "error", err)
		}
		sendGroupMessage(message, &Text{Text: "编辑词条成功"})
	} else { // 调用词条
		if len(message.Message) == 1 {
			if text, ok := message.Message[0].(*Text); ok {
				m := qunDb.GetStringMapString("data")
				s := m[dealKey(text.Text)]
				if len(s) > 0 {
					var ms MessageChain
					if err := json.Unmarshal([]byte(s), &ms); err != nil {
						slog.Error("json unmarshal failed", "error", err, "s", s)
						sendGroupMessage(message, &Text{Text: "调用词条失败"})
						return true
					}
					sendGroupMessage(message, ms...)
				}
			}
		}
	}
	return true
}

func saveImage(message MessageChain) error {
	for _, m := range message {
		if img, ok := m.(*Image); ok && len(img.Url) > 0 {
			u := img.Url
			_, err := url.Parse(u)
			if err != nil {
				slog.Error("userInput is not a valid URL, reject it", "error", err)
				return err
			}
			if err := os.MkdirAll("chat-images", 0755); err != nil {
				slog.Error("mkdir failed", "error", err)
				return errors.New("保存图片失败")
			}
			p := filepath.Join("chat-images", img.File)
			abs, err := filepath.Abs(p)
			if err != nil {
				slog.Error("filepath.Abs() failed", "error", err)
				return errors.New("保存图片失败")
			}
			cmd := exec.Command("curl", "-o", p, u)
			if out, err := cmd.CombinedOutput(); err != nil {
				slog.Error("cmd.Run() failed", "error", err)
				return errors.New("保存图片失败")
			} else {
				slog.Debug(string(out))
			}
			img.File = "file://" + abs
			img.Url = ""
		}
	}
	return nil
}

func sendGroupMessage(context *GroupMessage, messages ...SingleMessage) {
	replyGroupMessage(false, context, messages...)
}

func replyGroupMessage(reply bool, context *GroupMessage, messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		for _, m := range messages {
			if img, ok := m.(*Image); ok && len(img.File) > 0 {
				if strings.HasPrefix(img.File, "file://") {
					fileName := img.File[len("file://"):]
					buf, err := os.ReadFile(fileName)
					if err != nil {
						slog.Error("read file failed", "error", err)
						continue
					}
					img.File = "base64://" + base64.StdEncoding.EncodeToString(buf)
				}
			}
		}
		if !reply {
			_, err := B.SendGroupMessage(context.GroupId, messages)
			return err
		}
		return context.Reply(B, messages, false)
	}
	if err := f(messages); err != nil {
		slog.Error("send group message failed", "error", err)
		newMessages := make([]SingleMessage, 0, len(messages))
		for _, m := range messages {
			if image, ok := m.(*Image); !ok || !strings.HasPrefix(image.File, "http") {
				newMessages = append(newMessages, m)
			}
		}
		if len(newMessages) != len(messages) && len(newMessages) > 0 {
			if err = f(newMessages); err != nil {
				slog.Error("send group message failed", "error", err)
			}
		}
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

func clearExpiredImages() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered", "error", err)
		}
	}()
	data := qunDb.GetStringMapString("data")
	data2 := make(map[string]bool)
	for _, v := range data {
		var ms MessageChain
		if err := json.Unmarshal([]byte(v), &ms); err != nil {
			slog.Error("json unmarshal failed", "error", err)
			continue
		}
		for _, m := range ms {
			if img, ok := m.(*Image); ok && len(img.File) > 0 && strings.HasPrefix(img.File, "file://") {
				data2[filepath.Base(img.File[len("file://"):])] = true
			}
		}
	}
	files, err := os.ReadDir("chat-images")
	if err != nil {
		slog.Error("read dir failed", "error", err)
	}
	for _, file := range files {
		if !data2[file.Name()] {
			if err = os.Remove(filepath.Join("chat-images", file.Name())); err != nil {
				slog.Error("remove file failed", "error", err)
			}
		}
	}
}
