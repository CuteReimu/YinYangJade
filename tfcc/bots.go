package tfcc

import (
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/mirai-sdk-http"
	"log/slog"
	"slices"
	"strings"
)

var B *Bot

func Init(b *Bot) {
	initConfig()
	initBilibili()
	B = b
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(bilibiliAnalysis)
}

var cmdMap = make(map[string]iface.CmdHandler)

func cmdHandleFunc(message *GroupMessage) bool {
	if !slices.Contains(tfccConfig.GetIntSlice("qq.qq_group"), int(message.Sender.Group.Id)) {
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
