package tfcc

import (
	. "github.com/CuteReimu/mirai-sdk-http"
	"log/slog"
	"slices"
	"strings"
)

var B *Bot

// CmdHandler 这是聊天指令处理器的接口，当你想要新增自己的聊天指令处理器时，实现这个接口即可
type CmdHandler interface {
	// Name 群友输入聊天指令时，第一个空格前的内容。
	Name() string
	// ShowTips 在【帮助列表】中应该如何显示这个命令。空字符串表示不显示
	ShowTips(groupCode int64, senderId int64) string
	// CheckAuth 如果他有权限执行这个指令，则返回True，否则返回False
	CheckAuth(groupCode int64, senderId int64) bool
	// Execute content参数是除开指令名（第一个空格前的部分）以外剩下的所有内容。返回值是要发送的群聊消息为空就是不发送消息。
	Execute(msg *GroupMessage, content string) []SingleMessage
}

func Init(b *Bot) {
	initConfig()
	initBilibili()
	B = b
	B.ListenGroupMessage(cmdHandleFunc)
}

var cmdMap map[string]CmdHandler

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
		if len(chain) > 1 {
			chain = chain[1:]
		} else {
			chain[0] = &Plain{Text: "查看帮助"}
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

func addCmdListener(handler CmdHandler) {
	name := handler.Name()
	if _, ok := cmdMap[name]; ok {
		panic("repeat command: " + name)
	}
	cmdMap[name] = handler
}
