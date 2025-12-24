package tfcc

import (
	"encoding/base64"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
)

var restyClient = resty.New()

func init() {
	restyClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	restyClient.SetTimeout(20 * time.Second)
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69", //nolint:revive
		"connection":   "close",
	})
}

var bot *Bot

// Init 初始化函数
func Init(b *Bot) {
	initConfig()
	initBilibili()
	bot = b
	bot.ListenGroupMessage(cmdHandleFunc)
	bot.ListenGroupMessage(bilibiliAnalysis)
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
	if len(cmd) == 0 || strings.Contains(content, "\n") || strings.Contains(content, "\r") {
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

func fillSpecificMessage(messages []SingleMessage) {
	for _, m := range messages {
		if img, ok := m.(*Image); ok && len(img.File) > 0 {
			if strings.HasPrefix(img.File, "file://") {
				fileName := img.File[len("file://"):]
				buf, err := os.ReadFile(fileName)
				if err != nil {
					slog.Error("read file failed", "error", err)
					continue
				}
				img.File = "base64://" + base64.StdEncoding.EncodeToString(buf)
			}
		} else if node, ok := m.(*Node); ok {
			fillSpecificMessage(node.Content)
		}
	}
}

func sendGroupMessage(context *GroupMessage, messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		fillSpecificMessage(messages)
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

func replyGroupMessage(context *GroupMessage, messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		fillSpecificMessage(messages)
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
