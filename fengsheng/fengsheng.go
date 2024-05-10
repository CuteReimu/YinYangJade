package fengsheng

import (
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/tidwall/gjson"
	"log/slog"
	"strings"
)

func init() {
	addCmdListener(&getMyScore{})
	addCmdListener(&getScore{})
	addCmdListener(&rankList{})
	addCmdListener(&winRate{})
	addCmdListener(&register{})
}

type getMyScore struct{}

func (g *getMyScore) Name() string {
	return "查询我"
}

func (g *getMyScore) ShowTips(_ int64, senderId int64) string {
	if permData.IsSet(fmt.Sprintf("playerMap.%d", senderId)) {
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
	name := permData.GetString(fmt.Sprintf("playerMap.%d", msg.Sender.Id))
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
	return MessageChain{&Plain{Text: resp.String()}}
}

type getScore struct{}

func (g *getScore) Name() string {
	return "查询"
}

func (g *getScore) ShowTips(_ int64, senderId int64) string {
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
	return MessageChain{&Plain{Text: resp.String()}}
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
	if !permData.IsSet(fmt.Sprintf("playerMap.%d", senderId)) {
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
	if oldName := permData.GetString(fmt.Sprintf("playerMap.%d", msg.Sender.Id)); len(oldName) > 0 {
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
	permData.Set(fmt.Sprintf("playerMap.%d", msg.Sender.Id), name)
	if err = permData.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	return MessageChain{&Plain{Text: "注册成功"}}
}
