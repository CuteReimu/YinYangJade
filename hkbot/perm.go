package hkbot

import (
	"log/slog"
	"slices"
)

func isSuperAdmin(qq int64) bool {
	return qq == hkConfig.GetInt64("qq.super_admin_qq")
}

func isAdmin(qq int64) bool {
	return qq == hkConfig.GetInt64("qq.super_admin_qq") || slices.Contains(permData.GetIntSlice("admin"), int(qq))
}

func doAddAdmin(qq int64) bool {
	admins := permData.GetIntSlice("admin")
	if slices.Contains(admins, int(qq)) {
		return false
	}
	permData.Set("admin", append(admins, int(qq)))
	err := permData.WriteConfig()
	if err != nil {
		slog.Error("write config failed", "error", err)
	}
	return true
}

func doRemoveAdmin(qq int64) bool {
	admins := permData.GetIntSlice("admin")
	index := slices.Index(admins, int(qq))
	if index < 0 {
		return false
	}
	permData.Set("admin", append(admins[:index], admins[index+1:]...))
	err := permData.WriteConfig()
	if err != nil {
		slog.Error("write config failed", "error", err)
	}
	return true
}

func isWhitelist(qq int64) bool {
	return slices.Contains(permData.GetIntSlice("dictModify"), int(qq))
}

func doAddWhitelist(qq int64) bool {
	whitelist := permData.GetIntSlice("dictModify")
	if slices.Contains(whitelist, int(qq)) {
		return false
	}
	permData.Set("dictModify", append(whitelist, int(qq)))
	err := permData.WriteConfig()
	if err != nil {
		slog.Error("write config failed", "error", err)
	}
	return true
}

func doRemoveWhitelist(qq int64) bool {
	whitelist := permData.GetIntSlice("dictModify")
	index := slices.Index(whitelist, int(qq))
	if index < 0 {
		return false
	}
	permData.Set("dictModify", append(whitelist[:index], whitelist[index+1:]...))
	err := permData.WriteConfig()
	if err != nil {
		slog.Error("write config failed", "error", err)
	}
	return true
}
