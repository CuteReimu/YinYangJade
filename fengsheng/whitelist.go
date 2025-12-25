package fengsheng

import (
	"log/slog"
	"strconv"
	"strings"

	. "github.com/CuteReimu/onebot"
)

func init() {
	addCmdListener(delWhitelist{})
	addCmdListener(addWhitelist{})
}

type delWhitelist struct{}

func (delWhitelist) Name() string {
	return "删除词条权限"
}

func (delWhitelist) ShowTips(int64, int64) string {
	return "删除词条权限 对方QQ号"
}

func (delWhitelist) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (delWhitelist) Execute(_ *GroupMessage, content string) MessageChain {
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
		if !isWhitelist(qq) {
			return MessageChain{&Text{Text: s + "并没有词条权限"}}
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return MessageChain{&Text{Text: "指令格式如下：\n删除词条权限 对方QQ号"}}
	}
	for _, qq := range qqNumbers {
		doRemoveWhitelist(qq)
	}
	ret := "已删除词条权限"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain{&Text{Text: ret}}
}

type addWhitelist struct{}

func (addWhitelist) Name() string {
	return "增加词条权限"
}

func (addWhitelist) ShowTips(int64, int64) string {
	return "增加词条权限 对方QQ号"
}

func (addWhitelist) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (addWhitelist) Execute(msg *GroupMessage, content string) MessageChain {
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
		if isWhitelist(qq) {
			return MessageChain{&Text{Text: s + "已经有词条权限了"}}
		}
		if _, err := B.GetGroupMemberInfo(msg.GroupId, qq, false); err != nil {
			return MessageChain{&Text{Text: s + "不是群成员"}}
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return MessageChain{&Text{Text: "指令格式如下：\n增加词条权限 对方QQ号"}}
	}
	for _, qq := range qqNumbers {
		doAddWhitelist(qq)
	}
	ret := "已增加词条权限"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain{&Text{Text: ret}}
}
