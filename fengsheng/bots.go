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
	B.ListenGroupRequest(handleGroupRequest)
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
					sendGroupMessage(message, &Text{Text: "该玩家还未绑定"})
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
							sendGroupMessage(message, returnError.message...)
							return
						}
						sendGroupMessage(message, dealGetScore(result)...)
					}()
				}
			}
		}
	}
	return true
}

var approveGroupRequestStrings = []string{"抖音", "b站", "小红书", "贴吧", "搜的", "快手", "github"}

func handleGroupRequest(request *GroupRequest) bool {
	if request.SubType == GroupRequestAdd {
		if !slices.Contains(fengshengConfig.GetIntSlice("qq.qq_group"), int(request.GroupId)) {
			return true
		}
		if strings.Contains(request.Comment, "管理员你好") {
			return true
		}
		comment := request.Comment
		if index := strings.Index(comment, "\n"); index >= 0 {
			comment = comment[index+1:]
			comment = strings.TrimPrefix(comment, "答案：")
		}
		comment = strings.ToLower(comment)
		if slices.ContainsFunc(approveGroupRequestStrings, func(s string) bool {
			return (strings.HasPrefix(comment, s) || strings.HasSuffix(comment, s)) && len(comment) < 15
		}) {
			err := B.SetGroupAddRequest(request.Flag, request.SubType, true, "")
			if err != nil {
				slog.Error("同意申请请求失败", "approve", true, "error", err)
			} else {
				slog.Info("同意申请请求成功", "approve", true, "error", err)
			}
		}
	}
	return true
}
