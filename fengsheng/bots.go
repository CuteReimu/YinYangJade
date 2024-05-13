package fengsheng

import (
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"log/slog"
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
	})
}

var B *Bot

func Init(b *Bot) {
	initConfig()
	B = b
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(handleDictionary)
	B.ListenGroupMessage(searchAt)
}

var cmdMap = make(map[string]iface.CmdHandler)

func cmdHandleFunc(message *GroupMessage) bool {
	if !slices.Contains(fengshengConfig.GetIntSlice("qq.qq_group"), int(message.Sender.Group.Id)) {
		return true
	}
	chain := message.MessageChain
	if len(chain) == 0 {
		return true
	}
	if _, ok := chain[0].(*Source); ok {
		chain = chain[1:]
		if len(chain) == 0 {
			return true
		}
	}
	if at, ok := chain[0].(*At); ok && at.Target == B.QQ {
		chain = chain[1:]
		if len(chain) > 0 {
			if plain, ok := chain[0].(*Plain); ok && len(strings.TrimSpace(plain.Text)) == 0 {
				chain = chain[1:]
			}
		}
		if len(chain) == 0 {
			chain = append(chain, &Plain{Text: "查看帮助"})
		}
	}
	var cmd, content string
	if len(chain) == 1 {
		if plain, ok := chain[0].(*Plain); ok {
			arr := strings.SplitN(strings.TrimSpace(plain.Text), " ", 2)
			cmd = strings.TrimSpace(arr[0])
			if len(arr) > 1 {
				content = strings.TrimSpace(arr[1])
			}
		}
	}
	if len(cmd) == 0 || strings.Contains(content, "\n") || strings.Contains(content, "\r") {
		return true
	}
	if h, ok := cmdMap[cmd]; ok {
		if h.CheckAuth(message.Sender.Group.Id, message.Sender.Id) {
			groupMsg := h.Execute(message, content)
			if len(groupMsg) > 0 {
				_, err := B.SendGroupMessage(message.Sender.Group.Id, 0, groupMsg)
				if err != nil {
					slog.Error("发送群消息失败", "error", err)
				}
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

func searchAt(message *GroupMessage) bool {
	if len(message.MessageChain) <= 1 {
		return true
	}
	if !slices.Contains(fengshengConfig.GetIntSlice("qq.qq_group"), int(message.Sender.Group.Id)) {
		return true
	}
	if len(message.MessageChain) >= 3 {
		if plain, ok := message.MessageChain[1].(*Plain); ok && strings.TrimSpace(plain.Text) == "查询" {
			if at, ok := message.MessageChain[2].(*At); ok {
				data := permData.GetStringMapString("playerMap")
				name := data[strconv.FormatInt(at.Target, 10)]
				if len(name) == 0 {
					sendGroupMessage(message.Sender.Group.Id, &Plain{Text: "该玩家还未绑定"})
				} else {
					go func() {
						defer func() {
							if err := recover(); err != nil {
								slog.Error("panic recovered", "error", err)
							}
						}()
						resp, err := restyClient.R().SetQueryParam("name", name).Get(fengshengConfig.GetString("fengshengUrl") + "/getscore")
						if err != nil {
							slog.Error("请求失败", "error", err)
							return
						}
						if resp.StatusCode() != 200 {
							slog.Error("请求失败", "status", resp.StatusCode())
							return
						}
						result := gjson.GetBytes(resp.Body(), "result").String()
						if len(result) == 0 {
							return
						}
						sendGroupMessage(message.Sender.Group.Id, &Plain{Text: result})
					}()
				}
			}
		}
	}
	return true
}
