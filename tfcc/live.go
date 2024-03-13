package tfcc

import (
	"fmt"
	"github.com/CuteReimu/YinYangJade/bots"
	"github.com/CuteReimu/YinYangJade/db"
	"github.com/CuteReimu/bilibili"
	"github.com/CuteReimu/dets"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/ozgio/strutil"
	"github.com/spf13/viper"
	"log/slog"
)

func init() {
	bots.AddCmdListener(&getLiveState{})
	bots.AddCmdListener(&startLive{})
	bots.AddCmdListener(&stopLive{})
	bots.AddCmdListener(&changeLiveTitle{})
	bots.AddCmdListener(&changeLiveArea{})
}

type getLiveState struct{}

func (g *getLiveState) Name() string {
	return "直播状态"
}

func (g *getLiveState) ShowTips(int64, int64) string {
	return "直播状态"
}

func (g *getLiveState) CheckAuth(int64, int64) bool {
	return true
}

func (g *getLiveState) Execute(_ *GroupMessage, content string) []SingleMessage {
	if len(content) != 0 {
		return nil
	}
	rid := viper.GetInt("tfcc.bilibili.room_id")
	ret, err := bilibili.GetRoomInfo(rid)
	if err != nil {
		slog.Error("获取直播状态失败", "error", err)
		return nil
	}
	var text string
	if ret.LiveStatus == 0 {
		text = "直播间状态：未开播"
	} else {
		text = fmt.Sprintf("直播间状态：开播\n直播标题：%s\n人气：%d\n直播间地址：%s", ret.Title, ret.Online, getLiveUrl())
	}
	return MessageChain(&Plain{Text: text})
}

type startLive struct{}

func (s *startLive) Name() string {
	return "开始直播"
}

func (s *startLive) ShowTips(int64, int64) string {
	return "开始直播"
}

func (s *startLive) CheckAuth(_ int64, senderId int64) bool {
	return IsWhitelist(senderId)
}

func (s *startLive) Execute(msg *GroupMessage, content string) []SingleMessage {
	if len(content) != 0 {
		return nil
	}
	rid := viper.GetInt("tfcc.bilibili.room_id")
	area := viper.GetInt("tfcc.bilibili.area_v2")
	ret, err := bilibili.StartLive(rid, area)
	if err != nil {
		slog.Error("开启直播间失败", "error", err)
		return nil
	}
	var publicText string
	if ret.Change == 0 {
		uin := dets.GetInt64([]byte("bilibili_live"))
		if uin != 0 {
			if uin != msg.Sender.Id {
				publicText = fmt.Sprintf("已经有人正在直播了\n直播间地址：%s\n快来围观吧！", getLiveUrl())
				return MessageChain(&Plain{Text: publicText})
			}
		} else {
			dets.Put([]byte("bilibili_live"), msg.Sender.Id)
		}
		publicText = fmt.Sprintf("直播间本来就是开启的\n直播间地址：%s\n快来围观吧！", getLiveUrl())
	} else {
		dets.Put([]byte("bilibili_live"), msg.Sender.Id)
		publicText = fmt.Sprintf("直播间已开启，别忘了修改直播间标题哦！\n直播间地址：%s\n快来围观吧！", getLiveUrl())
	}
	return MessageChain(&Plain{Text: publicText})
}

type stopLive struct{}

func (s *stopLive) Name() string {
	return "关闭直播"
}

func (s *stopLive) ShowTips(int64, int64) string {
	return "关闭直播"
}

func (s *stopLive) CheckAuth(_ int64, senderId int64) bool {
	return IsWhitelist(senderId)
}

func (s *stopLive) Execute(msg *GroupMessage, content string) []SingleMessage {
	if len(content) != 0 {
		return nil
	}
	if !IsAdmin(msg.Sender.Id) {
		uin := dets.GetInt64([]byte("bilibili_live"))
		if uin != 0 && uin != msg.Sender.Id {
			return MessageChain(&Plain{Text: "谢绝唐突关闭直播"})
		}
	}
	rid := viper.GetInt("tfcc.bilibili.room_id")
	changed, err := bilibili.StopLive(rid)
	if err != nil {
		slog.Error("关闭直播间失败", "error", err)
		return nil
	}
	db.Del([]byte("bilibili_live"))
	var text string
	if !changed {
		text = "直播间本来就是关闭的"
	} else {
		text = "直播间已关闭"
	}
	return MessageChain(&Plain{Text: text})
}

type changeLiveTitle struct{}

func (c *changeLiveTitle) Name() string {
	return "修改直播标题"
}

func (c *changeLiveTitle) ShowTips(int64, int64) string {
	return "修改直播标题 新标题"
}

func (c *changeLiveTitle) CheckAuth(_ int64, senderId int64) bool {
	return IsWhitelist(senderId)
}

func (c *changeLiveTitle) Execute(msg *GroupMessage, content string) []SingleMessage {
	if len(content) == 0 {
		return MessageChain(&Plain{Text: "指令格式如下：\n修改直播标题 新标题"})
	}
	if strutil.Len(content) > 20 {
		return nil
	}
	if !IsAdmin(msg.Sender.Id) {
		uin := dets.GetInt64([]byte("bilibili_live"))
		if uin != 0 && uin != msg.Sender.Id {
			return MessageChain(&Plain{Text: "谢绝唐突修改直播标题"})
		}
	}
	rid := viper.GetInt("tfcc.bilibili.room_id")
	err := bilibili.UpdateLive(rid, content)
	var text string
	if err != nil {
		slog.Error("修改直播间标题失败", "error", err)
		text = "修改直播间标题失败，请联系管理员"
	} else {
		text = "直播间标题已修改为：" + content
	}
	return MessageChain(&Plain{Text: text})
}

type changeLiveArea struct{}

func (c *changeLiveArea) Name() string {
	return "修改直播分区"
}

func (c *changeLiveArea) ShowTips(int64, int64) string {
	return "修改直播分区 新分区"
}

func (c *changeLiveArea) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (c *changeLiveArea) Execute(_ *GroupMessage, content string) []SingleMessage {
	if len(content) == 0 {
		return MessageChain(&Plain{Text: "指令格式如下：\n修改直播分区 新分区"})
	}
	// TODO bilibili.GetDynamicLiveUserList()
	return nil
}

func getLiveUrl() string {
	return "https://live.bilibili.com/" + viper.GetString("tfcc.bilibili.room_id")
}
