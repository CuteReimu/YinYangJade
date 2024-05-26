package tfcc

import (
	"fmt"
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
	addCmdListener(&randDraw{})
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

func (r *roll) Execute(_ *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		return MessageChain{&Text{Text: "roll: " + strconv.Itoa(rand.IntN(100))}} //nolint:gosec
	}
	return nil
}

type randDraw struct{}

func (r *randDraw) Name() string {
	return "抽签"
}

func (r *randDraw) ShowTips(int64, int64) string {
	return "抽签 选手数量"
}

func (r *randDraw) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (r *randDraw) Execute(_ *GroupMessage, content string) MessageChain {
	count, err := strconv.Atoi(content)
	if err != nil {
		return MessageChain{&Text{Text: "指令格式如下：\n抽签 选手数量"}}
	}
	if count%2 != 0 {
		return MessageChain{&Text{Text: "选手数量必须为偶数"}}
	}
	if count > 32 {
		return MessageChain{&Text{Text: "选手数量太多"}}
	}
	if count <= 0 {
		return MessageChain{&Text{Text: "选手数量太少"}}
	}
	a := make([]int, 0, count)
	for i := count; i > 0; i-- {
		a = append(a, i)
	}
	ret := make([]string, 0, count/2)
	for i := 0; i < count/2; i++ {
		a1 := a[len(a)-1]
		a = a[:len(a)-1]
		index := rand.IntN(len(a)) //nolint:gosec
		a2 := a[index]
		a = append(a[:index], a[index+1:]...)
		ret = append(ret, fmt.Sprintf("%d号 对阵 %d号", a1, a2))
	}
	return MessageChain{&Text{Text: strings.Join(ret, "\n")}}
}
