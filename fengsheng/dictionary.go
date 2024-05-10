package fengsheng

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/CuteReimu/mirai-sdk-http"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var addDbQQList = make(map[int64]string)

func handleDictionary(message *GroupMessage) bool {
	if len(message.MessageChain) <= 1 {
		return true
	}
	if !slices.Contains(fengshengConfig.GetIntSlice("qq.qq_group"), int(message.Sender.Group.Id)) {
		return true
	}
	if len(message.MessageChain) == 2 {
		if plain, ok := message.MessageChain[1].(*Plain); ok {
			perm := message.Sender.Permission == PermAdministrator || message.Sender.Permission == PermOwner ||
				fengshengConfig.GetInt64("qq.super_admin_qq") == message.Sender.Id
			if perm && strings.HasPrefix(plain.Text, "添加词条 ") {
				key := dealKey(plain.Text[len("添加词条"):])
				if strings.Contains(key, ".") {
					sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "词条名称中不能包含 . 符号"})
					return true
				}
				if _, ok = cmdMap[key]; ok {
					sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "不能用" + key + "作为词条"})
					return true
				}
				if len(key) > 0 {
					m := qunDb.GetStringMapString("data")
					if _, ok = m[key]; ok {
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "词条已存在"})
					} else {
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "请输入要添加的内容"})
						addDbQQList[message.Sender.Id] = key
					}
				}
				return true
			} else if perm && strings.HasPrefix(plain.Text, "修改词条 ") {
				key := dealKey(plain.Text[len("修改词条"):])
				if len(key) > 0 {
					m := qunDb.GetStringMapString("data")
					if _, ok = m[key]; !ok {
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "词条不存在"})
					} else {
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "请输入要修改的内容"})
						addDbQQList[message.Sender.Id] = key
					}
				}
				return true
			} else if perm && strings.HasPrefix(plain.Text, "删除词条 ") {
				key := dealKey(plain.Text[len("删除词条"):])
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
				return true
			} else if strings.HasPrefix(plain.Text, "查询词条 ") || strings.HasPrefix(plain.Text, "搜索词条 ") {
				key := dealKey(plain.Text[len("搜索词条"):])
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
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "搜索到以下词条：\n" + strings.Join(res, "\n")})
					} else {
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "搜索不到词条(" + key + ")"})
					}
				}
				return true
			}
		}
	}
	if key, ok := addDbQQList[message.Sender.Id]; ok { // 添加词条
		delete(addDbQQList, message.Sender.Id)
		var ms MessageChain
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
		m := qunDb.GetStringMapString("data")
		m[key] = string(buf)
		qunDb.Set("data", m)
		if err = qunDb.WriteConfig(); err != nil {
			slog.Error("write data failed", "error", err)
		}
		sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "编辑词条成功"})
	} else { // 调用词条
		if len(message.MessageChain) == 2 {
			if plain, ok := message.MessageChain[1].(*Plain); ok {
				m := qunDb.GetStringMapString("data")
				s := m[dealKey(plain.Text)]
				if len(s) > 0 {
					var ms MessageChain
					if err := json.Unmarshal([]byte(s), &ms); err != nil {
						slog.Error("json unmarshal failed", "error", err, "s", s)
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "调用词条失败"})
						return true
					}
					sendGroupMessage(message.Sender.Group.Id, ms...)
				}
			}
		}
	}
	return true
}

func saveImage(message MessageChain) error {
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
			if err = os.MkdirAll("dictionary-images", 0755); err != nil {
				slog.Error("mkdir failed", "error", err)
				return errors.New("保存图片失败")
			}
			if err = os.WriteFile(filepath.Join("dictionary-images", img.ImageId), resp.Body(), 0600); err != nil {
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

type dictionaryCommand struct {
	name string
	tips string
}

func (d *dictionaryCommand) Name() string {
	return d.name
}

func (d *dictionaryCommand) ShowTips(int64, int64) string {
	return d.tips
}

func (d *dictionaryCommand) CheckAuth(int64, int64) bool {
	return true
}

func (d *dictionaryCommand) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) == 0 {
		return MessageChain{&Plain{Text: "命令格式：\n" + d.tips}}
	}
	return nil
}

func init() {
	addCmdListener(&dictionaryCommand{name: "添加词条", tips: "添加词条 词条名称"})
	addCmdListener(&dictionaryCommand{name: "删除词条", tips: "删除词条 词条名称"})
	addCmdListener(&dictionaryCommand{name: "修改词条", tips: "修改词条 词条名称"})
	addCmdListener(&dictionaryCommand{name: "搜索词条", tips: "搜索词条 关键词"})
}
