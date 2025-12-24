package tfcc

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/CuteReimu/bilibili/v2"
	. "github.com/CuteReimu/onebot"
	regexp "github.com/dlclark/regexp2"
	"github.com/pkg/errors"
)

var (
	avReg    = regexp.MustCompile(`^(?<![A-Za-z0-9])(?:https?://www\.bilibili\.com/video/)?av(\d+)$`, regexp.IgnoreCase)
	bvReg    = regexp.MustCompile(`^(?<![A-Za-z0-9])(?:https?://www\.bilibili\.com/video/|https?://b23\.tv)?bv([0-9A-Za-z]{10})$`, regexp.IgnoreCase) //nolint:revive
	liveReg  = regexp.MustCompile(`^(?<![A-Za-z0-9])https?://live\.bilibili\.com/(\d+)$`, regexp.IgnoreCase)
	shortReg = regexp.MustCompile(`^(?<![A-Za-z0-9])https?://b23\.tv/[0-9A-Za-z]{7}$`, regexp.IgnoreCase)
)

func bilibiliAnalysis(message *GroupMessage) bool {
	const (
		videoFormat = "%s\nhttps://www.bilibili.com/video/%s\nUP主：%s\n视频简介：%s"
		liveFormat  = "%s\nhttps://live.bilibili.com/%d\n人气：%d\n直播简介：%s"
	)
	if len(message.Message) != 1 {
		return true
	}
	test, ok := message.Message[0].(*Text)
	if !ok {
		return true
	}
	if !slices.Contains(tfccConfig.GetIntSlice("qq.qq_group"), int(message.GroupId)) {
		return true
	}
	content := strings.TrimSpace(test.Text)
	resultAny, found, err := getVideoInfo(content)
	if found {
		if err != nil {
			slog.Error("获取视频信息失败", "error", err)
			return true
		}
		var image *Image
		switch result := resultAny.(type) {
		case *bilibili.VideoInfo:
			if len(result.Pic) > 0 {
				image = &Image{File: result.Pic}
			}
			if len([]rune(result.Desc)) > 100 {
				result.Desc = string([]rune(result.Desc)[:100]) + "。。。"
			}
			test = &Text{Text: fmt.Sprintf(videoFormat, result.Title, result.Bvid, result.Owner.Name, result.Desc)}
		case *bilibili.LiveRoomInfo:
			if len(result.UserCover) > 0 {
				image = &Image{File: result.UserCover}
			}
			if len([]rune(result.Description)) > 100 {
				result.Description = string([]rune(result.Description)[:100]) + "。。。"
			}
			test = &Text{Text: fmt.Sprintf(liveFormat, result.Title, result.RoomId, result.Online, result.Description)}
		default:
			slog.Error("解析类型异常", "result", reflect.TypeOf(result))
			return true
		}
		_, err := bot.SendGroupMessage(message.GroupId, MessageChain{
			&Reply{Id: strconv.FormatInt(int64(message.MessageId), 10)}, image, test,
		})
		if err != nil {
			slog.Error("发送群消息失败", "error", err)
		}
	}
	return true
}

func getVideoInfo(content string) (any, bool, error) {
	if avRes, _ := avReg.FindStringMatch(content); avRes != nil {
		avid, err := strconv.Atoi(avRes.GroupByNumber(1).String())
		if err != nil {
			return nil, true, errors.Wrap(err, "解析avid失败："+avRes.GroupByNumber(1).String())
		}
		result, err := bili.GetVideoInfo(bilibili.VideoParam{Aid: avid})
		return result, true, err
	}
	if bvRes, _ := bvReg.FindStringMatch(content); bvRes != nil {
		result, err := bili.GetVideoInfo(bilibili.VideoParam{Bvid: bvRes.GroupByNumber(1).String()})
		return result, true, err
	}
	if liveRes, _ := liveReg.FindStringMatch(content); liveRes != nil {
		rid, err := strconv.Atoi(liveRes.GroupByNumber(1).String())
		if err != nil {
			return nil, true, errors.Wrap(err, "解析rid失败："+liveRes.GroupByNumber(1).String())
		}
		result, err := bili.GetLiveRoomInfo(bilibili.GetLiveRoomInfoParam{RoomId: rid})
		return result, true, err
	}
	if shortRes, _ := shortReg.FindStringMatch(content); shortRes != nil {
		typ, result, err := bili.UnwrapShortUrl(shortRes.String())
		if err != nil {
			return nil, true, err
		}
		if typ == "bvid" {
			bvid, ok := result.(string)
			if !ok {
				return nil, true, errors.New("invalid bvid type")
			}
			result, err := bili.GetVideoInfo(bilibili.VideoParam{Bvid: bvid})
			return result, true, err
		} else if typ == "live" {
			rid, ok := result.(int)
			if !ok {
				return nil, true, errors.New("invalid room id type")
			}
			result, err := bili.GetLiveRoomInfo(bilibili.GetLiveRoomInfoParam{RoomId: rid})
			return result, true, err
		}
		return result, true, err
	}
	return nil, false, nil
}
