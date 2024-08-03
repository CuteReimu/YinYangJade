package hkbot

import (
	"encoding/json"
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
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
	go func() {
		for range time.Tick(24 * time.Hour) {
			B.Run(clearExpiredImages)
		}
	}()
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(handleDictionary)
}

var cmdMap = make(map[string]iface.CmdHandler)

func cmdHandleFunc(message *GroupMessage) bool {
	if !slices.Contains(hkConfig.GetIntSlice("speedrun_push_qq_group"), int(message.GroupId)) {
		return true
	}
	chain := message.Message
	if len(chain) == 0 {
		return true
	}
	if at, ok := chain[0].(*At); ok && at.QQ == strconv.FormatInt(B.QQ, 10) {
		chain = chain[1:]
		if len(chain) > 0 {
			if text, ok := chain[0].(*Text); ok && len(strings.TrimSpace(text.Text)) == 0 {
				chain = chain[1:]
			}
		}
		if len(chain) == 0 {
			chain = append(chain, &Text{Text: "查看帮助"})
		}
	}
	var cmd, content string
	if len(chain) == 1 {
		if text, ok := chain[0].(*Text); ok {
			arr := strings.SplitN(strings.TrimSpace(text.Text), " ", 2)
			cmd = strings.TrimSpace(arr[0])
			if len(arr) > 1 {
				content = strings.TrimSpace(arr[1])
			}
		}
	}
	if len(cmd) == 0 {
		return true
	}
	if h, ok := cmdMap[cmd]; ok {
		if h.CheckAuth(message.GroupId, message.Sender.UserId) {
			groupMsg := h.Execute(message, content)
			if len(groupMsg) > 0 {
				sendGroupMessage(message, groupMsg...)
			}
			return true
		}
	}
	return true
}

func addCmdListener(handler iface.CmdHandler) {
	name := handler.Name()
	if _, ok := cmdMap[name]; ok {
		panic("repeat command: " + name)
	}
	cmdMap[name] = handler
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
	if changed {
		hkData.Set("pushedMessages", pushedMessages)
		if err = hkData.WriteConfig(); err != nil {
			slog.Error("write hkData failed", "error", err)
		}
	}
	if len(result1) == 0 {
		return
	}
	for _, groupId := range hkConfig.GetIntSlice("speedrun_push_qq_group") {
		_, err := B.SendGroupMessage(int64(groupId), MessageChain{&Text{Text: strings.Join(result1, "\n")}})
		if err != nil {
			slog.Error("send group message failed", "error", err)
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
