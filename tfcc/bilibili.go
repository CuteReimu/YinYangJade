package tfcc

import (
	"github.com/CuteReimu/bilibili/v2"
	"net/http"
	"slices"
	"strings"
)

var bili = bilibili.New()

func initBilibili() {
	cookies := bilibiliData.GetStringSlice("cookies")
	bili.SetCookiesString(strings.Join(cookies, "\n"))
	if !slices.ContainsFunc(bili.GetCookies(), func(cookie *http.Cookie) bool {
		return strings.HasPrefix(cookie.Name, "bili_jct")
	}) { // TODO
		panic("暂未支持登录，请先复制Cookies使用")
	}
}
