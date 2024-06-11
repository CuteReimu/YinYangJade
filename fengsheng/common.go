package fengsheng

import (
	"github.com/CuteReimu/YinYangJade/slicegame"
	. "github.com/CuteReimu/onebot"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
)

func init() {
	addCmdListener(&showTips{})
	addCmdListener(&ping{})
	addCmdListener(&roll{})
	addCmdListener(&sliceGame{})
}

type showTips struct{}

func (t *showTips) Name() string {
	return "查看帮助"
}

func (t *showTips) ShowTips(int64, int64) string {
	return ""
}

func (t *showTips) CheckAuth(int64, int64) bool {
	return true
}

func (t *showTips) Execute(msg *GroupMessage, _ string) MessageChain {
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

func (p *ping) Name() string {
	return "ping"
}

func (p *ping) ShowTips(int64, int64) string {
	return ""
}

func (p *ping) CheckAuth(int64, int64) bool {
	return true
}

func (p *ping) Execute(_ *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		return MessageChain{&Text{Text: "pong"}}
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

func (r *roll) Execute(message *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		replyGroupMessage(true, message, &Text{Text: "roll: " + strconv.Itoa(rand.IntN(100))})
	}
	return nil
}

type sliceGame struct{}

func (r *sliceGame) Name() string {
	return "滑块"
}

func (r *sliceGame) ShowTips(int64, int64) string {
	return ""
}

func (r *sliceGame) CheckAuth(int64, int64) bool {
	return true
}

func (r *sliceGame) Execute(message *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		sendGroupMessage(message, slicegame.DoStuff()...)
	}
	return nil
}
