package fengsheng

import (
	"log/slog"
	"strconv"
	"strings"

	. "github.com/CuteReimu/onebot"
)

func init() {
	addCmdListener(&delAdmin{})
	addCmdListener(&addAdmin{})
	addCmdListener(&listAllAdmin{})
}

type delAdmin struct{}

func (*delAdmin) Name() string {
	return "删除管理员"
}

func (*delAdmin) ShowTips(int64, int64) string {
	return "删除管理员 对方QQ号"
}

func (*delAdmin) CheckAuth(_ int64, senderID int64) bool {
	return isSuperAdmin(senderID)
}

func (*delAdmin) Execute(_ *GroupMessage, content string) MessageChain {
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
		if isSuperAdmin(qq) {
			return MessageChain{&Text{Text: "你不能删除自己"}}
		}
		if !isAdmin(qq) {
			return MessageChain{&Text{Text: s + "并不是管理员"}}
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return nil
	}
	for _, qq := range qqNumbers {
		doRemoveAdmin(qq)
	}
	ret := "已删除管理员"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain{&Text{Text: ret}}
}

type addAdmin struct{}

func (*addAdmin) Name() string {
	return "增加管理员"
}

func (*addAdmin) ShowTips(int64, int64) string {
	return "增加管理员 对方QQ号"
}

func (*addAdmin) CheckAuth(_ int64, senderID int64) bool {
	return isSuperAdmin(senderID)
}

func (*addAdmin) Execute(_ *GroupMessage, content string) MessageChain {
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
		if isSuperAdmin(qq) || isAdmin(qq) {
			return MessageChain{&Text{Text: s + "已经是管理员了"}}
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return nil
	}
	for _, qq := range qqNumbers {
		doAddAdmin(qq)
	}
	ret := "已增加管理员"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain{&Text{Text: ret}}
}

type listAllAdmin struct{}

func (*listAllAdmin) Name() string {
	return "查看管理员"
}

func (*listAllAdmin) ShowTips(int64, int64) string {
	return ""
}

func (*listAllAdmin) CheckAuth(int64, int64) bool {
	return true
}

func (*listAllAdmin) Execute(*GroupMessage, string) MessageChain {
	superAdmin := fengshengConfig.GetInt64("qq.super_admin_qq")
	admin := permData.GetIntSlice("admin")
	s := make([]string, 0, len(admin)+2)
	s = append(s, "管理员列表：")
	s = append(s, strconv.FormatInt(superAdmin, 10))
	for _, qq := range admin {
		s = append(s, strconv.Itoa(qq))
	}
	return MessageChain{&Text{Text: strings.Join(s, "\n")}}
}
