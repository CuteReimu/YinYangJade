package tfcc

import (
	"github.com/CuteReimu/YinYangJade/db"
	"github.com/spf13/viper"
	"strconv"
)

const adminPrefix = "tfcc-admin:"

func IsSuperAdmin(qq int64) bool {
	return qq == viper.GetInt64("qq.super_admin_qq")
}

func IsAdmin(qq int64) bool {
	if IsSuperAdmin(qq) {
		return true
	}
	buf := db.Get([]byte(adminPrefix + strconv.FormatInt(qq, 10)))
	return buf != nil
}

func AddAdmin(qq int64) {
	db.Set([]byte(adminPrefix+strconv.FormatInt(qq, 10)), []byte{'1'})
}

func DelAdmin(qq int64) {
	db.Del([]byte(adminPrefix + strconv.FormatInt(qq, 10)))
}

// ListAdmin 因为这个接口一般用来展示，所以返回[]string
func ListAdmin() (list []string) {
	list = append(list, strconv.FormatInt(viper.GetInt64("qq.super_admin_qq"), 10))
	db.PrefixScanKeyValue([]byte(adminPrefix), func(key, value []byte) error {
		if len(key) > len(adminPrefix) {
			list = append(list, string(key)[len(adminPrefix):])
		}
		return nil
	})
	return
}

const whitelistPrefix = "whitelist:"

func IsWhitelist(qq int64) bool {
	buf := db.Get([]byte(whitelistPrefix + strconv.FormatInt(qq, 10)))
	return buf != nil && string(buf) == "1"
}

func AddWhitelist(qq int64) {
	db.Set([]byte(whitelistPrefix+strconv.FormatInt(qq, 10)), []byte{'1'})
}

func DelWhitelist(qq int64) {
	db.Del([]byte(whitelistPrefix + strconv.FormatInt(qq, 10)))
}

func DisableAllWhitelist() (count int) {
	db.PrefixUpdateKey([]byte(whitelistPrefix), func([]byte) ([]byte, error) {
		count++
		return []byte{'0'}, nil
	})
	return
}

func EnableAllWhitelist() (count int) {
	db.PrefixUpdateKey([]byte(whitelistPrefix), func([]byte) ([]byte, error) {
		count++
		return []byte{'1'}, nil
	})
	return
}

// ListWhitelist 因为这个接口一般用来展示，所以返回[]string
func ListWhitelist() (list []string) {
	db.PrefixScanKeyValue([]byte(whitelistPrefix), func(key, val []byte) error {
		if len(key) > len(whitelistPrefix) && val != nil && string(val) == "1" {
			list = append(list, string(key)[len(whitelistPrefix):])
		}
		return nil
	})
	return
}
