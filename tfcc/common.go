package tfcc

import (
	"fmt"
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
	addCmdListener(&randDraw{})
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

type randDraw struct{}

func (r *randDraw) Name() string {
	return "抽签"
}

func (r *randDraw) ShowTips(int64, int64) string {
	return "抽签 选手数量"
}

func (r *randDraw) CheckAuth(int64, int64) bool {
	return true
}

func (r *randDraw) Execute(_ *GroupMessage, content string) MessageChain {
	count, err := strconv.Atoi(content)
	if err != nil {
		return MessageChain{&Text{Text: "指令格式如下：\n抽签 选手数量"}}
	}
	if count > 32 {
		return MessageChain{&Text{Text: "选手数量太多"}}
	}
	if count <= 2 {
		return MessageChain{&Text{Text: "选手数量太少"}}
	}
	isOdd := count%2 == 1
	if isOdd {
		count++
	}
	a := make([]int, 0, count)
	for i := 1; i <= count; i++ {
		a = append(a, i)
	}
	rand.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
	result := slices.Collect(slices.Chunk(a, 2))
	for _, pair := range result {
		if pair[0] > pair[1] {
			pair[0], pair[1] = pair[1], pair[0]
		}
	}
	slices.SortFunc(result, func(pair1 []int, pair2 []int) int {
		return pair1[0] - pair2[0]
	})
	ret := make([]string, 0, count/2)
	for _, pair := range result {
		if isOdd && pair[1] == count {
			ret = append(ret, fmt.Sprintf("%d号 轮空", pair[0]))
		} else {
			ret = append(ret, fmt.Sprintf("%d号 对阵 %d号", pair[0], pair[1]))
		}
	}
	return MessageChain{&Text{Text: strings.Join(ret, "\n")}}
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
