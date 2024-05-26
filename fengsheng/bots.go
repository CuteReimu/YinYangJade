package fengsheng

import (
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"
)

var B *Bot

func Init(b *Bot) {
	initConfig()
	B = b
	go func() {
		for range time.Tick(24 * time.Hour) {
			B.Run(clearExpiredImages)
		}
	}()
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(handleDictionary)
	B.ListenGroupMessage(searchAt)
}

var cmdMap = make(map[string]iface.CmdHandler)

func cmdHandleFunc(message *GroupMessage) bool {
	if !slices.Contains(fengshengConfig.GetIntSlice("qq.qq_group"), int(message.GroupId)) {
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
				replyGroupMessage(message, groupMsg...)
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
	if len(message.Message) == 0 {
		return true
	}
	if !slices.Contains(fengshengConfig.GetIntSlice("qq.qq_group"), int(message.GroupId)) {
		return true
	}
	if len(message.Message) >= 2 {
		if text, ok := message.Message[0].(*Text); ok && strings.TrimSpace(text.Text) == "查询" {
			if at, ok := message.Message[1].(*At); ok {
				data := permData.GetStringMapString("playerMap")
				name := data[at.QQ]
				if len(name) == 0 {
					replyGroupMessage(message, &Text{Text: "该玩家还未绑定"})
				} else {
					go func() {
						defer func() {
							if err := recover(); err != nil {
								slog.Error("panic recovered", "error", err)
							}
						}()
						result, returnError := httpGetString("/getscore", map[string]string{"name": name})
						if returnError != nil {
							slog.Error("请求失败", "error", returnError.error)
							replyGroupMessage(message, returnError.message...)
							return
						}
						replyGroupMessage(message, &Text{Text: result})
					}()
				}
			}
		}
	}
	return true
}
