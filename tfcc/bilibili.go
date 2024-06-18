package tfcc

import (
	"fmt"
	"github.com/CuteReimu/bilibili/v2"
	"log/slog"
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
	}) {
		code, err := bili.GetQRCode()
		if err != nil {
			slog.Error("获取二维码失败", "error", err)
			fmt.Println("获取二维码失败，很多需要登录B站的功能将无法使用", err)
			return
		}
		fmt.Println("请扫描二维码登录B站")
		code.Print()
		result, err := bili.LoginWithQRCode(bilibili.LoginWithQRCodeParam{
			QrcodeKey: code.QrcodeKey,
		})
		if err != nil {
			slog.Error("登录失败", "error", err)
			fmt.Println("登录失败，很多需要登录B站的功能将无法使用", err)
			return
		}
		if result.Code != 0 {
			slog.Error("登录失败", "code", result.Code, "message", result.Message)
			fmt.Println("登录失败，很多需要登录B站的功能将无法使用", result.Code)
			return
		}
		cookiesString := bili.GetCookiesString()
		bilibiliData.Set("cookies", strings.Split(cookiesString, "\n"))
	}
}
