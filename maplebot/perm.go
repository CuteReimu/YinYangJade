package maplebot

import (
	"log/slog"

	"github.com/CuteReimu/onebot"
)

func IsSuperAdmin(qq int64) bool {
	return qq == config.GetInt64("admin")
}

func IsAdmin(group, qq int64) bool {
	if IsSuperAdmin(qq) {
		return true
	}
	info, err := B.GetGroupMemberInfo(group, qq, false)
	if err != nil {
		slog.Error("获取群成员信息失败", "error", err)
		return false
	}
	return info.Role == onebot.RoleAdmin || info.Role == onebot.RoleOwner
}
