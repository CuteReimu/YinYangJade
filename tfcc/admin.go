package tfcc

import (
	"github.com/CuteReimu/YinYangJade/bots"
	. "github.com/CuteReimu/mirai-sdk-http"
	"log/slog"
	"strconv"
	"strings"
)

func init() {
	bots.AddCmdListener(&delAdmin{})
	bots.AddCmdListener(&addAdmin{})
	bots.AddCmdListener(&listAllAdmin{})
}

type delAdmin struct{}

func (d *delAdmin) Name() string {
	return "删除管理员"
}

func (d *delAdmin) ShowTips(int64, int64) string {
	return "删除管理员 对方QQ号"
}

func (d *delAdmin) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (d *delAdmin) Execute(_ *GroupMessage, content string) []SingleMessage {
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
		if IsSuperAdmin(qq) {
			return MessageChain(&Plain{Text: "你不能删除自己"})
		}
		if !IsAdmin(qq) {
			return MessageChain(&Plain{Text: s + "并不是管理员"})
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return nil
	}
	for _, qq := range qqNumbers {
		DelAdmin(qq)
	}
	ret := "已删除管理员"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain(&Plain{Text: ret})
}

type addAdmin struct{}

func (a *addAdmin) Name() string {
	return "增加管理员"
}

func (a *addAdmin) ShowTips(int64, int64) string {
	return "增加管理员 对方QQ号"
}

func (a *addAdmin) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (a *addAdmin) Execute(_ *GroupMessage, content string) []SingleMessage {
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
		if IsSuperAdmin(qq) || IsAdmin(qq) {
			return MessageChain(&Plain{Text: s + "已经是管理员了"})
		}
		qqNumbers = append(qqNumbers, qq)
	}
	if len(qqNumbers) == 0 {
		return nil
	}
	for _, qq := range qqNumbers {
		AddAdmin(qq)
	}
	ret := "已增加管理员"
	if len(qqNumbers) == 1 {
		ret += "：" + strconv.FormatInt(qqNumbers[0], 10)
	}
	return MessageChain(&Plain{Text: ret})
}

type listAllAdmin struct{}

func (g *listAllAdmin) Name() string {
	return "查看管理员"
}

func (g *listAllAdmin) ShowTips(int64, int64) string {
	return "查看管理员"
}

func (g *listAllAdmin) CheckAuth(int64, int64) bool {
	return true
}

func (g *listAllAdmin) Execute(_ *GroupMessage, content string) []SingleMessage {
	if len(content) != 0 {
		return nil
	}
	list := ListAdmin()
	if len(list) > 0 {
		return MessageChain(&Plain{Text: "管理员列表：\n" + strings.Join(list, "\n")})
	}
	return nil
}
