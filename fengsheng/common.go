package fengsheng

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"github.com/CuteReimu/YinYangJade/slicegame"
	. "github.com/CuteReimu/onebot"
)

func init() {
	addCmdListener(&showTips{})
	addCmdListener(&ping{})
	addCmdListener(&roll{})
	addCmdListener(&sliceGame{})
}

type showTips struct{}

func (showTips) Name() string {
	return "查看帮助"
}

func (showTips) ShowTips(int64, int64) string {
	return ""
}

func (showTips) CheckAuth(int64, int64) bool {
	return true
}

func (showTips) Execute(msg *GroupMessage, _ string) MessageChain {
	var ret []string
	for _, h := range cmdMap {
		if h.CheckAuth(msg.GroupId, msg.Sender.UserId) {
			if tip := h.ShowTips(msg.GroupId, msg.Sender.UserId); len(tip) > 0 {
				ret = append(ret, tip)
			}
		}
	}
	slices.Sort(ret)
	return MessageChain{&Text{Text: "你可以使用以下功能：\n" + strings.Join(ret, "\n")}}
}

type ping struct{}

func (ping) Name() string {
	return "ping"
}

func (ping) ShowTips(int64, int64) string {
	return ""
}

func (ping) CheckAuth(int64, int64) bool {
	return true
}

func (ping) Execute(_ *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		return MessageChain{&Text{Text: "pong"}}
	}
	return nil
}

type roll struct{}

func (roll) Name() string {
	return "roll"
}

func (roll) ShowTips(int64, int64) string {
	return ""
}

func (roll) CheckAuth(int64, int64) bool {
	return true
}

func (roll) Execute(message *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		replyGroupMessage(message, &Text{Text: "roll: " + strconv.Itoa(rand.IntN(100))})
	}
	return nil
}

type sliceGame struct{}

func (sliceGame) Name() string {
	return "滑块"
}

func (sliceGame) ShowTips(int64, int64) string {
	return ""
}

func (sliceGame) CheckAuth(int64, int64) bool {
	return true
}

func (sliceGame) Execute(message *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		sendGroupMessage(message, slicegame.DoStuff()...)
	}
	return nil
}
