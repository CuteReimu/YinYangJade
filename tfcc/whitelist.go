package tfcc

import (
	"fmt"
	"github.com/CuteReimu/YinYangJade/bots"
	. "github.com/CuteReimu/mirai-sdk-http"
	"log/slog"
	"strconv"
	"strings"
)

func init() {
	bots.AddCmdListener(&delWhitelist{})
	bots.AddCmdListener(&addWhitelist{})
	bots.AddCmdListener(&listAllWhitelist{})
	bots.AddCmdListener(&checkWhitelist{})
	bots.AddCmdListener(&enableAllWhitelist{})
	bots.AddCmdListener(&disableAllWhitelist{})
}

type delWhitelist struct{}

func (d *delWhitelist) Name() string {
	return "删除白名单"
}

func (d *delWhitelist) ShowTips(int64, int64) string {
	return "删除白名单 对方QQ号"
}

func (d *delWhitelist) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (d *delWhitelist) Execute(_ *GroupMessage, content string) []SingleMessage {
	var qqNumbers []int64
	for _, s := range strings.Split(content, " ") {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		qq, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			slog.Error("parse failed: "+s, "error", err)
			return nil
		}
		if !IsWhitelist(qq) {
			return MessageChain(&Plain{Text: s + "并不是白名单"})
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return MessageChain(&Plain{Text: "指令格式如下：\n删除白名单 对方QQ号"})
	}
	for _, qq := range qqNumbers {
		DelWhitelist(qq)
	}
	ret := "已删除白名单"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain(&Plain{Text: ret})
}

type addWhitelist struct{}

func (a *addWhitelist) Name() string {
	return "增加白名单"
}

func (a *addWhitelist) ShowTips(int64, int64) string {
	return "增加白名单 对方QQ号"
}

func (a *addWhitelist) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *addWhitelist) Execute(_ *GroupMessage, content string) []SingleMessage {
	var qqNumbers []int64
	for _, s := range strings.Split(content, " ") {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		qq, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			slog.Error("parse failed: "+s, "error", err)
			return nil
		}
		if IsWhitelist(qq) {
			return MessageChain(&Plain{Text: s + "已经是白名单了"})
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return MessageChain(&Plain{Text: "指令格式如下：\n增加白名单 对方QQ号"})
	}
	for _, qq := range qqNumbers {
		AddWhitelist(qq)
	}
	ret := "已增加白名单"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain(&Plain{Text: ret})
}

type listAllWhitelist struct{}

func (g *listAllWhitelist) Name() string {
	return "列出所有白名单"
}

func (g *listAllWhitelist) ShowTips(int64, int64) string {
	return "列出所有白名单"
}

func (g *listAllWhitelist) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (g *listAllWhitelist) Execute(_ *GroupMessage, content string) []SingleMessage {
	if len(content) != 0 {
		return nil
	}
	list := ListWhitelist()
	if len(list) > 0 {
		return MessageChain(&Plain{Text: "白名单列表：\n" + strings.Join(list, "\n")})
	}
	return nil
}

type checkWhitelist struct{}

func (e *checkWhitelist) Name() string {
	return "查看白名单"
}

func (e *checkWhitelist) ShowTips(int64, int64) string {
	return "查看白名单 对方QQ号"
}

func (e *checkWhitelist) CheckAuth(int64, int64) bool {
	return true
}

func (e *checkWhitelist) Execute(_ *GroupMessage, content string) []SingleMessage {
	qq, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		return MessageChain(&Plain{Text: "指令格式如下：\n查看白名单 对方QQ号"})
	}
	if IsWhitelist(qq) {
		return MessageChain(&Plain{Text: content + "是白名单"})
	} else {
		return MessageChain(&Plain{Text: content + "不是白名单"})
	}
}

type enableAllWhitelist struct{}

func (e *enableAllWhitelist) Name() string {
	return "启用所有白名单"
}

func (e *enableAllWhitelist) ShowTips(int64, int64) string {
	return "启用所有白名单"
}

func (e *enableAllWhitelist) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (e *enableAllWhitelist) Execute(*GroupMessage, string) []SingleMessage {
	count := EnableAllWhitelist()
	return MessageChain(&Plain{Text: fmt.Sprintf("已启用%d个白名单", count)})
}

type disableAllWhitelist struct{}

func (d *disableAllWhitelist) Name() string {
	return "禁用所有白名单"
}

func (d *disableAllWhitelist) ShowTips(int64, int64) string {
	return "禁用所有白名单"
}

func (d *disableAllWhitelist) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (d *disableAllWhitelist) Execute(*GroupMessage, string) []SingleMessage {
	count := DisableAllWhitelist()
	return MessageChain(&Plain{Text: fmt.Sprintf("已禁用%d个白名单", count)})
}
