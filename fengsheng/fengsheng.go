package fengsheng

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/CuteReimu/YinYangJade/db"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"log/slog"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"
)

var tierName = map[string]string{
	"🥉":  "青铜",
	"🥈":  "白银",
	"🥇":  "黄金",
	"💍":  "铂金",
	"💠":  "钻石",
	"👑":  "大师",
	"☀️": "至尊",
}

func dealGetScore(result string) MessageChain {
	var isWinRate, isHistory bool
	var resultBuilder strings.Builder
	var winRateData, historyData [][]string
	for _, line := range strings.Split(result, "\n") {
		if len(line) == 0 {
			continue
		}
		if line == "---------------------------------" {
			if !isWinRate {
				isWinRate = true
			} else {
				isWinRate = false
				isHistory = true
			}
			continue
		}
		if strings.HasPrefix(line, "剩余精力") {
			_, _ = resultBuilder.WriteString("，" + line)
		} else if strings.HasPrefix(line, "身份\t 胜率\t 平均胜率\t 场次") || strings.HasPrefix(line, "最近") && strings.HasSuffix(line, "场战绩") {
		} else if isWinRate {
			var r []string
			for _, s := range strings.Split(line, "\t") {
				r = append(r, strings.TrimSpace(s))
			}
			winRateData = append(winRateData, r)
		} else if isHistory {
			arr := strings.Split(line, ",")
			identity := strings.ReplaceAll(strings.ReplaceAll(arr[1], "神秘人[", ""), "]", "")
			role := strings.ReplaceAll(arr[0], "(死亡)", "")
			alive := "存活"
			if strings.Contains(arr[0], "(死亡)") {
				alive = "死亡"
			}
			tier := arr[3]
			for t, t1 := range tierName {
				tier = strings.Replace(tier, t, t1, 1)
			}
			historyData = append(historyData, []string{role, alive, identity, arr[2], tier, arr[4]})
		} else {
			_, _ = resultBuilder.WriteString(line)
		}
	}
	slices.Reverse(historyData)
	ret := make(MessageChain, 0, 3)
	ret = append(ret, &Text{Text: resultBuilder.String()})
	if len(winRateData) > 0 {
		p, err := TableOptionRender(TableChartOption{
			Header:     []string{"身份", "胜率", "平均胜率", "场次"},
			Data:       winRateData,
			Width:      440,
			TextAligns: []string{AlignLeft, AlignLeft, AlignLeft, AlignLeft},
		})
		if err != nil {
			slog.Error("render chart failed", "error", err)
		} else if buf, err := p.Bytes(); err != nil {
			slog.Error("render chart failed", "error", err)
		} else {
			ret = append(ret, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)})
		}
	}
	if len(historyData) > 0 {
		p, err := TableOptionRender(TableChartOption{
			Header:     []string{"角色", "存活", "身份", "胜负", "段位", "分数"},
			Data:       historyData,
			Width:      720,
			TextAligns: []string{AlignLeft, AlignLeft, AlignLeft, AlignLeft, AlignLeft, AlignLeft},
		})
		if err != nil {
			slog.Error("render chart failed", "error", err)
		} else if buf, err := p.Bytes(); err != nil {
			slog.Error("render chart failed", "error", err)
		} else {
			ret = append(ret, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)})
		}
	}
	return ret
}

func init() {
	addCmdListener(&getMyScore{})
	addCmdListener(&getScore{})
	addCmdListener(&rankList{})
	addCmdListener(&seasonRankList{})
	addCmdListener(&winRate{})
	addCmdListener(&register{})
	addCmdListener(&addNotifyOnStart{})
	addCmdListener(&addNotifyOnEnd{})
	addCmdListener(&addNotifyOnEnd2{})
	addCmdListener(&atPlayer{})
	addCmdListener(&resetPwd{})
	addCmdListener(&sign{})
	addCmdListener(&frequency{})
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
	name := data[strconv.FormatInt(msg.Sender.UserId, 10)]
	if len(name) == 0 {
		return MessageChain{&Text{Text: "请先绑定"}}
	}
	result, returnError := httpGetString("/getscore", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return dealGetScore(result)
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
	result, returnError := httpGetString("/getscore", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	return dealGetScore(result)
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
	return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(resp.Body())}}
}

type seasonRankList struct{}

func (r *seasonRankList) Name() string {
	return "赛季最高分排行"
}

func (r *seasonRankList) ShowTips(int64, int64) string {
	return "赛季最高分排行"
}

func (r *seasonRankList) CheckAuth(int64, int64) bool {
	return true
}

func (r *seasonRankList) Execute(_ *GroupMessage, content string) MessageChain {
	content = strings.TrimSpace(content)
	if len(content) > 0 {
		return nil
	}
	resp, err := restyClient.R().SetQueryParam("season_rank", "true").Get(fengshengConfig.GetString("fengshengUrl") + "/ranklist")
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	if resp.StatusCode() != 200 {
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(resp.Body())}}
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
	return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(resp.Body())}}
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
		return MessageChain{&Text{Text: "命令格式：\n注册 名字"}}
	}
	data := permData.GetStringMapString("playerMap")
	if oldName := data[strconv.FormatInt(msg.Sender.UserId, 10)]; len(oldName) > 0 {
		return MessageChain{&Text{Text: "你已经注册过：" + oldName}}
	}
	result, returnError := httpGetBool("/register", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !result {
		return MessageChain{&Text{Text: "用户名重复"}}
	}
	data[strconv.FormatInt(msg.Sender.UserId, 10)] = name
	permData.Set("playerMap", data)
	if err := permData.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	return MessageChain{&Text{Text: "注册成功"}}
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
	result, returnError := httpGetBool("/addnotify", map[string]string{
		"qq": strconv.FormatInt(msg.Sender.UserId, 10),
	})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !result {
		return MessageChain{&Text{Text: "太多人预约了，不能再添加了"}}
	}
	return MessageChain{&Text{Text: "好的，开了喊你"}}
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
	result, returnError := httpGetBool("/addnotify", map[string]string{
		"qq":   strconv.FormatInt(msg.Sender.UserId, 10),
		"when": "1",
	})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !result {
		return MessageChain{&Text{Text: "太多人预约了，不能再添加了"}}
	}
	return MessageChain{&Text{Text: "好的，结束喊你"}}
}

type addNotifyOnEnd2 struct{}

func (a *addNotifyOnEnd2) Name() string {
	return "好了喊我"
}

func (a *addNotifyOnEnd2) ShowTips(int64, int64) string {
	return ""
}

func (a *addNotifyOnEnd2) CheckAuth(int64, int64) bool {
	return true
}

func (a *addNotifyOnEnd2) Execute(msg *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) > 0 {
		return nil
	}
	result, returnError := httpGetBool("/addnotify", map[string]string{
		"qq":   strconv.FormatInt(msg.Sender.UserId, 10),
		"when": "1",
	})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !result {
		return MessageChain{&Text{Text: "太多人预约了，不能再添加了"}}
	}
	return MessageChain{&Text{Text: "好的，好了喊你"}}
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
		return MessageChain{&Text{Text: "命令格式：\n艾特 游戏内的名字"}}
	}
	data := permData.GetStringMapString("playerMap")
	for id, v := range data {
		if v == content {
			_, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				slog.Error("parse int failed: " + id)
				return nil
			}
			return MessageChain{&At{QQ: id}}
		}
	}
	return MessageChain{&Text{Text: "没能找到此玩家，可能还未绑定"}}
}

type resetPwd struct{}

func (u *resetPwd) Name() string {
	return "重置密码"
}

func (u *resetPwd) ShowTips(_ int64, senderId int64) string {
	data := permData.GetStringMapString("playerMap")
	if _, ok := data[strconv.FormatInt(senderId, 10)]; ok {
		return "重置密码"
	}
	if IsAdmin(senderId) {
		return "重置密码 名字"
	}
	return ""
}

func (u *resetPwd) CheckAuth(int64, int64) bool {
	return true
}

func (u *resetPwd) Execute(msg *GroupMessage, content string) MessageChain {
	name := strings.TrimSpace(content)
	data := permData.GetStringMapString("playerMap")
	var result string
	var returnError *errorWithMessage
	if len(name) == 0 {
		playerName := data[strconv.FormatInt(msg.Sender.UserId, 10)]
		if len(playerName) == 0 {
			if !IsAdmin(msg.Sender.UserId) {
				return nil
			}
			return MessageChain{&Text{Text: "重置密码 名字"}}
		}
		result, returnError = httpGetString("/resetpwd", map[string]string{"name": playerName})
	} else {
		if !IsAdmin(msg.Sender.UserId) {
			return nil
		}
		result, returnError = httpGetString("/resetpwd", map[string]string{"name": name})
	}
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if len(result) == 0 {
		return nil
	}
	return MessageChain{&Text{Text: result}}
}

type sign struct{}

func (s *sign) Name() string {
	return "签到"
}

func (s *sign) ShowTips(int64, int64) string {
	return "签到"
}

func (s *sign) CheckAuth(_ int64, _ int64) bool {
	return true
}

func (s *sign) Execute(msg *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) != 0 {
		return nil
	}
	qq := strconv.FormatInt(msg.Sender.UserId, 10)
	data := permData.GetStringMapString("playerMap")
	name := data[qq]
	if len(name) == 0 {
		return MessageChain{&Text{Text: "请先注册"}}
	}
	var lastSignTime int64
	lastSignTimeStr, ok := db.Get("fengsheng_sign:" + qq)
	if ok {
		lastSignTime, _ = strconv.ParseInt(lastSignTimeStr, 10, 64)
	}
	now := time.Now()
	if lastSignTime > 0 {
		y1, m1, d1 := time.UnixMilli(lastSignTime).Date()
		y2, m2, d2 := now.Date()
		if y1 == y2 && m1 == m2 && d1 == d2 {
			return MessageChain{&Text{Text: "今天已经签到过了，明天再来吧"}}
		}
	}
	lastTime, returnError := httpGetInt("/getlasttime", map[string]string{"name": name})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if lastTime >= 7*24*3600*1000 {
		lastTime1 := time.Now().Add(-(time.Duration(lastTime) * time.Millisecond))
		return MessageChain{&Text{Text: "一周内未进行过游戏，无法进行签到 最近时间为：" + lastTime1.Format("2006-01-02 15:04:05")}}
	}
	energy := rand.IntN(10)/3 + 1
	success, returnError := httpGetBool("/addenergy", map[string]string{"name": name, "energy": strconv.Itoa(energy)})
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	if !success {
		return MessageChain{&Text{Text: "签到失败"}}
	}
	db.Set("fengsheng_sign:"+qq, strconv.FormatInt(now.UnixMilli(), 10), time.Hour*48)
	switch energy {
	case 1:
		return MessageChain{&Text{Text: "太背了，获得1点精力"}}
	case 2:
		return MessageChain{&Text{Text: "运气还行，获得2点精力"}}
	case 3:
		return MessageChain{&Text{Text: "运气不错，获得3点精力"}}
	default:
		return MessageChain{&Text{Text: fmt.Sprintf("运气爆棚，获得%d点精力", energy)}}
	}
}

type frequency struct{}

func (f *frequency) Name() string {
	return "活跃"
}

func (f *frequency) ShowTips(int64, int64) string {
	return "活跃"
}

func (f *frequency) CheckAuth(int64, int64) bool {
	return true
}

func (f *frequency) Execute(*GroupMessage, string) MessageChain {
	frequencyData, returnError := httpGetData("/frequency", nil)
	if returnError != nil {
		slog.Error("请求失败", "error", returnError.error)
		return returnError.message
	}
	type DataType struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
		Pc    int    `json:"pc"`
	}
	var arr []DataType
	err := json.Unmarshal([]byte(frequencyData.Raw), &arr)
	if err != nil {
		slog.Error("json unmarshal failed", "error", err)
		return MessageChain{&Text{Text: "数据格式错误"}}
	}
	maxDisplayDate, err := time.Parse("2006-01-02", arr[len(arr)-1].Date)
	if err != nil {
		slog.Error("parse date failed", "error", err)
		return MessageChain{&Text{Text: "数据格式错误"}}
	}
	minDisplayDate := maxDisplayDate.AddDate(0, 0, -30)
	var (
		labels       []string
		completeData = [][]float64{{}, {}, {}}
	)
	currentDate := minDisplayDate
	var maxValue float64

	for range 31 {
		dateStr := currentDate.Format("2006-01-02")
		value := slices.IndexFunc(arr, func(value DataType) bool { return value.Date == dateStr })
		labels = append(labels, dateStr[5:])
		if value >= 0 {
			completeData[0] = append(completeData[0], float64(arr[value].Count))
			completeData[1] = append(completeData[1], float64(arr[value].Pc))
			completeData[2] = append(completeData[2], float64(arr[value].Pc-arr[value].Count))
		} else {
			completeData[0] = append(completeData[0], 0)
			completeData[1] = append(completeData[1], 0)
			completeData[2] = append(completeData[2], 0)
		}
		for _, v := range completeData {
			maxValue = max(maxValue, v[len(v)-1])
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	p, err := Render(
		ChartOption{SeriesList: SeriesList{
			NewSeriesFromValues(completeData[0], ChartTypeLine),
			NewSeriesFromValues(completeData[1], ChartTypeLine),
			NewSeriesFromValues(completeData[2], ChartTypeBar),
		}},
		LegendOptionFunc(LegendOption{
			Data: []string{"场次", "参与人次", "活人局系数"},
			Top:  "-30",
		}),
		FontFamilyOptionFunc("simhei"),
		ThemeOptionFunc(ThemeLight),
		PaddingOptionFunc(Box{Top: 40, Left: 10, Right: 45, Bottom: 10}),
		XAxisDataOptionFunc(labels),
		MarkPointOptionFunc(0, SeriesMarkDataTypeMax),
		MarkPointOptionFunc(1, SeriesMarkDataTypeMax),
		MarkLineOptionFunc(0, SeriesMarkDataTypeAverage),
		MarkLineOptionFunc(1, SeriesMarkDataTypeAverage),
		MarkLineOptionFunc(2, SeriesMarkDataTypeAverage),
		func(opt *ChartOption) {
			opt.XAxis.FontSize = 7.5
			opt.XAxis.FirstAxis = -1
			opt.YAxisOptions = []YAxisOption{{
				Max: NewFloatPoint(math.Ceil(float64(maxValue)/50) * 50),
			}}
		},
	)
	if err != nil {
		slog.Error("render chart failed", "error", err)
	} else if buf, err := p.Bytes(); err != nil {
		slog.Error("render chart failed", "error", err)
	} else {
		return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
	}
	return MessageChain{&Text{Text: "render chart failed"}}
}
