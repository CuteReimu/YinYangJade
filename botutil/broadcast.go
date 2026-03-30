package botutil

import (
	"log/slog"
	"slices"
	"time"

	. "github.com/CuteReimu/onebot"
)

var broadcasting bool

var broadcastGroups []int64

// AddBroadcastGroup 添加广播群，广播消息会发送到这些群里
func AddBroadcastGroup(groupIDs []int) {
	var count int
	for _, id := range groupIDs {
		if !slices.Contains(broadcastGroups, int64(id)) {
			broadcastGroups = append(broadcastGroups, int64(id))
			count++
		}
	}
	slog.Info("添加广播群成功", "count", count, "total", len(broadcastGroups))
}

// HandleBroadcastRequest 处理广播请求，管理员发送"广播"后，下一条消息会被当作广播内容发送到所有广播群
func HandleBroadcastRequest(b *Bot, message *PrivateMessage) {
	if len(message.Message) == 1 {
		if text, ok := message.Message[0].(*Text); ok && text.Text == "广播" {
			_, _ = b.SendPrivateMessage(message.Sender.UserId, MessageChain{&Text{Text: "请输入要广播的消息内容，至少10字以上才有效"}})
			broadcasting = true
			return
		}
	}
	if !broadcasting {
		return
	}
	broadcasting = false
	var textCount int
	for _, c := range message.Message {
		if text, ok := c.(*Text); ok {
			textCount += len(text.Text)
		}
	}
	if textCount < 10 {
		_, _ = b.SendPrivateMessage(message.Sender.UserId, MessageChain{&Text{Text: "消息内容太短了，广播失败"}})
		return
	}
	for _, groupID := range broadcastGroups {
		_, err := b.SendGroupMessage(groupID, message.Message)
		if err == nil {
			slog.Info("广播消息成功", "groupId", groupID)
		} else {
			slog.Error("广播消息失败", "groupId", groupID, "error", err)
		}
		time.Sleep(3 * time.Second)
	}
}
