package hkbot

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/CuteReimu/YinYangJade/hkbot/scripts"
	. "github.com/CuteReimu/onebot"
	"golang.org/x/sync/singleflight"
)

func init() {
	addCmdListener(&speedrunLeaderboards{
		availableInputs: []string{"Any%", "TE", "100%", "Judgement", "Low%", "AB", "Twisted%"},
		aliasInputs:     []string{"all bosses", "all boss", "allbosses", "allboss"},
	})
}

type speedrunLeaderboards struct {
	availableInputs []string
	aliasInputs     []string
	sf              singleflight.Group
}

func (t *speedrunLeaderboards) Name() string {
	return "查榜"
}

func (t *speedrunLeaderboards) ShowTips(int64, int64) string {
	return "查榜"
}

func (t *speedrunLeaderboards) CheckAuth(int64, int64) bool {
	return true
}

func (t *speedrunLeaderboards) Execute(_ *GroupMessage, content string) MessageChain {
	content = strings.ToLower(strings.ReplaceAll(content, "%", ""))
	eq := func(s string) bool {
		return strings.EqualFold(content, strings.ReplaceAll(s, "%", ""))
	}
	if !slices.ContainsFunc(t.availableInputs, eq) && !slices.ContainsFunc(t.aliasInputs, eq) {
		return MessageChain{&Text{Text: "支持的榜单类型有：" + strings.Join(t.availableInputs, "，")}}
	}
	result, err, _ := t.sf.Do(content, func() (any, error) {
		return scripts.RunPythonScript("speedrun_silksong.py", content)
	})
	if err != nil {
		output, ok := result.([]byte)
		if ok {
			slog.Error("查询失败", "output", string(output), "error", err)
		} else {
			slog.Error("查询失败", "output", result, "error", err)
		}
		return nil
	}
	output, ok := result.([]byte)
	if !ok {
		slog.Error("查询结果类型错误", "type", fmt.Sprintf("%T", result))
		return nil
	}
	return MessageChain{&Text{Text: string(output)}}
}
