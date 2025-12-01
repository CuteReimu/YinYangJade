package maplebot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	. "github.com/CuteReimu/onebot"
)

var addDbQQList = make(map[int64]string)

func handleDictionary(message *GroupMessage) bool {
	if len(message.Message) == 0 {
		return true
	}
	if !slices.Contains(config.GetIntSlice("qq_groups"), int(message.GroupId)) {
		return true
	}
	if len(message.Message) == 1 {
		if text, ok := message.Message[0].(*Text); ok {
			perm := IsAdmin(message.GroupId, message.Sender.UserId)
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
		if msg, err := saveImage(message.Message); err != nil {
			sendGroupMessage(message, &Text{Text: "编辑词条失败，" + err.Error()})
			return true
		} else {
			message.Message = msg
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

func (d *dictionaryCommand) CheckAuth(group, senderId int64) bool {
	return !d.checkPerm || IsAdmin(group, senderId)
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
