package maplebot

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/CuteReimu/YinYangJade/db"
	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
)

var restyClient = resty.New()

func init() {
	restyClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	restyClient.SetTimeout(20 * time.Second)
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69",
		"connection":   "close",
	})
}

var B *Bot

func Init(b *Bot) {
	initConfig()
	B = b
	go func() {
		B.Run(clearExpiredImages)
		B.Run(clearExpiredImages2)
		for range time.Tick(24 * time.Hour) {
			B.Run(clearExpiredImages)
			B.Run(clearExpiredImages2)
		}
	}()
	B.ListenGroupMessage(cmdHandleFunc)
	B.ListenGroupMessage(handleDictionary)
	B.ListenGroupMessage(searchAt)
}

var cmdMap = make(map[string]iface.CmdHandler)

func cmdHandleFunc(message *GroupMessage) bool {
	if !slices.Contains(config.GetIntSlice("qq_groups"), int(message.GroupId)) {
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

func addSimpleCmdListenerNoContent(name string, f func() MessageChain) {
	addSimpleCmdListener(name, func(content string) MessageChain {
		if len(content) > 0 {
			return nil
		}
		return f()
	})
}

func addSimpleCmdListener(name string, f func(string) MessageChain) {
	if _, ok := cmdMap[name]; ok {
		panic("repeat command: " + name)
	}
	cmdMap[name] = &iface.SimpleCmdHandler{HandlerName: name, Handler: f}
}

func searchAt(message *GroupMessage) bool {
	if len(message.Message) == 0 {
		return true
	}
	if !slices.Contains(config.GetIntSlice("qq_groups"), int(message.GroupId)) {
		return true
	}
	if len(message.Message) >= 2 {
		if len(message.Message) >= 2 {
			if text, ok := message.Message[0].(*Text); ok && strings.TrimSpace(text.Text) == "查询" {
				if at, ok := message.Message[1].(*At); ok {
					data := findRoleData.GetStringMapString("data")
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
							sendGroupMessage(message, findRole(name)...)
						}()
					}
				}
				return true
			}
		}
	}
	return true
}

func init() {
	addSimpleCmdListenerNoContent("8421", calculatePotion)
	addSimpleCmdListener("等级压制", calculateExpDamage)
	addSimpleCmdListener("生成表格", genTable)
	addCmdListener(&levelUpExp{})
	addCmdListener(&boomCount{})
	addCmdListener(&searchMe{})
	addCmdListener(&searchSomeone{})
	addCmdListener(&searchBind{})
	addCmdListener(&bind{})
	addCmdListener(&unbind{})
	addSimpleCmdListenerNoContent("神秘压制", GetMoreDamageArc)
	tryStarForce := func(content string) MessageChain {
		if len(content) == 0 {
			return MessageChain{&Text{Text: "命令格式：\r\n模拟升星 200 0 22\r\n后面可以加：七折、减爆、保护"}}
		}
		return calculateStarForce1(true, content)
	}
	tryStarForceOld := func(content string) MessageChain {
		if len(content) == 0 {
			return MessageChain{&Text{Text: "命令格式：\r\n模拟升星旧 200 0 22\r\n后面可以加：七折、减爆、保护"}}
		}
		return calculateStarForce1(false, content)
	}
	for _, s := range []string{"模拟升星", "模拟上星", "升星期望", "上星期望"} {
		addSimpleCmdListener(s, tryStarForce)
		addSimpleCmdListener(s+"旧", tryStarForceOld)
	}
	addCmdListener(&tryCube{})
	bossList := []rune{'3', '6', '7', '8', '9', '4', 'M', '绿', '黑', '赛', '狗'}
	var sb strings.Builder
	_, _ = sb.WriteString("订阅开车")
	for _, boss := range bossList {
		_, _ = sb.WriteString(" ")
		_, _ = sb.WriteRune(boss)
	}
	addCmdListener(&iWannaFormParty{bossList: bossList})
	addCmdListener(&registerFormParty{bossList: bossList, tips: sb.String()})
	addCmdListener(&cancelRegisterFormParty{bossList: bossList})
}

type levelUpExp struct{}

func (l *levelUpExp) Name() string {
	return "升级经验"
}

func (l *levelUpExp) ShowTips(int64, int64) string {
	return ""
}

func (l *levelUpExp) CheckAuth(int64, int64) bool {
	return true
}

func (l *levelUpExp) Execute(_ *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		return calculateLevelExp()
	}
	parts := strings.SplitN(content, " ", 2)
	if len(parts) == 2 {
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		return calculateExpBetweenLevel(int64(start), int64(end))
	}
	return nil
}

type boomCount struct{}

func (b *boomCount) Name() string {
	return "爆炸次数"
}

func (b *boomCount) ShowTips(int64, int64) string {
	return ""
}

func (b *boomCount) CheckAuth(int64, int64) bool {
	return true
}

func (b *boomCount) Execute(_ *GroupMessage, content string) MessageChain {
	return slices.Concat(calculateBoomCount(content, true), calculateBoomCount(content, false))
}

type searchMe struct{}

func (s *searchMe) Name() string {
	return "查询我"
}

func (s *searchMe) ShowTips(int64, int64) string {
	return "查询我"
}

func (s *searchMe) CheckAuth(int64, int64) bool {
	return true
}

func (s *searchMe) Execute(msg *GroupMessage, content string) MessageChain {
	if len(content) > 0 {
		return nil
	}
	data := findRoleData.GetStringMapString("data")
	name := data[strconv.FormatInt(msg.Sender.UserId, 10)]
	if len(name) == 0 {
		return MessageChain{&Text{Text: "你还未绑定"}}
	} else {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("panic recovered", "error", err)
				}
			}()
			sendGroupMessage(msg, findRole(name)...)
		}()
	}
	return nil
}

type searchSomeone struct{}

func (s *searchSomeone) Name() string {
	return "查询"
}

func (s *searchSomeone) ShowTips(int64, int64) string {
	return "查询 游戏名"
}

func (s *searchSomeone) CheckAuth(int64, int64) bool {
	return true
}

func (s *searchSomeone) Execute(msg *GroupMessage, content string) MessageChain {
	if len(content) == 0 || strings.Contains(content, " ") {
		return nil
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "stack", string(debug.Stack()))
			}
		}()
		sendGroupMessage(msg, findRole(content)...)
	}()
	return nil
}

type searchBind struct{}

func (s *searchBind) Name() string {
	return "查询绑定"
}

func (s *searchBind) ShowTips(int64, int64) string {
	return ""
}

func (s *searchBind) CheckAuth(int64, int64) bool {
	return true
}

func (s *searchBind) Execute(_ *GroupMessage, content string) MessageChain {
	qq, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		return MessageChain{&Text{Text: "命令格式：“查询绑定 QQ号”"}}
	}
	data := findRoleData.GetStringMapString("data")
	if name := data[strconv.FormatInt(qq, 10)]; name != "" {
		return MessageChain{&Text{Text: "该玩家绑定了：" + name}}
	} else {
		return MessageChain{&Text{Text: "该玩家还未绑定"}}
	}
}

type bind struct{}

func (b *bind) Name() string {
	return "绑定"
}

func (b *bind) ShowTips(int64, int64) string {
	return ""
}

func (b *bind) CheckAuth(int64, int64) bool {
	return true
}

func (b *bind) Execute(msg *GroupMessage, content string) MessageChain {
	data := findRoleData.GetStringMapString("data")
	if data[strconv.FormatInt(msg.Sender.UserId, 10)] != "" {
		return MessageChain{&Text{Text: "你已经绑定过了，如需更换请先解绑"}}
	} else if len(content) > 0 && !strings.Contains(content, " ") {
		data[strconv.FormatInt(msg.Sender.UserId, 10)] = content
		findRoleData.Set("data", data)
		if err := findRoleData.WriteConfig(); err != nil {
			slog.Error("write config failed", "error", err)
		}
		return MessageChain{&Text{Text: "绑定成功"}}
	}
	return nil
}

type unbind struct{}

func (u *unbind) Name() string {
	return "解绑"
}

func (u *unbind) ShowTips(int64, int64) string {
	return ""
}

func (u *unbind) CheckAuth(int64, int64) bool {
	return true
}

func (u *unbind) Execute(msg *GroupMessage, content string) MessageChain {
	if len(content) > 0 {
		return nil
	}
	data := findRoleData.GetStringMapString("data")
	if data[strconv.FormatInt(msg.Sender.UserId, 10)] != "" {
		delete(data, strconv.FormatInt(msg.Sender.UserId, 10))
		findRoleData.Set("data", data)
		if err := findRoleData.WriteConfig(); err != nil {
			slog.Error("write config failed", "error", err)
		}
		return MessageChain{&Text{Text: "解绑成功"}}
	} else {
		return MessageChain{&Text{Text: "你还未绑定"}}
	}
}

type tryCube struct{}

func (t *tryCube) Name() string {
	return "洗魔方"
}

func (t *tryCube) ShowTips(int64, int64) string {
	return ""
}

func (t *tryCube) CheckAuth(int64, int64) bool {
	return true
}

func (t *tryCube) Execute(_ *GroupMessage, content string) MessageChain {
	if len(content) == 0 {
		return calculateCubeAll()
	}
	return calculateCube(content)
}

type iWannaFormParty struct {
	bossList []rune
}

func (i *iWannaFormParty) Name() string {
	return "我要开车"
}

func (i *iWannaFormParty) ShowTips(int64, int64) string {
	return ""
}

func (i *iWannaFormParty) CheckAuth(int64, int64) bool {
	return true
}

func (i *iWannaFormParty) Execute(msg *GroupMessage, data string) MessageChain {
	arr := getBossNumber(i.bossList, data)
	if len(arr) == 0 {
		return MessageChain{&Text{Text: "不准开车!"}}
	} else {
		var messageArr []SingleMessage
		qqNumbers := make(map[string]bool) // 用map，随机顺序取人，公平一些
		messageArr = append(messageArr, &Text{Text: string(arr) + " 发车了! "})
		for _, num := range arr {
			subscribed, _ := db.Get("boss_subscribe_" + string(num))
			subArr := strings.SplitSeq(subscribed, ",")
			for qqNumber := range subArr {
				qqNumbers[qqNumber] = true
			}
		}
		delete(qqNumbers, strconv.Itoa(int(msg.Sender.UserId))) // 自己发车不要艾特自己
		members, err := B.GetGroupMemberList(msg.GroupId)
		if err != nil {
			slog.Error("获取群成员列表失败", "error", err, "group_id", msg.GroupId)
		}
		groupMembers := make(map[int64]bool) // 把list转成map，方便查找
		for _, member := range members {
			groupMembers[member.UserId] = true
		}
		for qqNumber := range qqNumbers {
			if err == nil { // 获取群成员列表失败，就不检查了
				qq, err2 := strconv.ParseInt(qqNumber, 10, 64)
				if err2 != nil {
					slog.Error("QQ号解析错误", "error", err2, "qq", qqNumber)
					continue
				}
				if !groupMembers[qq] { // 群里无此人
					continue
				}
			}
			messageArr = append(messageArr, &At{QQ: qqNumber})
			if len(messageArr) > 20 { // 最多艾特20个人
				break
			}
		}
		return messageArr
	}
}

type registerFormParty struct {
	bossList []rune
	tips     string
}

func (r *registerFormParty) Name() string {
	return "订阅开车"
}

func (r *registerFormParty) ShowTips(int64, int64) string {
	return r.tips
}

func (r *registerFormParty) CheckAuth(int64, int64) bool {
	return true
}

func (r *registerFormParty) Execute(msg *GroupMessage, content string) MessageChain {
	arr := getBossNumber(r.bossList, content)
	if len(arr) == 0 {
		return MessageChain{&Text{Text: "这是去幼儿园的车"}}
	} else {
		userId := strconv.Itoa(int(msg.Sender.UserId))
		subscribe(arr, userId)
		return MessageChain{&Text{Text: "订阅成功 " + string(arr)}}
	}
}

type cancelRegisterFormParty struct {
	bossList []rune
}

func (c *cancelRegisterFormParty) Name() string {
	return "取消订阅"
}

func (c *cancelRegisterFormParty) ShowTips(int64, int64) string {
	return "取消订阅"
}

func (c *cancelRegisterFormParty) CheckAuth(int64, int64) bool {
	return true
}

func (c *cancelRegisterFormParty) Execute(msg *GroupMessage, content string) MessageChain {
	userId := strconv.Itoa(int(msg.Sender.UserId))
	if len(content) == 0 {
		unSubscribe(c.bossList, userId)
		return MessageChain{&Text{Text: "取消全部订阅成功"}}
	} else {
		arr := getBossNumber(c.bossList, content)
		unSubscribe(arr, userId)
		return MessageChain{&Text{Text: "取消订阅成功 " + string(arr)}}
	}
}

func subscribe(arr []rune, userId string) {
	for _, num := range arr {
		subscribed, _ := db.Get("boss_subscribe_" + string(num))
		subArr := strings.Split(subscribed, ",")
		if slices.Contains(subArr, userId) {
			continue
		}
		subArr = append(subArr, userId)
		if len(subArr) > 50 { // 最多存50个人，多了就把旧的删了
			subArr = subArr[:50]
		}
		subscribed = strings.Join(subArr, ",")
		db.Set("boss_subscribe_"+string(num), subscribed)
	}
}

func unSubscribe(arr []rune, userId string) {
	for _, num := range arr {
		subscribed, _ := db.Get("boss_subscribe_" + string(num))
		subArr := strings.Split(subscribed, ",")
		pos := slices.Index(subArr, userId)
		if pos >= 0 {
			subArr = append(subArr[:pos], subArr[pos+1:]...)
			subscribed = strings.Join(subArr, ",")
			db.Set("boss_subscribe_"+string(num), subscribed)
		}
	}
}

func getBossNumber(bossList []rune, numberString string) []rune {
	numberString = strings.ToUpper(numberString)
	var result []rune
	for _, char := range numberString {
		if slices.Contains(bossList, char) {
			if slices.Contains(result, char) {
				continue
			}
			result = append(result, char)
		}
	}
	return result
}

func saveImage(message MessageChain) (MessageChain, error) {
	for _, m := range message {
		if img, ok := m.(*Image); ok && len(img.Url) > 0 {
			u := img.Url
			_, err := url.Parse(u)
			if err != nil {
				slog.Error("userInput is not a valid URL, reject it", "error", err)
				return message, err
			}
			if err := os.MkdirAll("chat-images", 0755); err != nil {
				slog.Error("mkdir failed", "error", err)
				return message, errors.New("保存图片失败")
			}
			nameLen := len(filepath.Ext(img.File)) + 32
			if len(img.File) > nameLen {
				img.File = img.File[len(img.File)-nameLen:]
			}
			p := filepath.Join("chat-images", img.File)
			abs, err := filepath.Abs(p)
			if err != nil {
				slog.Error("filepath.Abs() failed", "error", err)
				return message, errors.New("保存图片失败")
			}
			cmd := exec.Command("curl", "-o", p, u)
			if out, err := cmd.CombinedOutput(); err != nil {
				slog.Error("cmd.Run() failed", "error", err)
				return message, errors.New("保存图片失败")
			} else {
				slog.Debug(string(out))
			}
			img.File = "file://" + abs
			img.Url = ""
		} else if forward, ok := m.(*Forward); ok {
			msgs, err := B.GetForwardMessage(forward.Id)
			if err != nil {
				slog.Error("获取转发消息失败", "forwardId", forward.Id)
				return message, errors.New("获取转发消息失败")
			}
			var ret MessageChain
			for _, msg := range msgs {
				if m, ok := msg.(*GroupMessage); ok {
					newMsg, err := saveImage(m.Message)
					if err != nil {
						slog.Error("嵌套saveImage失败", "error", err)
					}
					ret = append(ret, &Node{
						UserId:   strconv.FormatInt(m.UserId, 10),
						Nickname: m.Sender.Nickname,
						Content:  newMsg,
					})
				}
			}
			return ret, nil
		}
	}
	return message, nil
}

func sendGroupMessage(context *GroupMessage, messages ...SingleMessage) {
	replyGroupMessage(false, context, messages...)
}

func replyGroupMessage(reply bool, context *GroupMessage, messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		var f1 func(messages []SingleMessage)
		f1 = func(messages []SingleMessage) {
			for _, m := range messages {
				if img, ok := m.(*Image); ok && len(img.File) > 0 {
					if strings.HasPrefix(img.File, "file://") {
						fileName := img.File[len("file://"):]
						buf, err := os.ReadFile(fileName)
						if err != nil {
							slog.Error("read file failed", "error", err)
							continue
						}
						img.File = "base64://" + base64.StdEncoding.EncodeToString(buf)
					}
				} else if node, ok := m.(*Node); ok {
					f1(node.Content)
				}
			}
		}
		f1(messages)
		if !reply {
			_, err := B.SendGroupMessage(context.GroupId, messages)
			return err
		}
		_, err := B.SendGroupMessage(context.GroupId, append(MessageChain{
			&Reply{Id: strconv.FormatInt(int64(context.MessageId), 10)},
		}, messages...))
		return err
	}
	if err := f(messages); err != nil {
		slog.Error("send group message failed", "error", err)
		newMessages := make([]SingleMessage, 0, len(messages))
		for _, m := range messages {
			if image, ok := m.(*Image); !ok || !strings.HasPrefix(image.File, "http") {
				newMessages = append(newMessages, m)
			}
		}
		if len(newMessages) != len(messages) && len(newMessages) > 0 {
			if err = f(newMessages); err != nil {
				slog.Error("send group message failed", "error", err)
			}
		}
	}
}

func dealKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "零", "0")
	s = strings.ReplaceAll(s, "一", "1")
	s = strings.ReplaceAll(s, "二", "2")
	s = strings.ReplaceAll(s, "三", "3")
	s = strings.ReplaceAll(s, "四", "4")
	s = strings.ReplaceAll(s, "五", "5")
	s = strings.ReplaceAll(s, "六", "6")
	s = strings.ReplaceAll(s, "七", "7")
	s = strings.ReplaceAll(s, "八", "8")
	s = strings.ReplaceAll(s, "九", "9")
	return strings.ToLower(s)
}

func clearExpiredImages() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered", "error", err)
		}
	}()
	data := qunDb.GetStringMapString("data")
	data2 := make(map[string]bool)
	var f func(ms MessageChain)
	f = func(ms MessageChain) {
		for _, m := range ms {
			if img, ok := m.(*Image); ok && len(img.File) > 0 && strings.HasPrefix(img.File, "file://") {
				data2[filepath.Base(img.File[len("file://"):])] = true
			} else if node, ok := m.(*Node); ok {
				f(node.Content)
			}
		}
	}
	for _, v := range data {
		var ms MessageChain
		if err := json.Unmarshal([]byte(v), &ms); err != nil {
			slog.Error("json unmarshal failed", "error", err)
			continue
		}
		f(ms)
	}
	files, err := os.ReadDir("chat-images")
	if err != nil {
		slog.Error("read dir failed", "error", err)
	}
	for _, file := range files {
		if !data2[file.Name()] {
			if err = os.Remove(filepath.Join("chat-images", file.Name())); err != nil {
				slog.Error("remove file failed", "error", err)
			}
		}
	}
}
