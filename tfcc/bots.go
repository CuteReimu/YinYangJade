package tfcc

import (
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/CuteReimu/threp"
	"github.com/go-resty/resty/v2"
	"io"
	"log/slog"
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
	})
}

var B *Bot

func Init(b *Bot) {
	initConfig()
	initBilibili()
	B = b
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(bilibiliAnalysis)
	B.ListenGroupMessage(repAnalysis)
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

func repAnalysis(message *GroupMessage) bool {
	if !slices.Contains(tfccConfig.GetIntSlice("qq.qq_group"), int(message.Sender.Group.Id)) {
		return true
	}
	for _, chain := range message.MessageChain {
		if file, ok := chain.(*File); ok && (strings.HasSuffix(file.Name, ".rpy") && file.Size <= 10*1024*1024) {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic recovered", "error", r)
					}
				}()
				info, err := B.GetFileInfo(FileParam{
					Id:               file.Id,
					Group:            message.Sender.Group.Id,
					WithDownloadInfo: true,
				})
				if err != nil || info.DownloadInfo == nil {
					slog.Error("获取文件信息失败", "error", err, "info", info)
					return
				}
				resp, err := restyClient.R().SetDoNotParseResponse(true).Get(info.DownloadInfo.Url)
				if err != nil {
					slog.Error("请求失败", "error", err)
					return
				}
				body := resp.RawBody()
				defer func(body io.ReadCloser) { _ = body.Close() }(body)
				replay, err := threp.DecodeReplay(body)
				if err != nil {
					slog.Error("解析失败", "error", err)
					return
				}
				_, err = B.SendGroupMessage(message.Sender.Group.Id, 0, MessageChain{&Plain{Text: replay.String()}})
				if err != nil {
					slog.Error("发送群消息失败", "error", err)
				}
			}()
		}
	}
	return true
}
