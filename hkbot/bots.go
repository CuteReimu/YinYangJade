package hkbot

import (
	"encoding/json"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"
)

var restyClient = resty.New()

func init() {
	restyClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	restyClient.SetTimeout(20 * time.Second)
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69",
		"connection":   "close",
		"Accept":       "application/json",
	})
}

var B *Bot

func Init(b *Bot) {
	initConfig()
	B = b
	go func() {
		for range time.Tick(time.Duration(hkConfig.GetInt64("speedrun_push_delay")) * time.Second) {
			doTimer()
		}
	}()
}

var re = regexp.MustCompile("<.*?>")

func doTimer() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered", "error", err)
		}
	}()
	resp, err := restyClient.R().SetHeader("X-API-Key", hkConfig.GetString("speedrun_api_key")).
		Get("https://www.speedrun.com/api/v1/notifications")
	if err != nil {
		slog.Error("get speedrun notifications failed", "error", err)
		return
	}
	if resp.StatusCode() != 200 {
		slog.Error("get speedrun notifications failed", "status", resp.Status(), "body", resp.String())
		return
	}
	var data *RespData
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		slog.Error("unmarshal speedrun notifications failed", "error", err)
		return
	}
	var result1 []string
	var changed bool
	pushedMessages := hkData.GetStringSlice("pushedMessages")
	for _, v := range data.Data {
		if slices.Contains(pushedMessages, v.Id) {
			continue
		}
		pushedMessages = append(pushedMessages, v.Id)
		changed = true
		s := re.ReplaceAllString(v.Text, "")
		if strings.Contains(s, "beat the WR") || strings.Contains(s, "got a new top 3 PB") {
			result1 = append(result1, translate(s))
		}
	}
	if len(pushedMessages) > 100 {
		pushedMessages = pushedMessages[len(pushedMessages)-100:]
		changed = true
	}
	for _, groupId := range hkConfig.GetIntSlice("speedrun_push_qq_group") {
		_, err := B.SendGroupMessage(int64(groupId), 0, MessageChain{&Plain{Text: strings.Join(result1, "\n")}})
		if err != nil {
			slog.Error("send group message failed", "error", err)
		}
	}
	if changed {
		hkData.Set("pushedMessages", pushedMessages)
		if err = hkData.WriteConfig(); err != nil {
			slog.Error("write hkData failed", "error", err)
		}
	}
}

type RespData struct {
	Data []SpeedrunData `json:"data"`
}

type SpeedrunData struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}
