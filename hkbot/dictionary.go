package hkbot

import (
	"encoding/json"
	"log/slog"
	"slices"
	"strings"

	"github.com/CuteReimu/YinYangJade/botutil"
	. "github.com/CuteReimu/onebot"
)

var addDbQQList = make(map[int64]string)

func dealAddDict(message *GroupMessage, key string) {
	botutil.DealAddDict(message, key, qunDb, cmdMap, addDbQQList, sendGroupMessage)
}

func dealModifyDict(message *GroupMessage, key string) {
	botutil.DealModifyDict(message, key, qunDb, addDbQQList, sendGroupMessage)
}

func dealRemoveDict(message *GroupMessage, key string) {
	botutil.DealRemoveDict(message, key, qunDb, sendGroupMessage)
}

func dealSearchDict(message *GroupMessage, key string) {
	botutil.DealSearchDict(message, key, qunDb, sendGroupMessage)
}

func handleDictionary(message *GroupMessage) bool {
	if len(message.Message) == 0 {
		return true
	}
	if !slices.Contains(hkConfig.GetIntSlice("speedrun_push_qq_group"), int(message.GroupId)) {
		return true
	}
	if len(message.Message) != 1 {
		return true
	}
	if text, ok := message.Message[0].(*Text); ok {
		if strings.HasPrefix(text.Text, "查询词条 ") || strings.HasPrefix(text.Text, "搜索词条 ") {
			key := botutil.DealKey(text.Text[len("搜索词条"):])
			if len(key) > 0 {
				dealSearchDict(message, key)
			}
			return true
		}
		if isWhitelist(message.Sender.UserId) {
			switch {
			case strings.HasPrefix(text.Text, "添加词条 "):
				key := botutil.DealKey(text.Text[len("添加词条"):])
				dealAddDict(message, key)
				return true
			case strings.HasPrefix(text.Text, "修改词条 "):
				key := botutil.DealKey(text.Text[len("修改词条"):])
				if len(key) > 0 {
					dealModifyDict(message, key)
				}
				return true
			case strings.HasPrefix(text.Text, "删除词条 "):
				key := botutil.DealKey(text.Text[len("删除词条"):])
				if len(key) > 0 {
					dealRemoveDict(message, key)
				}
				return true
			}
		}
	}
	if key, ok := addDbQQList[message.Sender.UserId]; ok { // 添加词条
		delete(addDbQQList, message.Sender.UserId)
		msg, err := saveImage(message.Message)
		if err != nil {
			sendGroupMessage(message, &Text{Text: "编辑词条失败，" + err.Error()})
			return true
		}
		message.Message = msg
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
				s := m[botutil.DealKey(text.Text)]
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

func saveImage(message MessageChain) (MessageChain, error) {
	return botutil.SaveImage(message, "hk-images", B)
}

func sendGroupMessage(context *GroupMessage, messages ...SingleMessage) {
	replyGroupMessage(false, context, messages...)
}

func replyGroupMessage(reply bool, context *GroupMessage, messages ...SingleMessage) {
	if reply {
		botutil.ReplyGroupMessageWithRetry(B, context, botutil.FillSpecificMessage, messages...)
	} else {
		botutil.SendGroupMessageWithRetry(B, context, botutil.FillSpecificMessage, messages...)
	}
}

func clearExpiredImages() {
	botutil.ClearExpiredImages(qunDb, "hk-images")
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

func (d *dictionaryCommand) CheckAuth(_ int64, senderID int64) bool {
	return !d.checkPerm || isWhitelist(senderID)
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
