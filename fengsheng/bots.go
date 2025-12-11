package fengsheng

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
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
	B.ListenPrivateMessage(handlePrivateRequest)
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
				operatorName := data[strconv.FormatInt(message.Sender.UserId, 10)]
				if len(operatorName) == 0 {
					sendGroupMessage(message, &Text{Text: getScoreFailResponse[rand.IntN(len(getScoreFailResponse))]})
					return true
				}
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
						result, returnError := httpGetString("/getscore", map[string]string{"name": name, "operator": operatorName})
						if returnError != nil {
							slog.Error("请求失败", "error", returnError.error)
							sendGroupMessage(message, returnError.message...)
							return
						}
						if result == "差距太大，无法查询" {
							sendGroupMessage(message, &Text{Text: getScoreFailResponse[rand.IntN(len(getScoreFailResponse))]})
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

var approveGroupRequestStrings = []string{"抖音", "b站", "bilibili", "哔哩哔哩", "小红书", "百度贴吧", "贴吧", "搜的", "快手", "github"}

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

func handlePrivateRequest(request *PrivateMessage) bool {
	if IsSuperAdmin(request.Sender.UserId) && len(request.Message) == 1 {
		if text, ok := request.Message[0].(*Text); ok && text.Text == "管理" {
			qqGroups := fengshengConfig.GetIntSlice("qq.qq_group")
			if len(qqGroups) == 0 {
				return true
			}
			qqGroup := int64(qqGroups[0])
			data := permData.GetStringMapString("playerMap")
			reverseMap := make(map[string]int64, len(data))
			for qq, name := range data {
				qqInt, err := strconv.ParseInt(qq, 10, 64)
				if err != nil {
					slog.Error("QQ号转换失败", "qq", qq, "error", err)
					sendPrivateMessage(request.UserId, MessageChain{&Text{Text: "格式错误"}})
					continue
				}
				reverseMap[name] = qqInt
			}
			result, returnError := httpGetString("/ranklist2", nil)
			if returnError != nil {
				slog.Error("请求失败", "error", returnError.error)
				sendPrivateMessage(request.UserId, returnError.message)
				return true
			}
			infos, err := B.GetGroupMemberList(qqGroup)
			if err != nil {
				slog.Error("获取群成员列表失败", "group", qqGroup, "error", err)
				sendPrivateMessage(request.UserId, MessageChain{&Text{Text: "获取群成员列表失败"}})
				return true
			}
			oldAdmins := make(map[int64]struct{}, 20)
			memberMap := make(map[int64]bool)
			memberName := make(map[int64]string)
			for _, info := range infos {
				memberMap[info.UserId] = info.Role == RoleOwner || info.Role == RoleAdmin
				memberName[info.UserId] = info.CardOrNickname()
				if (info.Role == RoleOwner || info.Role == RoleAdmin) && info.UserId != B.QQ {
					oldAdmins[info.UserId] = struct{}{}
				}
			}
			var adminCount int
			var newAdmin, removeAdmin []string
			for line := range strings.SplitSeq(result, "\n") {
				if adminCount >= 20 {
					break
				}
				arr := strings.SplitN(line, "：", 2)
				if len(arr) != 2 {
					slog.Error("格式错误", "line", line)
					sendPrivateMessage(request.UserId, MessageChain{&Text{Text: "格式错误"}})
					return true
				}
				line = arr[1]
				arr = strings.SplitN(line, "·", 2)
				if len(arr) != 2 {
					slog.Error("格式错误", "line", line)
					sendPrivateMessage(request.UserId, MessageChain{&Text{Text: "格式错误"}})
					return true
				}
				name := arr[0]
				qq := reverseMap[name]
				if qq == 0 {
					slog.Error("该玩家未绑定", "name", name)
					continue
				}
				isAdmin, isMember := memberMap[qq]
				if !isMember {
					continue
				}
				adminCount++
				delete(oldAdmins, qq)
				if !isAdmin {
					newAdmin = append(newAdmin, fmt.Sprintf("%d(%s)", qq, name))
				}
			}
			for qq := range oldAdmins {
				removeAdmin = append(removeAdmin, fmt.Sprintf("%d(%s)", qq, memberName[qq]))
			}
			if len(oldAdmins) > 0 || len(removeAdmin) > 0 {
				sendPrivateMessage(request.UserId, MessageChain{&Text{Text: "需要增加的管理员：\r\n" + strings.Join(newAdmin, "，") + "\r\n需要移除的管理员：\r\n" + strings.Join(removeAdmin, "，")}})
			} else {
				sendPrivateMessage(request.UserId, MessageChain{&Text{Text: "无需调整"}})
			}
		}
	}
	return true
}

func sendPrivateMessage(userId int64, messages MessageChain) {
	_, err := B.SendPrivateMessage(userId, messages)
	if err != nil {
		slog.Error("发送私聊消息失败", "error", err)
	}
}
