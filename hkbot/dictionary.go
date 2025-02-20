package hkbot

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/CuteReimu/onebot"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

var addDbQQList = make(map[int64]string)

func handleDictionary(message *GroupMessage) bool {
	if len(message.Message) == 0 {
		return true
	}
	if !slices.Contains(hkConfig.GetIntSlice("speedrun_push_qq_group"), int(message.GroupId)) {
		return true
	}
	if len(message.Message) == 1 {
		if text, ok := message.Message[0].(*Text); ok {
			perm := IsWhitelist(message.Sender.UserId)
			if perm && strings.HasPrefix(text.Text, "添加词条 ") {
				key := dealKey(text.Text[len("添加词条"):])
				if strings.Contains(key, ".") {
					sendGroupMessage(message, &Text{Text: "词条名称中不能包含 . 符号"})
					return true
				}
				if _, ok = cmdMap[key]; ok {
					sendGroupMessage(message, &Text{Text: "不能用" + key + "作为词条"})
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
	if key, ok := addDbQQList[message.Sender.UserId]; ok { // 添加词条
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
			if err := os.MkdirAll("hk-images", 0755); err != nil {
				slog.Error("mkdir failed", "error", err)
				return errors.New("保存图片失败")
			}
			nameLen := len(filepath.Ext(img.File)) + 32
			if len(img.File) > nameLen {
				img.File = img.File[len(img.File)-nameLen:]
			}
			p := filepath.Join("hk-images", img.File)
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
		_, err := B.SendGroupMessage(context.GroupId, append(MessageChain{
			&Reply{Id: strconv.FormatInt(int64(context.MessageId), 10)},
		}, messages...))
		return err
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
	files, err := os.ReadDir("hk-images")
	if err != nil {
		slog.Error("read dir failed", "error", err)
	}
	for _, file := range files {
		if !data2[file.Name()] {
			if err = os.Remove(filepath.Join("hk-images", file.Name())); err != nil {
				slog.Error("remove file failed", "error", err)
			}
		}
	}
}

type dictionaryCommand struct {
	name      string
	tips      string
	checkPerm bool
}

func (d *dictionaryCommand) Name() string {
	return d.name
}

func (d *dictionaryCommand) ShowTips(int64, int64) string {
	return d.tips
}

func (d *dictionaryCommand) CheckAuth(_ int64, senderId int64) bool {
	return !d.checkPerm || IsWhitelist(senderId)
}

func (d *dictionaryCommand) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n" + d.tips}}
	}
	return nil
}

func init() {
	addCmdListener(&dictionaryCommand{name: "添加词条", tips: "添加词条 词条名称", checkPerm: true})
	addCmdListener(&dictionaryCommand{name: "删除词条", tips: "删除词条 词条名称", checkPerm: true})
	addCmdListener(&dictionaryCommand{name: "修改词条", tips: "修改词条 词条名称", checkPerm: true})
	addCmdListener(&dictionaryCommand{name: "搜索词条", tips: "搜索词条 关键词"})
}
