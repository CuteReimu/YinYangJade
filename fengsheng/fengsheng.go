package fengsheng

import (
	"encoding/base64"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/tidwall/gjson"
	"log/slog"
	"strconv"
	"strings"
)

func init() {
	addCmdListener(&getMyScore{})
	addCmdListener(&getScore{})
	addCmdListener(&rankList{})
	addCmdListener(&winRate{})
	addCmdListener(&register{})
	addCmdListener(&addNotifyOnStart{})
	addCmdListener(&addNotifyOnEnd{})
	addCmdListener(&atPlayer{})
	addCmdListener(&updateTitle{})
}

type getMyScore struct{}

func (g *getMyScore) Name() string {
	return "查询我"
}

func (g *getMyScore) ShowTips(_ int64, senderId int64) string {
	data := permData.GetStringMapString("playerMap")
	if _, ok := data[strconv.FormatInt(senderId, 10)]; ok {
		return "查询我"
	}
	return ""
}

func (g *getMyScore) CheckAuth(int64, int64) bool {
	return true
}

func (g *getMyScore) Execute(msg *GroupMessage, content string) MessageChain {
	content = strings.TrimSpace(content)
	if len(content) > 0 {
		return nil
	}
	data := permData.GetStringMapString("playerMap")
	name := data[strconv.FormatInt(msg.Sender.Id, 10)]
	if len(name) == 0 {
		return MessageChain{&Plain{Text: "请先绑定"}}
	}
	resp, err := restyClient.R().SetQueryParam("name", name).Get(fengshengConfig.GetString("fengshengUrl") + "/getscore")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	result := gjson.GetBytes(resp.Body(), "result").String()
	if len(result) == 0 {
		return nil
	}
	return MessageChain{&Plain{Text: result}}
}

type getScore struct{}

func (g *getScore) Name() string {
	return "查询"
}

func (g *getScore) ShowTips(int64, int64) string {
	return "查询 名字"
}

func (g *getScore) CheckAuth(int64, int64) bool {
	return true
}

func (g *getScore) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return nil
	}
	resp, err := restyClient.R().SetQueryParam("name", name).Get(fengshengConfig.GetString("fengshengUrl") + "/getscore")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	result := gjson.GetBytes(resp.Body(), "result").String()
	if len(result) == 0 {
		return nil
	}
	return MessageChain{&Plain{Text: result}}
}

type rankList struct{}

func (r *rankList) Name() string {
	return "排行"
}

func (r *rankList) ShowTips(int64, int64) string {
	return "排行"
}

func (r *rankList) CheckAuth(int64, int64) bool {
	return true
}

func (r *rankList) Execute(_ *GroupMessage, content string) MessageChain {
	content = strings.TrimSpace(content)
	if len(content) > 0 {
		return nil
	}
	resp, err := restyClient.R().Get(fengshengConfig.GetString("fengshengUrl") + "/ranklist")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	return MessageChain{&Image{Base64: base64.StdEncoding.EncodeToString(resp.Body())}}
}

type winRate struct{}

func (r *winRate) Name() string {
	return "胜率"
}

func (r *winRate) ShowTips(int64, int64) string {
	return "胜率"
}

func (r *winRate) CheckAuth(int64, int64) bool {
	return true
}

func (r *winRate) Execute(_ *GroupMessage, content string) MessageChain {
	content = strings.TrimSpace(content)
	if len(content) > 0 {
		return nil
	}
	resp, err := restyClient.R().Get(fengshengConfig.GetString("fengshengUrl") + "/winrate")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	return MessageChain{&Image{Base64: base64.StdEncoding.EncodeToString(resp.Body())}}
}

type register struct{}

func (r *register) Name() string {
	return "注册"
}

func (r *register) ShowTips(_ int64, senderId int64) string {
	data := permData.GetStringMapString("playerMap")
	if _, ok := data[strconv.FormatInt(senderId, 10)]; !ok {
		return "注册 名字"
	}
	return ""
}

func (r *register) CheckAuth(int64, int64) bool {
	return true
}

func (r *register) Execute(msg *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Plain{Text: "命令格式：\n注册 名字"}}
	}
	data := permData.GetStringMapString("playerMap")
	if oldName := data[strconv.FormatInt(msg.Sender.Id, 10)]; len(oldName) > 0 {
		return MessageChain{&Plain{Text: "你已经注册过：" + oldName}}
	}
	resp, err := restyClient.R().SetQueryParam("name", name).Get(fengshengConfig.GetString("fengshengUrl") + "/register")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	body := resp.Body()
	if !gjson.GetBytes(body, "result").Bool() {
		msg := gjson.GetBytes(body, "error").String()
		if len(msg) == 0 {
			msg = "用户名重复"
		}
		return MessageChain{&Plain{Text: msg}}
	}
	data[strconv.FormatInt(msg.Sender.Id, 10)] = name
	permData.Set("playerMap", data)
	if err = permData.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	return MessageChain{&Plain{Text: "注册成功"}}
}

type addNotifyOnStart struct{}

func (a *addNotifyOnStart) Name() string {
	return "开了喊我"
}

func (a *addNotifyOnStart) ShowTips(int64, int64) string {
	return "开了喊我"
}

func (a *addNotifyOnStart) CheckAuth(int64, int64) bool {
	return true
}

func (a *addNotifyOnStart) Execute(msg *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	resp, err := restyClient.R().SetQueryParam("qq", strconv.FormatInt(msg.Sender.Id, 10)).
		Get(fengshengConfig.GetString("fengshengUrl") + "/addnotify")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	body := resp.Body()
	if !gjson.GetBytes(body, "result").Bool() {
		msg := gjson.GetBytes(body, "error").String()
		if len(msg) == 0 {
			msg = "太多人预约了，不能再添加了"
		}
		return MessageChain{&Plain{Text: msg}}
	}
	return MessageChain{&Plain{Text: "好的，开了喊你"}}
}

type addNotifyOnEnd struct{}

func (a *addNotifyOnEnd) Name() string {
	return "结束喊我"
}

func (a *addNotifyOnEnd) ShowTips(int64, int64) string {
	return "结束喊我"
}

func (a *addNotifyOnEnd) CheckAuth(int64, int64) bool {
	return true
}

func (a *addNotifyOnEnd) Execute(msg *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	resp, err := restyClient.R().SetQueryParam("qq", strconv.FormatInt(msg.Sender.Id, 10)).
		SetQueryParam("when", "1").
		Get(fengshengConfig.GetString("fengshengUrl") + "/addnotify")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	body := resp.Body()
	if !gjson.GetBytes(body, "result").Bool() {
		msg := gjson.GetBytes(body, "error").String()
		if len(msg) == 0 {
			msg = "太多人预约了，不能再添加了"
		}
		return MessageChain{&Plain{Text: msg}}
	}
	return MessageChain{&Plain{Text: "好的，结束喊你"}}
}

type atPlayer struct{}

func (a *atPlayer) Name() string {
	return "艾特"
}

func (a *atPlayer) ShowTips(int64, int64) string {
	return "艾特 游戏内的名字"
}

func (a *atPlayer) CheckAuth(int64, int64) bool {
	return true
}

func (a *atPlayer) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Plain{Text: "命令格式：\n艾特 游戏内的名字"}}
	}
	data := permData.GetStringMapString("playerMap")
	for id, v := range data {
		if v == content {
			qq, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				slog.Error("parse int failed: " + id)
				return nil
			}
			return MessageChain{&At{Target: qq}}
		}
	}
	return MessageChain{&Plain{Text: "没能找到此玩家，可能还未绑定"}}
}

type updateTitle struct{}

func (u *updateTitle) Name() string {
	return "修改称号"
}

func (u *updateTitle) ShowTips(_ int64, senderId int64) string {
	data := permData.GetStringMapString("playerMap")
	if _, ok := data[strconv.FormatInt(senderId, 10)]; ok {
		return "修改称号 称号"
	}
	return ""
}

func (u *updateTitle) CheckAuth(int64, int64) bool {
	return true
}

func (u *updateTitle) Execute(msg *GroupMessage, content string) MessageChain {
	title := strings.TrimSpace(content)
	if len(title) == 0 {
		return MessageChain{&Plain{Text: "命令格式：\n修改称号 称号"}}
	}
	data := permData.GetStringMapString("playerMap")
	name := data[strconv.FormatInt(msg.Sender.Id, 10)]
	if len(name) == 0 {
		return MessageChain{&Plain{Text: "请先注册"}}
	}
	resp, err := restyClient.R().SetQueryParam("name", name).
		SetQueryParam("title", title).
		Get(fengshengConfig.GetString("fengshengUrl") + "/updatetitle")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	body := resp.Body()
	if !gjson.GetBytes(body, "result").Bool() {
		msg := gjson.GetBytes(body, "error").String()
		if len(msg) == 0 {
			msg = "你的段位太低，请提升段位后再来使用此功能"
		}
		return MessageChain{&Plain{Text: msg}}
	}
	return MessageChain{&Plain{Text: "修改称号成功"}}
}
