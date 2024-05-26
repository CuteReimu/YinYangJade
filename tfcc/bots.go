package tfcc

import (
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
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
	initBilibili()
	B = b
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(bilibiliAnalysis)
}

var cmdMap = make(map[string]iface.CmdHandler)

func cmdHandleFunc(message *GroupMessage) bool {
	if !slices.Contains(tfccConfig.GetIntSlice("qq.qq_group"), int(message.GroupId)) {
		return true
	}
	chain := message.Message
	if len(chain) == 0 {
		return true
	}
	if at, ok := chain[0].(*At); ok && at.QQ == strconv.FormatInt(B.QQ, 10) {
		chain = chain[1:]
		if len(chain) > 0 {
			if Text, ok := chain[0].(*Text); ok && len(strings.TrimSpace(Text.Text)) == 0 {
				chain = chain[1:]
			}
		}
		if len(chain) == 0 {
			chain = append(chain, &Text{Text: "查看帮助"})
		}
	}
	var cmd, content string
	if len(chain) == 1 {
		if Text, ok := chain[0].(*Text); ok {
			arr := strings.SplitN(strings.TrimSpace(Text.Text), " ", 2)
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

func replyGroupMessage(context *GroupMessage, messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		if slices.ContainsFunc(messages, func(message SingleMessage) bool {
			switch message.(type) {
			case *Text, *Image, *Face:
				return false
			}
			return true
		}) {
			_, err := B.SendGroupMessage(context.GroupId, messages)
			return err
		}
		return context.Reply(B, messages, true)
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
