package tfcc

import (
	"fmt"
	"github.com/CuteReimu/bilibili/v2"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/spf13/viper"
	"log/slog"
	"slices"
	"strings"
)

func init() {
	addCmdListener(&getLiveState{})
	addCmdListener(&startLive{})
	addCmdListener(&stopLive{})
	addCmdListener(&changeLiveTitle{})
	addCmdListener(&changeLiveArea{})
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
	rid := tfccConfig.GetInt("bilibili.room_id")
	ret, err := bili.GetLiveRoomInfo(bilibili.GetLiveRoomInfoParam{RoomId: rid})
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
	return []SingleMessage{&Plain{Text: text}}
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
	rid := tfccConfig.GetInt("bilibili.room_id")
	area := viper.GetInt("bilibili.area_v2")
	ret, err := bili.StartLive(bilibili.StartLiveParam{
		RoomId:   rid,
		AreaV2:   area,
		Platform: "pc",
	})
	if err != nil {
		slog.Error("开启直播间失败", "error", err)
		return []SingleMessage{&Plain{Text: "开始直播失败，" + err.Error()}}
	}
	var publicText string
	if ret.Change == 0 {
		uin := bilibiliData.GetInt64("live")
		if uin != 0 {
			if uin != msg.Sender.Id {
				publicText = fmt.Sprintf("已经有人正在直播了\n直播间地址：%s\n快来围观吧！", getLiveUrl())
				return []SingleMessage{&Plain{Text: publicText}}
			}
		} else {
			bilibiliData.Set("live", msg.Sender.Id)
			if err = bilibiliData.WriteConfig(); err != nil {
				slog.Error("write config failed", "error", err)
			}
		}
		publicText = fmt.Sprintf("直播间本来就是开启的\n直播间地址：%s\n快来围观吧！", getLiveUrl())
	} else {
		bilibiliData.Set("live", msg.Sender.Id)
		if err = bilibiliData.WriteConfig(); err != nil {
			slog.Error("write config failed", "error", err)
		}
		publicText = fmt.Sprintf("直播间已开启，别忘了修改直播间标题哦！\n直播间地址：%s\n快来围观吧！", getLiveUrl())
	}
	return []SingleMessage{&Plain{Text: publicText}}
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
		uin := bilibiliData.GetInt64("live")
		if uin != 0 && uin != msg.Sender.Id {
			return []SingleMessage{&Plain{Text: "谢绝唐突关闭直播"}}
		}
	}
	rid := tfccConfig.GetInt("bilibili.room_id")
	stopLiveResult, err := bili.StopLive(bilibili.StopLiveParam{
		RoomId: rid,
	})
	if err != nil {
		slog.Error("关闭直播间失败", "error", err)
		return nil
	}
	bilibiliData.Set("live", 0)
	if err = bilibiliData.WriteConfig(); err != nil {
		slog.Error("write config failed", "error", err)
	}
	var text string
	if stopLiveResult.Change == 0 {
		text = "直播间本来就是关闭的"
	} else {
		text = "直播间已关闭"
	}
	return []SingleMessage{&Plain{Text: text}}
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
		return []SingleMessage{&Plain{Text: "指令格式如下：\n修改直播标题 新标题"}}
	}
	if !IsAdmin(msg.Sender.Id) {
		uin := bilibiliData.GetInt64("live")
		if uin != 0 && uin != msg.Sender.Id {
			return []SingleMessage{&Plain{Text: "谢绝唐突修改直播标题"}}
		}
	}
	rid := viper.GetInt("tfcc.bilibili.room_id")
	err := bili.UpdateLiveRoomTitle(bilibili.UpdateLiveRoomTitleParam{
		RoomId: rid,
		Title:  content,
	})
	var text string
	if err != nil {
		slog.Error("修改直播间标题失败", "error", err)
		text = "修改直播间标题失败，请联系管理员"
	} else {
		text = "直播间标题已修改为：" + content
	}
	return []SingleMessage{&Plain{Text: text}}
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
		return []SingleMessage{&Plain{Text: "指令格式如下：\n修改直播分区 新分区"}}
	}
	name := strings.TrimSpace(content)
	areaList, err := bili.GetLiveAreaList()
	if err != nil {
		slog.Error("获取直播分区列表失败", "error", err)
		return []SingleMessage{&Plain{Text: "获取直播分区列表失败，" + err.Error()}}
	}
	index := slices.IndexFunc(areaList, func(area bilibili.LiveAreaList) bool {
		return area.Name == name
	})
	if index < 0 {
		return []SingleMessage{&Plain{Text: "没有这个分区"}}
	}
	tfccConfig.Set("bilibili.area_v2", areaList[index].Id)
	if err = tfccConfig.WriteConfig(); err != nil {
		slog.Error("write config failed", "error", err)
	}
	return []SingleMessage{&Plain{Text: "直播分区已修改为" + name}}
}

func getLiveUrl() string {
	return "https://live.bilibili.com/" + tfccConfig.GetString("bilibili.room_id")
}
