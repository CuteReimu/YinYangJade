// Package botutil provides common utilities for bot modules, including
// message handling, dictionary operations, and image management.
package botutil

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
)

// CmdHandleFunc 处理命令的通用逻辑
func CmdHandleFunc(
	message *GroupMessage,
	bot *Bot,
	validGroups []int,
	cmdMap map[string]iface.CmdHandler,
	sendFunc func(*GroupMessage, ...SingleMessage),
) bool {
	if !slices.Contains(validGroups, int(message.GroupId)) {
		return true
	}
	chain := message.Message
	if len(chain) == 0 {
		return true
	}
	if at, ok := chain[0].(*At); ok && at.QQ == strconv.FormatInt(bot.QQ, 10) {
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
				sendFunc(message, groupMsg...)
			}
			return true
		}
	}
	return true
}

// AddCmdListener 添加命令监听器
func AddCmdListener(handler iface.CmdHandler, cmdMap map[string]iface.CmdHandler) {
	name := handler.Name()
	if _, ok := cmdMap[name]; ok {
		panic("repeat command: " + name)
	}
	cmdMap[name] = handler
}

// SendGroupMessageWithRetry 发送群消息，失败时自动过滤http图片重试
func SendGroupMessageWithRetry(bot *Bot, context *GroupMessage, fillFunc func([]SingleMessage), messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		fillFunc(messages)
		_, err := bot.SendGroupMessage(context.GroupId, messages)
		return err
	}
	if err := f(messages); err != nil {
		slog.Error("send group message failed", "error", err)
		newMessages := make([]SingleMessage, 0, len(messages))
		for _, m := range messages {
			if image, ok := m.(*Image); !ok || !strings.HasPrefix(image.File, "http") {
				newMessages = append(newMessages, m)
			}
		}
		if len(newMessages) != len(messages) && len(newMessages) > 0 {
			if err = f(newMessages); err != nil {
				slog.Error("send group message failed", "error", err)
			}
		}
	}
}

// ReplyGroupMessageWithRetry 回复群消息，失败时自动过滤http图片重试
func ReplyGroupMessageWithRetry(bot *Bot, context *GroupMessage, fillFunc func([]SingleMessage), messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		fillFunc(messages)
		_, err := bot.SendGroupMessage(context.GroupId, append(MessageChain{
			&Reply{Id: strconv.FormatInt(int64(context.MessageId), 10)},
		}, messages...))
		return err
	}
	if err := f(messages); err != nil {
		slog.Error("send group message failed", "error", err)
		newMessages := make([]SingleMessage, 0, len(messages))
		for _, m := range messages {
			if image, ok := m.(*Image); !ok || !strings.HasPrefix(image.File, "http") {
				newMessages = append(newMessages, m)
			}
		}
		if len(newMessages) != len(messages) && len(newMessages) > 0 {
			if err = f(newMessages); err != nil {
				slog.Error("send group message failed", "error", err)
			}
		}
	}
}
