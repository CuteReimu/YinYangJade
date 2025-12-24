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
		availableInputs: []string{"Any%", "TE", "100%", "Judgement", "Low%", "AB", "Twisted%", "苔穴", "PoP", "白宫"},
		aliasInputs:     []string{"all bosses", "all boss", "allbosses", "allboss", "苦痛", "苦痛之路", "白色宫殿"},
		url: map[string]string{
			"anylp":     "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/zd39j4nd?var-ylq4yvzn=qzne828q&var-rn1kmmvl=qj70747q",
			"anyrp":     "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/zd39j4nd?var-ylq4yvzn=qzne828q&var-rn1kmmvl=10vzvmol",
			"te":        "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/n2y0m18d?var-dloed1dn=qyzod221",
			"100noab":   "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/rkl6zprk?var-rn1k7xol=lx5o7641&var-38dg4448=1w4p4dmq",
			"100ab":     "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/rkl6zprk?var-rn1k7xol=lx5o7641&var-38dg4448=qoxpx35q",
			"judgement": "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wk6544o2?var-jlz631q8=1w4ozxvq",
			"low":       "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/wkp4r60k?var-9l7geqpl=1397dnx1",
			"ab":        "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/w206ox52?var-kn0eyxz8=10vzo8wl",
			"twisted":   "https://www.speedrun.com/api/v1/leaderboards/y65r7g81/category/9kvvl0ok?var-yn26pzel=le2z97kl",
			"苔穴":        "https://www.speedrun.com/api/v1/leaderboards/yd4r2x51/level/9m58yezd/xd1ypjwd?var-r8r69958=qvvpvrrq",
			"pop":       "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/r9g1qop9/wkpq608d",
			"白色宫殿":      "https://www.speedrun.com/api/v1/leaderboards/76rqmld8/level/69znevg9/wkpq608d?var-r8r11k7n=klr8rr21",
		},
		categoryNames: map[string]string{
			"anylp":     "丝之歌 — Any% 新",
			"anyrp":     "丝之歌 — Any% 旧",
			"te":        "丝之歌 — True Ending",
			"100noab":   "丝之歌 — 100% No AB",
			"100ab":     "丝之歌 — 100% All Bosses",
			"judgement": "丝之歌 — Judgement",
			"low":       "丝之歌 — Low%",
			"ab":        "丝之歌 — All Bosses",
			"twisted":   "丝之歌 — Twisted%",
			"苔穴":        "丝之歌 — 苔穴",
			"pop":       "空洞骑士 — 苦痛之路",
			"白色宫殿":      "空洞骑士 — 白色宫殿 1.3.1.5+",
		},
		mapKeys: map[string][]string{
			"any":        {"anyrp", "anylp"},
			"100":        {"100noab", "100ab"},
			"all bosses": {"ab"},
			"all boss":   {"ab"},
			"allbosses":  {"ab"},
			"allboss":    {"ab"},
			"苦痛之路":       {"pop"},
			"苦痛":         {"pop"},
			"白宫":         {"白色宫殿"},
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

func (t *speedrunLeaderboards) Name() string {
	return "查榜"
}

func (t *speedrunLeaderboards) ShowTips(int64, int64) string {
	return "查榜"
}

func (t *speedrunLeaderboards) CheckAuth(int64, int64) bool {
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

type speedrunApiResp struct {
	Data struct {
		Runs []struct {
			Place int `json:"place"`
			Run   struct {
				Times struct {
					PrimaryT float64 `json:"primary_t"`
				} `json:"times"`
				Players []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"players"`
				Date string `json:"date"`
			} `json:"run"`
		} `json:"runs"`
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

func (t *speedrunLeaderboards) getPlayerName(refID string, players []speedrunPlayerData, refName string) string {
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
	var ar speedrunApiResp
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
	for _, entry := range runs {
		place := entry.Place
		run := entry.Run
		timeStr := formatTime(run.Times.PrimaryT)
		playerRef := run.Players
		playerName := "Unknown"
		if len(playerRef) > 0 {
			playerName = t.getPlayerName(playerRef[0].ID, players, playerRef[0].Name)
		}
		rel := ""
		if run.Date != "" {
			rel = " — " + formatRelativeDate(run.Date)
		}
		_, _ = fmt.Fprintf(&buf, "%d. %s — %s%s\r\n", place, playerName, timeStr, rel)
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
	return strings.Join(ss, ""), nil
}
