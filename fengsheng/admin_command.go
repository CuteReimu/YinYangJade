package fengsheng

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	. "github.com/CuteReimu/onebot"
)

func init() {
	addCmdListener(bind{})
	addCmdListener(unbind{})
	addCmdListener(unbindExpired{})
	addCmdListener(forbidPlayer{})
	addCmdListener(releasePlayer{})
	addCmdListener(forbidRole{})
	addCmdListener(releaseRole{})
	addCmdListener(setVersion{})
	addCmdListener(forceEnd{})
	addCmdListener(setNotice{})
	addCmdListener(setWaitSecond{})
	addCmdListener(createAccount{})
	addCmdListener(addEnergy{})
}

type bind struct{}

func (bind) Name() string {
	return "绑定"
}

func (bind) ShowTips(int64, int64) string {
	return "绑定 QQ号 名字"
}

func (bind) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (bind) Execute(msg *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	arr := strings.SplitN(name, " ", 2)
	if len(arr) != 2 {
		return MessageChain{&Text{Text: "命令格式：\n绑定 QQ号 名字"}}
	}
	qq, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil {
		return MessageChain{&Text{Text: "命令格式：\n绑定 QQ号 名字"}}
	}
	if _, err = B.GetGroupMemberInfo(msg.GroupId, qq, false); err != nil {
		return MessageChain{&Text{Text: fmt.Sprintf("%d不在群里", qq)}}
	}
	name = strings.TrimSpace(arr[1])
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n绑定 QQ号 名字"}}
	}
	data := permData.GetStringMapString("playerMap")
	id := strconv.FormatInt(qq, 10)
	for id0, name0 := range data {
		if id == id0 {
			return MessageChain{&Text{Text: "不能重复绑定"}}
		}
		if name == name0 {
			qq0, _ := strconv.ParseInt(id0, 10, 64)
			memberInfo, _ := B.GetGroupMemberInfo(msg.GroupId, qq0, false)
			s := "该玩家已被" + id0
			if memberInfo != nil {
				s += fmt.Sprintf("（%s<%d>）", memberInfo.CardOrNickname(), qq0)
			} else {
				s += fmt.Sprintf("（<%d>）", qq0)
			}
			s += "绑定"
			return MessageChain{&Text{Text: s}}
		}
	}
	result, returnError := httpClient.HTTPGetString("/getscore", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	if strings.HasSuffix(result, "已身死道消") {
		return MessageChain{&Text{Text: "不存在的玩家"}}
	}
	data[id] = name
	permData.Set("playerMap", data)
	if err = permData.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	return MessageChain{&Text{Text: "绑定成功"}}
}

type unbind struct{}

func (unbind) Name() string {
	return "解绑"
}

func (unbind) ShowTips(int64, int64) string {
	return "解绑 QQ号"
}

func (unbind) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (unbind) Execute(_ *GroupMessage, content string) MessageChain {
	id := strings.TrimSpace(content)
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return MessageChain{&Text{Text: "命令格式：\n解绑 QQ号"}}
	}
	data := permData.GetStringMapString("playerMap")
	oldLen := len(data)
	delete(data, id)
	if len(data) == oldLen {
		return MessageChain{&Text{Text: "玩家没有绑定"}}
	}
	permData.Set("playerMap", data)
	if err := permData.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	return MessageChain{&Text{Text: "解绑成功"}}
}

type unbindExpired struct{}

func (unbindExpired) Name() string {
	return "解绑所有0分玩家"
}

func (unbindExpired) ShowTips(int64, int64) string {
	return "解绑所有0分玩家"
}

func (unbindExpired) CheckAuth(_ int64, senderID int64) bool {
	return isSuperAdmin(senderID)
}

func (unbindExpired) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	m1 := permData.GetStringMapString("playerMap")
	m2 := make(map[string]string, len(m1))
	for id, name := range m1 {
		score, returnError := httpClient.HTTPGetString("/getscore", map[string]string{"name": name})
		if returnError != nil {
			slog.Error("请求失败", "error", returnError)
			return returnError.Message
		}
		if !strings.HasSuffix(score, "已身死道消") {
			m2[id] = name
		}
	}
	permData.Set("playerMap", m2)
	if err := permData.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	return MessageChain{&Text{Text: "解绑成功"}}
}

type forbidPlayer struct{}

func (forbidPlayer) Name() string {
	return "封号"
}

func (forbidPlayer) ShowTips(int64, int64) string {
	return "封号 名字 小时"
}

func (forbidPlayer) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (forbidPlayer) Execute(_ *GroupMessage, content string) MessageChain {
	c := strings.TrimSpace(content)
	if len(c) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n封号 名字 小时"}}
	}
	spaceIndex := strings.LastIndex(c, " ")
	if spaceIndex == -1 {
		return MessageChain{&Text{Text: "命令格式：\n封号 名字 小时"}}
	}
	name := strings.TrimSpace(c[:spaceIndex])
	if _, err := strconv.ParseInt(strings.TrimSpace(c[spaceIndex+1:]), 10, 64); err != nil {
		return MessageChain{&Text{Text: "命令格式：\n封号 名字 小时"}}
	}
	result, returnError := httpClient.HTTPGetString("/forbidplayer", map[string]string{
		"name": name,
		"hour": c[spaceIndex+1:],
	})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	return MessageChain{&Text{Text: result}}
}

type releasePlayer struct{}

func (releasePlayer) Name() string {
	return "解封"
}

func (releasePlayer) ShowTips(int64, int64) string {
	return "解封 名字"
}

func (releasePlayer) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (releasePlayer) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n解封 名字"}}
	}
	result, returnError := httpClient.HTTPGetString("/releaseplayer", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	return MessageChain{&Text{Text: result}}
}

type forbidRole struct{}

func (forbidRole) Name() string {
	return "禁用角色"
}

func (forbidRole) ShowTips(int64, int64) string {
	return "禁用角色 名字"
}

func (forbidRole) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (forbidRole) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n禁用角色 名字"}}
	}
	result, returnError := httpClient.HTTPGetBool("/forbidrole", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	if result {
		return MessageChain{&Text{Text: "禁用成功"}}
	}
	return MessageChain{&Text{Text: "禁用失败"}}
}

type releaseRole struct{}

func (releaseRole) Name() string {
	return "启用角色"
}

func (releaseRole) ShowTips(int64, int64) string {
	return "启用角色 名字"
}

func (releaseRole) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (releaseRole) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n启用角色 名字"}}
	}
	result, returnError := httpClient.HTTPGetBool("/releaserole", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	if result {
		return MessageChain{&Text{Text: "启用成功"}}
	}
	return MessageChain{&Text{Text: "启用失败"}}
}

type setVersion struct{}

func (setVersion) Name() string {
	return "修改版本号"
}

func (setVersion) ShowTips(int64, int64) string {
	return "修改版本号 版本号"
}

func (setVersion) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (setVersion) Execute(_ *GroupMessage, content string) MessageChain {
	if _, err := strconv.Atoi(content); err != nil {
		return MessageChain{&Text{Text: "命令格式：\n修改版本号 版本号"}}
	}
	returnError := httpClient.HTTPGet("/setversion", map[string]string{"version": content})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	return MessageChain{&Text{Text: "版本号已修改为" + content}}
}

type forceEnd struct{}

func (forceEnd) Name() string {
	return "强制结束所有游戏"
}

func (forceEnd) ShowTips(int64, int64) string {
	return "强制结束所有游戏"
}

func (forceEnd) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (forceEnd) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	returnError := httpClient.HTTPGet("/forceend", nil)
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	return MessageChain{&Text{Text: "已执行"}}
}

type setNotice struct{}

func (setNotice) Name() string {
	return "修改公告"
}

func (setNotice) ShowTips(int64, int64) string {
	return "修改公告 公告内容"
}

func (setNotice) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (setNotice) Execute(_ *GroupMessage, content string) MessageChain {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n修改公告 公告内容"}}
	}
	returnError := httpClient.HTTPGet("/setnotice", map[string]string{"notice": content})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	return MessageChain{&Text{Text: "公告已变更"}}
}

type setWaitSecond struct{}

func (setWaitSecond) Name() string {
	return "修改出牌时间"
}

func (setWaitSecond) ShowTips(int64, int64) string {
	return "修改出牌时间 秒数"
}

func (setWaitSecond) CheckAuth(_ int64, senderID int64) bool {
	return isAdmin(senderID)
}

func (setWaitSecond) Execute(_ *GroupMessage, content string) MessageChain {
	if second, err := strconv.Atoi(strings.TrimSpace(content)); err != nil {
		return MessageChain{&Text{Text: "命令格式：\n修改出牌时间 秒数"}}
	} else if second <= 0 {
		return MessageChain{&Text{Text: "出牌时间必须大于0"}}
	}
	returnError := httpClient.HTTPGet("/updatewaitsecond", map[string]string{"second": content})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	return MessageChain{&Text{Text: "默认出牌时间已修改为" + content + "秒"}}
}

type createAccount struct{}

func (createAccount) Name() string {
	return "创号"
}

func (createAccount) ShowTips(int64, int64) string {
	return ""
}

func (createAccount) CheckAuth(_ int64, senderID int64) bool {
	return isSuperAdmin(senderID)
}

func (createAccount) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n创号 名字"}}
	}
	result, returnError := httpClient.HTTPGetBool("/register", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	if !result {
		return MessageChain{&Text{Text: "用户名重复"}}
	}
	return MessageChain{&Text{Text: "创号成功"}}
}

type addEnergy struct{}

func (addEnergy) Name() string {
	return "增加精力"
}

func (addEnergy) ShowTips(int64, int64) string {
	return "增加精力 名字 数量"
}

func (addEnergy) CheckAuth(_ int64, senderID int64) bool {
	return isSuperAdmin(senderID)
}

func (addEnergy) Execute(_ *GroupMessage, content string) MessageChain {
	arr := strings.Split(strings.TrimSpace(content), " ")
	if len(arr) != 2 {
		return MessageChain{&Text{Text: "命令格式：\n增加精力 名字 数量"}}
	}
	name := strings.TrimSpace(arr[0])
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n增加精力 名字 数量"}}
	}
	energy := strings.TrimSpace(arr[1])
	_, err := strconv.Atoi(energy)
	if err != nil {
		return MessageChain{&Text{Text: "命令格式：\n增加精力 名字 数量"}}
	}
	success, returnError := httpClient.HTTPGetBool("/addenergy", map[string]string{"name": name, "energy": energy})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError)
		return returnError.Message
	}
	if !success {
		return MessageChain{&Text{Text: "增加精力失败"}}
	}
	return MessageChain{&Text{Text: "增加精力成功"}}
}
