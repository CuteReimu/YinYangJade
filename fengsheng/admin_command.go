package fengsheng

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	. "github.com/CuteReimu/onebot"
)

func init() {
	addCmdListener(&bind{})
	addCmdListener(&unbind{})
	addCmdListener(&unbindExpired{})
	addCmdListener(&forbidPlayer{})
	addCmdListener(&releasePlayer{})
	addCmdListener(&forbidRole{})
	addCmdListener(&releaseRole{})
	addCmdListener(&setVersion{})
	addCmdListener(&forceEnd{})
	addCmdListener(&setNotice{})
	addCmdListener(&setWaitSecond{})
	addCmdListener(&createAccount{})
	addCmdListener(&addEnergy{})
}

type bind struct{}

func (a *bind) Name() string {
	return "绑定"
}

func (a *bind) ShowTips(int64, int64) string {
	return "绑定 QQ号 名字"
}

func (a *bind) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *bind) Execute(msg *GroupMessage, content string) MessageChain {
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
	result, returnError := httpGetString("/getscore", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
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

func (a *unbind) Name() string {
	return "解绑"
}

func (a *unbind) ShowTips(int64, int64) string {
	return "解绑 QQ号"
}

func (a *unbind) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *unbind) Execute(_ *GroupMessage, content string) MessageChain {
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

func (a *unbindExpired) Name() string {
	return "解绑所有0分玩家"
}

func (a *unbindExpired) ShowTips(int64, int64) string {
	return "解绑所有0分玩家"
}

func (a *unbindExpired) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (a *unbindExpired) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	m1 := permData.GetStringMapString("playerMap")
	m2 := make(map[string]string, len(m1))
	for id, name := range m1 {
		score, returnError := httpGetString("/getscore", map[string]string{"name": name})
		if returnError != nil {
			slog.Error("请求失败", "error", returnError.error)
			return returnError.message
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

func (a *forbidPlayer) Name() string {
	return "封号"
}

func (a *forbidPlayer) ShowTips(int64, int64) string {
	return "封号 名字 小时"
}

func (a *forbidPlayer) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *forbidPlayer) Execute(_ *GroupMessage, content string) MessageChain {
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
	result, returnError := httpGetString("/forbidplayer", map[string]string{
		"name": name,
		"hour": c[spaceIndex+1:],
	})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return MessageChain{&Text{Text: result}}
}

type releasePlayer struct{}

func (a *releasePlayer) Name() string {
	return "解封"
}

func (a *releasePlayer) ShowTips(int64, int64) string {
	return "解封 名字"
}

func (a *releasePlayer) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *releasePlayer) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n解封 名字"}}
	}
	result, returnError := httpGetString("/releaseplayer", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return MessageChain{&Text{Text: result}}
}

type forbidRole struct{}

func (a *forbidRole) Name() string {
	return "禁用角色"
}

func (a *forbidRole) ShowTips(int64, int64) string {
	return "禁用角色 名字"
}

func (a *forbidRole) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *forbidRole) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n禁用角色 名字"}}
	}
	result, returnError := httpGetBool("/forbidrole", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if result {
		return MessageChain{&Text{Text: "禁用成功"}}
	}
	return MessageChain{&Text{Text: "禁用失败"}}
}

type releaseRole struct{}

func (a *releaseRole) Name() string {
	return "启用角色"
}

func (a *releaseRole) ShowTips(int64, int64) string {
	return "启用角色 名字"
}

func (a *releaseRole) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (a *releaseRole) Execute(_ *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n启用角色 名字"}}
	}
	result, returnError := httpGetBool("/releaserole", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if result {
		return MessageChain{&Text{Text: "启用成功"}}
	}
	return MessageChain{&Text{Text: "启用失败"}}
}

type setVersion struct{}

func (s *setVersion) Name() string {
	return "修改版本号"
}

func (s *setVersion) ShowTips(int64, int64) string {
	return "修改版本号 版本号"
}

func (s *setVersion) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (s *setVersion) Execute(_ *GroupMessage, content string) MessageChain {
	if _, err := strconv.Atoi(content); err != nil {
		return MessageChain{&Text{Text: "命令格式：\n修改版本号 版本号"}}
	}
	returnError := httpGet("/setversion", map[string]string{"version": content})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return MessageChain{&Text{Text: "版本号已修改为" + content}}
}

type forceEnd struct{}

func (f *forceEnd) Name() string {
	return "强制结束所有游戏"
}

func (f *forceEnd) ShowTips(int64, int64) string {
	return "强制结束所有游戏"
}

func (f *forceEnd) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (f *forceEnd) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	returnError := httpGet("/forceend", nil)
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return MessageChain{&Text{Text: "已执行"}}
}

type setNotice struct{}

func (s *setNotice) Name() string {
	return "修改公告"
}

func (s *setNotice) ShowTips(int64, int64) string {
	return "修改公告 公告内容"
}

func (s *setNotice) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (s *setNotice) Execute(_ *GroupMessage, content string) MessageChain {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n修改公告 公告内容"}}
	}
	returnError := httpGet("/setnotice", map[string]string{"notice": content})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return MessageChain{&Text{Text: "公告已变更"}}
}

type setWaitSecond struct{}

func (s *setWaitSecond) Name() string {
	return "修改出牌时间"
}

func (s *setWaitSecond) ShowTips(int64, int64) string {
	return "修改出牌时间 秒数"
}

func (s *setWaitSecond) CheckAuth(_ int64, senderId int64) bool {
	return IsAdmin(senderId)
}

func (s *setWaitSecond) Execute(_ *GroupMessage, content string) MessageChain {
	if second, err := strconv.Atoi(strings.TrimSpace(content)); err != nil {
		return MessageChain{&Text{Text: "命令格式：\n修改出牌时间 秒数"}}
	} else if second <= 0 {
		return MessageChain{&Text{Text: "出牌时间必须大于0"}}
	}
	returnError := httpGet("/updatewaitsecond", map[string]string{"second": content})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return MessageChain{&Text{Text: "默认出牌时间已修改为" + content + "秒"}}
}

type createAccount struct{}

func (c *createAccount) Name() string {
	return "创号"
}

func (c *createAccount) ShowTips(int64, int64) string {
	return ""
}

func (c *createAccount) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (c *createAccount) Execute(msg *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	if len(name) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n创号 名字"}}
	}
	result, returnError := httpGetBool("/register", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !result {
		return MessageChain{&Text{Text: "用户名重复"}}
	}
	return MessageChain{&Text{Text: "创号成功"}}
}

type addEnergy struct{}

func (c *addEnergy) Name() string {
	return "增加精力"
}

func (c *addEnergy) ShowTips(int64, int64) string {
	return "增加精力 名字 数量"
}

func (c *addEnergy) CheckAuth(_ int64, senderId int64) bool {
	return IsSuperAdmin(senderId)
}

func (c *addEnergy) Execute(msg *GroupMessage, content string) MessageChain {
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
	success, returnError := httpGetBool("/addenergy", map[string]string{"name": name, "energy": energy})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !success {
		return MessageChain{&Text{Text: "增加精力失败"}}
	}
	return MessageChain{&Text{Text: "增加精力成功"}}
}
