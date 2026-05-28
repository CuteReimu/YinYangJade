package hkbot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/singleflight"
)

func init() {
	cmd := &speedrunLeaderboards{
		availableInputs: []string{"Any%", "HKAny%", "TE", "HKTE", "HKLow", "HKAB", "100%", "112%", "107%", "全技能",
			"全成就", "钢魂", "GE", "第一幕", "Low%", "AB", "Twisted%", "苔穴", "PoP", "白色宫殿",
			"一门", "二门", "三门", "四门", "五门", "jjc"},
		aliasInputs: []string{"all bosses", "all boss", "allbosses", "allboss", "苦痛", "苦痛之路", "白宫", "act1", "as", "aa", "ss",
			"jjc1", "jjc2", "jjc3"},
		url: map[string]string{
			"hkany":            "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/02q8o4p2?var-yn2p3085=21gyy061",
			"hklow":            "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/w20w0v5d?var-5lyjjd2l=4lxz6641",
			"aa":               "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/q25epyg2?var-onv7r95n=21g8poml",
			"ss":               "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/wkp31j02?var-e8mrpyxl=814jwwv1",
			"hkab":             "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/824m6ng2?var-wle6d0x8=10v97vwl&var-e8m1ye86=xqkomxk1",
			"ge":               "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/8241w7w2?var-5ly7kkkl=z1983n8q",
			"hkte":             "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/wk617wxd?var-jlz32x82=5leope5q",
			"as":               "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/n2y577zk?var-38dopp1l=4lxogy4l",
			"107":              "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/vdo5xe6k?var-ql6165x8=jq6w78nl",
			"112":              "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/category/xk9vrl6d?var-onvj96mn=5q870z6l",
			"anylp":            "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/zd39j4nd?var-ylq4yvzn=qzne828q&var-rn1kmmvl=qj70747q",
			"anyrp":            "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/zd39j4nd?var-ylq4yvzn=qzne828q&var-rn1kmmvl=10vzvmol",
			"te":               "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/n2y0m18d?var-dloed1dn=qyzod221",
			"100noab":          "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/rkl6zprk?var-rn1k7xol=lx5o7641&var-38dg4448=1w4p4dmq",
			"100ab":            "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/rkl6zprk?var-rn1k7xol=lx5o7641&var-38dg4448=qoxpx35q",
			"judgement":        "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wk6544o2?var-jlz631q8=1w4ozxvq&var-j8415y58=qox0j35q",
			"sinner":           "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wk6544o2?var-jlz631q8=1w4ozxvq&var-j8415y58=1390xmr1",
			"lowst":            "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wkp4r60k?var-9l7geqpl=1397dnx1&var-p850r65n=1923x9yq",
			"lowte":            "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wkp4r60k?var-9l7geqpl=1397dnx1&var-p850r65n=12vn9wdq",
			"abact1":           "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/w206ox52?var-kn0eyxz8=10vzo8wl&var-rn16z5p8=qvv7xrrq",
			"abact2":           "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/w206ox52?var-kn0eyxz8=10vzo8wl&var-rn16z5p8=le2on4ml",
			"abact3":           "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/w206ox52?var-kn0eyxz8=10vzo8wl&var-rn16z5p8=q5v6457l",
			"twisted":          "https://www.speedrun.com/api/v1/leaderboards/yd4r2x51/category/5dwm145k?var-wle492kn=10v8m0pl",
			"苔穴":             "https://www.speedrun.com/api/v1/leaderboards/yd4r2x51/level/9m58yezd/xd1ypjwd?var-r8r69958=qvvpvrrq",
			"pop":              "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/r9g1qop9/wkpq608d",
			"白色宫殿":         "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/69znevg9/wkpq608d?var-r8r11k7n=klr8rr21",
			"一门任意锁":       "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/495lx03d/wkpq608d?var-5lyppm9l=klrnokw1",
			"一门任意锁无亡怒": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/495lx03d/wkpq608d?var-5lyppm9l=10vg8d5l",
			"一门四锁":         "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/495lx03d/wkpq608d?var-5lyppm9l=21drx8gq",
			"一门四锁当前版本": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/495lx03d/wkpq608d?var-5lyppm9l=jqz78pkl",
			"二门任意锁":       "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/o9x3rvp9/wkpq608d?var-e8mrrwql=5q85wv6q",
			"二门任意锁无亡怒": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/o9x3rvp9/wkpq608d?var-e8mrrwql=qj7vr8gq",
			"二门四锁":         "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/o9x3rvp9/wkpq608d?var-e8mrrwql=4qy3g2dl",
			"二门四锁当前版本": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/o9x3rvp9/wkpq608d?var-e8mrrwql=81pwvpnl",
			"三门任意锁":       "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/rdq54v2d/wkpq608d?var-ylqmme3n=mlnrjvnq",
			"三门任意锁无亡怒": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/rdq54v2d/wkpq608d?var-ylqmme3n=q65nme7l",
			"三门四锁":         "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/rdq54v2d/wkpq608d?var-ylqmme3n=8103knp1",
			"三门四锁当前版本": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/rdq54v2d/wkpq608d?var-ylqmme3n=xqkk2j4q",
			"四门任意锁":       "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/5d7zqm6w/wkpq608d?var-gnxvvy6l=9qjjkmoq",
			"四门任意锁无亡怒": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/5d7zqm6w/wkpq608d?var-gnxvvy6l=lmovym41",
			"四门四锁":         "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/5d7zqm6w/wkpq608d?var-gnxvvy6l=jq636ooq",
			"四门四锁当前版本": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/5d7zqm6w/wkpq608d?var-gnxvvy6l=gq7e0jr1",
			"五门任意锁":       "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/kwj14q7w/wkpq608d?var-dloyyge8=5lmz030q",
			"五门任意锁无亡怒": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/kwj14q7w/wkpq608d?var-dloyyge8=1w4zg05q",
			"五门四锁":         "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/kwj14q7w/wkpq608d?var-dloyyge8=81w3x26l",
			"五门四锁当前版本": "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/kwj14q7w/wkpq608d?var-dloyyge8=21grv8oq",
			"jjc1":             "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/gdr16vlw/wkpq608d",
			"jjc2":             "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/nwlp4ve9/wkpq608d",
			"jjc3":             "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/ywemx77d/wkpq608d",
		},
		categoryNames: map[string]string{
			"hkany":            "空洞骑士 — Any% 当前版本",
			"hklow":            "空洞骑士 — Low% 当前版本",
			"aa":               "空洞骑士 — 全成就",
			"ss":               "空洞骑士 — 钢魂Any% 当前版本",
			"hkab":             "空洞骑士 — 全Boss 生命血版本",
			"ge":               "空洞骑士 — 神居结局",
			"hkte":             "空洞骑士 — 真结局",
			"as":               "空洞骑士 — 全技能",
			"107":              "空洞骑士 — 107%AB",
			"112":              "空洞骑士 — 112%APB",
			"anylp":            "丝之歌 — Any% 斗篷",
			"anyrp":            "丝之歌 — Any% 无斗篷",
			"te":               "丝之歌 — True Ending",
			"100noab":          "丝之歌 — 100% No AB",
			"100ab":            "丝之歌 — 100% All Bosses",
			"judgement":        "丝之歌 — 第一幕 - 末代裁决者",
			"sinner":           "丝之歌 — 第一幕 - 罪途",
			"lowst":            "丝之歌 — Low%",
			"lowte":            "丝之歌 — Low% True Ending",
			"abact1":           "丝之歌 — All Bosses - Act1",
			"abact2":           "丝之歌 — All Bosses - Act2",
			"abact3":           "丝之歌 — All Bosses - Act3",
			"twisted":          "丝之歌 — Twisted%",
			"苔穴":             "丝之歌 — 苔穴",
			"pop":              "空洞骑士 — 苦痛之路",
			"白色宫殿":         "空洞骑士 — 白色宫殿 1.3.1.5+",
			"一门任意锁":       "空洞骑士 — 大师万神殿 — 任意锁",
			"一门任意锁无亡怒": "空洞骑士 — 大师万神殿 — 任意锁 无亡怒",
			"一门四锁":         "空洞骑士 — 大师万神殿 — 四锁",
			"一门四锁当前版本": "空洞骑士 — 大师万神殿 — 四锁 当前版本",
			"二门任意锁":       "空洞骑士 — 艺术家万神殿 — 任意锁",
			"二门任意锁无亡怒": "空洞骑士 — 艺术家万神殿 — 任意锁 无亡怒",
			"二门四锁":         "空洞骑士 — 艺术家万神殿 — 四锁",
			"二门四锁当前版本": "空洞骑士 — 艺术家万神殿 — 四锁 当前版本",
			"三门任意锁":       "空洞骑士 — 贤者万神殿 — 任意锁",
			"三门任意锁无亡怒": "空洞骑士 — 贤者万神殿 — 任意锁 无亡怒",
			"三门四锁":         "空洞骑士 — 贤者万神殿 — 四锁",
			"三门四锁当前版本": "空洞骑士 — 贤者万神殿 — 四锁 当前版本",
			"四门任意锁":       "空洞骑士 — 骑士万神殿 — 任意锁",
			"四门任意锁无亡怒": "空洞骑士 — 骑士万神殿 — 任意锁 无亡怒",
			"四门四锁":         "空洞骑士 — 骑士万神殿 — 四锁",
			"四门四锁当前版本": "空洞骑士 — 骑士万神殿 — 四锁 当前版本",
			"五门任意锁":       "空洞骑士 — 圣巢万神殿 — 任意锁",
			"五门任意锁无亡怒": "空洞骑士 — 圣巢万神殿 — 任意锁 无亡怒",
			"五门四锁":         "空洞骑士 — 圣巢万神殿 — 四锁",
			"五门四锁当前版本": "空洞骑士 — 圣巢万神殿 — 四锁 当前版本",
			"jjc1":             "空洞骑士 — 勇士的试炼",
			"jjc2":             "空洞骑士 — 征服者的试炼",
			"jjc3":             "空洞骑士 — 愚人的试炼",
		},
		mapKeys: map[string][]string{
			"any":        {"anyrp", "anylp"},
			"100":        {"100noab", "100ab"},
			"low":        {"lowst", "lowte"},
			"ab":         {"abact1", "abact2", "abact3"},
			"all bosses": {"abact1", "abact2", "abact3"},
			"all boss":   {"abact1", "abact2", "abact3"},
			"allbosses":  {"abact1", "abact2", "abact3"},
			"allboss":    {"abact1", "abact2", "abact3"},
			"苦痛之路":   {"pop"},
			"苦痛":       {"pop"},
			"白宫":       {"白色宫殿"},
			"act1":       {"judgement", "sinner"},
			"第一幕":     {"judgement", "sinner"},
			"全技能":     {"as"},
			"全成就":     {"aa"},
			"钢魂":       {"ss"},
			"一门":       {"一门任意锁", "一门任意锁无亡怒", "一门四锁", "一门四锁当前版本"},
			"二门":       {"二门任意锁", "二门任意锁无亡怒", "二门四锁", "二门四锁当前版本"},
			"三门":       {"三门任意锁", "三门任意锁无亡怒", "三门四锁", "三门四锁当前版本"},
			"四门":       {"四门任意锁", "四门任意锁无亡怒", "四门四锁", "四门四锁当前版本"},
			"五门":       {"五门任意锁", "五门任意锁无亡怒", "五门四锁", "五门四锁当前版本"},
			"jjc":        {"jjc1", "jjc2", "jjc3"},
		},
		resty: resty.New(),
	}
	cmd.resty.SetTimeout(time.Minute)
	addCmdListener(cmd)
}

type speedrunLeaderboards struct {
	availableInputs []string
	aliasInputs     []string
	sf              singleflight.Group
	url             map[string]string
	categoryNames   map[string]string
	mapKeys         map[string][]string
	resty           *resty.Client
}

func (*speedrunLeaderboards) Name() string {
	return "查榜"
}

func (*speedrunLeaderboards) ShowTips(int64, int64) string {
	return "查榜"
}

func (*speedrunLeaderboards) CheckAuth(int64, int64) bool {
	return true
}

func (t *speedrunLeaderboards) Execute(msg *GroupMessage, content string) MessageChain {
	content = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(content, "%", ""), "％", ""))
	eq := func(s string) bool {
		return strings.EqualFold(content, strings.ReplaceAll(s, "%", ""))
	}
	if !slices.ContainsFunc(t.availableInputs, eq) && !slices.ContainsFunc(t.aliasInputs, eq) {
		return MessageChain{&Text{Text: "支持的榜单类型有：" + strings.Join(t.availableInputs, "，")}}
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err)
			}
		}()
		result, err, _ := t.sf.Do(content, func() (any, error) {
			return t.requestSpeedrun(content)
		})
		if err != nil {
			slog.Error("查询失败", "output", result, "error", err)
			return
		}
		output, ok := result.(string)
		if !ok {
			slog.Error("查询结果类型错误", "type", fmt.Sprintf("%T", result))
			return
		}
		sendGroupMessage(msg, &Text{Text: strings.TrimSpace(output)})
	}()
	return nil
}

type speedrunPlayerData struct {
	ID    string `json:"id"`
	Names struct {
		International string `json:"international"`
	} `json:"names"`
}

type speedrunRunData struct {
	Place int `json:"place"`
	Run   struct {
		Times struct {
			PrimaryT float64 `json:"primary_t"`
		} `json:"times"`
		Players []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"players"`
		Date      string `json:"date"`
		Submitted string `json:"submitted"`
		Status    struct {
			VerifyDate string `json:"verify-date"`
		} `json:"status"`
	} `json:"run"`
}

type speedrunAPIResp struct {
	Data struct {
		Runs    []speedrunRunData `json:"runs"`
		Players struct {
			Data []speedrunPlayerData `json:"data"`
		} `json:"players"`
	} `json:"data"`
}

func formatTime(tm float64) string {
	m := int(tm) / 60
	sFloat := tm - float64(m*60)
	h := m / 60
	if h > 0 {
		m = m - h*60
		return fmt.Sprintf("%d:%02d:%02d", h, m, int(sFloat))
	}
	if m < 10 {
		return fmt.Sprintf("%d:%06.3f", m, sFloat)
	}
	return fmt.Sprintf("%02d:%02d", m, int(sFloat))
}

func formatRelativeDate(dateStr string) string {
	if dateStr == "" {
		return dateStr
	}
	tm, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	days := int(today.Sub(tm).Hours() / 24)
	switch {
	case days == 0:
		return "今天"
	case days == 1:
		return "昨天"
	case days == 2:
		return "前天"
	case days >= 3 && days < 30:
		return fmt.Sprintf("%d天前", days)
	case days >= 30 && days < 60:
		return "上个月"
	case days >= 60:
		months := days / 30
		if months >= 12 {
			return fmt.Sprintf("%d年前", months/12)
		}
		return fmt.Sprintf("%d个月前", months)
	default:
		// 未来日期或其他情况，返回原字符串
		return dateStr
	}
}

func getPlayerName(refID string, players []speedrunPlayerData, refName string) string {
	for _, p := range players {
		if p.ID == refID {
			return p.Names.International
		}
	}
	if refName != "" {
		return refName
	}
	return "Unknown"
}

func (t *speedrunLeaderboards) fetch(key string) (string, error) {
	u, ok := t.url[key]
	if !ok {
		return "", fmt.Errorf("未知的分类: %s", key)
	}
	if strings.Contains(u, "?") {
		u += "&embed=players&top=5"
	} else {
		u += "?embed=players&top=5"
	}
	resp, err := t.resty.R().Get(u)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() >= 400 {
		return "", fmt.Errorf("http 错误: %d", resp.StatusCode())
	}
	var ar speedrunAPIResp
	if err := json.Unmarshal(resp.Body(), &ar); err != nil {
		return "", err
	}
	runs := ar.Data.Runs
	if len(runs) > 5 {
		runs = runs[:5]
	}
	players := ar.Data.Players.Data

	var buf strings.Builder
	_, _ = fmt.Fprintf(&buf, "=== %s — NMG ===\r\n", t.categoryNames[key])
	if len(runs) == 0 {
		buf.WriteString("暂无记录\r\n")
		return buf.String(), nil
	}
	var totalDiffDate, totalDate float64
	for _, entry := range runs {
		place := entry.Place
		run := entry.Run
		timeStr := formatTime(run.Times.PrimaryT)
		playerRef := run.Players
		playerName := "Unknown"
		if len(playerRef) > 0 {
			playerName = getPlayerName(playerRef[0].ID, players, playerRef[0].Name)
		}
		rel := ""
		if run.Date != "" {
			rel = " — " + formatRelativeDate(run.Date)
		}
		if run.Submitted != "" && run.Status.VerifyDate != "" {
			submittedTime, err1 := time.Parse(time.RFC3339, run.Submitted)
			verifyDate, err2 := time.Parse(time.RFC3339, run.Status.VerifyDate)
			if err1 == nil && err2 == nil {
				diff := verifyDate.Sub(submittedTime).Hours() / 24
				totalDiffDate += diff
				totalDate++
			}
		}
		_, _ = fmt.Fprintf(&buf, "%d. %s — %s%s\r\n", place, playerName, timeStr, rel)
	}
	if totalDate > 0 {
		if totalDiffDate/totalDate >= 0.5 {
			_, _ = fmt.Fprintf(&buf, "平均审核时间: %.0f天\r\n", totalDiffDate/totalDate)
		} else {
			_, _ = fmt.Fprint(&buf, "平均审核时间: 小于1天\r\n")
		}
	}
	return buf.String(), nil
}

func (t *speedrunLeaderboards) requestSpeedrun(arg string) (string, error) {
	keys, ok := t.mapKeys[arg]
	if !ok {
		keys = []string{arg}
	}

	ss := make([]string, 0, len(keys))
	for _, k := range keys {
		s, err := t.fetch(k)
		if err != nil {
			return "", err
		}
		ss = append(ss, s)
	}
	return strings.Join(ss, "\r\n"), nil
}
