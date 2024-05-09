package tfcc

import (
	. "github.com/CuteReimu/mirai-sdk-http"
	"math/rand"
	"strconv"
	"strings"
)

func init() {
	addCmdListener(&showTips{})
	addCmdListener(&ping{})
	addCmdListener(&randDice{})
	addCmdListener(&roll{})
}

type showTips struct{}

func (t *showTips) Name() string {
	return "查看帮助"
}

func (t *showTips) ShowTips(int64, int64) string {
	return "查看帮助"
}

func (t *showTips) CheckAuth(int64, int64) bool {
	return true
}

func (t *showTips) Execute(msg *GroupMessage, _ string) []SingleMessage {
	var ret []string
	cmdMap.Range(func(_, v any) bool {
		hList, _ := v.([]CmdHandler)
		for _, h := range hList {
			if h.CheckAuth(msg.Sender.Group.Id, msg.Sender.Id) {
				if tip := h.ShowTips(msg.Sender.Group.Id, msg.Sender.Id); len(tip) > 0 {
					ret = append(ret, tip)
				}
			}
		}
		return true
	})
	return []SingleMessage{&Plain{Text: "你可以使用以下功能：\n" + strings.Join(ret, "\n")}}
}

type ping struct{}

func (p *ping) Name() string {
	return "ping"
}

func (p *ping) ShowTips(int64, int64) string {
	return ""
}

func (p *ping) CheckAuth(int64, int64) bool {
	return true
}

func (p *ping) Execute(_ *GroupMessage, content string) []SingleMessage {
	if len(content) == 0 {
		return []SingleMessage{&Plain{Text: "pong"}}
	}
	return nil
}

type randDice struct{}

func (r *randDice) Name() string {
	return "随机骰子"
}

func (r *randDice) ShowTips(int64, int64) string {
	return ""
}

func (r *randDice) CheckAuth(int64, int64) bool {
	return true
}

func (r *randDice) Execute(_ *GroupMessage, content string) []SingleMessage {
	if len(content) == 0 {
		return []SingleMessage{&Dice{Value: rand.Int31n(6) + 1}}
	}
	return nil
}

type roll struct{}

func (r *roll) Name() string {
	return "roll"
}

func (r *roll) ShowTips(int64, int64) string {
	return ""
}

func (r *roll) CheckAuth(int64, int64) bool {
	return true
}

func (r *roll) Execute(message *GroupMessage, content string) []SingleMessage {
	if len(content) == 0 {
		return []SingleMessage{&Plain{Text: message.Sender.MemberName + " roll: " + strconv.Itoa(rand.Intn(100))}}
	}
	return nil
}
