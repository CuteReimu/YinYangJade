package tfcc

import (
	. "github.com/CuteReimu/mirai-sdk-http"
	"log/slog"
	"strconv"
	"strings"
)

func init() {
	addCmdListener(&delWhitelist{})
	addCmdListener(&addWhitelist{})
	addCmdListener(&checkWhitelist{})
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
	ss := strings.Split(content, " ")
	qqNumbers := make([]int64, 0, len(ss))
	for _, s := range ss {
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
			return []SingleMessage{&Plain{Text: s + "并不是白名单"}}
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return []SingleMessage{&Plain{Text: "指令格式如下：\n删除白名单 对方QQ号"}}
	}
	for _, qq := range qqNumbers {
		RemoveWhitelist(qq)
	}
	ret := "已删除白名单"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return []SingleMessage{&Plain{Text: ret}}
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
	ss := strings.Split(content, " ")
	qqNumbers := make([]int64, 0, len(ss))
	for _, s := range ss {
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
			return []SingleMessage{&Plain{Text: s + "已经是白名单了"}}
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return []SingleMessage{&Plain{Text: "指令格式如下：\n增加白名单 对方QQ号"}}
	}
	for _, qq := range qqNumbers {
		AddWhitelist(qq)
	}
	ret := "已增加白名单"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return []SingleMessage{&Plain{Text: ret}}
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
		return []SingleMessage{&Plain{Text: "指令格式如下：\n查看白名单 对方QQ号"}}
	}
	if IsWhitelist(qq) {
		return []SingleMessage{&Plain{Text: content + "是白名单"}}
	} else {
		return []SingleMessage{&Plain{Text: content + "不是白名单"}}
	}
}
