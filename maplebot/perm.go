package maplebot

import (
	"log/slog"

	"github.com/CuteReimu/onebot"
)

func isSuperAdmin(qq int64) bool {
	return qq == config.GetInt64("admin")
}

func isAdmin(group, qq int64) bool {
	if isSuperAdmin(qq) {
		return true
	}
	info, err := bot.GetGroupMemberInfo(group, qq, false)
	if err != nil {
		slog.Error("获取群成员信息失败", "error", err)
		return false
	}
	return info.Role == onebot.RoleAdmin || info.Role == onebot.RoleOwner
}
