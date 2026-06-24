//nolint:all
package hkbot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
)

var _srcNameDict = map[string]string{
	"Hollow Knight":                               "空洞骑士",
	"Hollow Knight Category Extensions":           "空洞骑士副榜",
	"Hollow Knight: Silksong":                     "丝之歌",
	"Hollow Knight: Silksong Category Extensions": "丝之歌副榜",
}

func srcTryTranslate(s string) string {
	if v, ok := _srcNameDict[s]; ok {
		return v
	}
	return translateDict.ReplaceAll(s)
}

func init() {
	cmd := &speedrunPersonalBests{
		resty: resty.New(),
	}
	cmd.resty.SetTimeout(time.Minute)
	addCmdListener(cmd)
}

type speedrunPersonalBests struct {
	resty *resty.Client
}

func (*speedrunPersonalBests) Name() string {
	return "查个人"
}

func (*speedrunPersonalBests) ShowTips(int64, int64) string {
	return "查个人"
}

func (*speedrunPersonalBests) CheckAuth(int64, int64) bool {
	return true
}

func (t *speedrunPersonalBests) Execute(msg *GroupMessage, content string) MessageChain {
	return MessageChain{&Text{Text: "请在命令前加 / ，例如 /查个人"}}
	if content == "" {
		return MessageChain{&Text{Text: "用法：查个人 <用户名>\r\n示例：查个人 SclicheD"}}
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err)
			}
		}()
		text, err := t.getPersonalBestsText(content)
		if err != nil {
			slog.Error("查询个人 PB 失败", "error", err)
			sendGroupMessage(msg, &Text{Text: "查询失败"})
			return
		}
		sendGroupMessage(msg, &Text{Text: text})
	}()
	return nil
}

// ---------------------------------------------------------------------------
// API structs
// ---------------------------------------------------------------------------

type srcUserResp struct {
	Data []struct {
		ID    string `json:"id"`
		Names struct {
			International string `json:"international"`
		} `json:"names"`
		Signup string `json:"signup"`
	} `json:"data"`
}

type srcPBResp struct {
	Data       []srcPBEntry `json:"data"`
	Pagination struct {
		Links []struct {
			Rel string `json:"rel"`
		} `json:"links"`
	} `json:"pagination"`
}

type srcPBEntry struct {
	Place int `json:"place"`
	Run   struct {
		Level string `json:"level"`
		Times struct {
			PrimaryT float64 `json:"primary_t"`
		} `json:"times"`
		Values map[string]string `json:"values"`
	} `json:"run"`
	Game struct {
		Data struct {
			Names struct {
				International string `json:"international"`
			} `json:"names"`
		} `json:"data"`
	} `json:"game"`
	Category json.RawMessage `json:"category"`
	Level    json.RawMessage `json:"level"`
}

type categoryData struct {
	Data *struct {
		Name      string `json:"name"`
		Variables struct {
			Data []srcVariable `json:"data"`
		} `json:"variables"`
	} `json:"data"`
}

type srcVariable struct {
	ID            string `json:"id"`
	IsSubcategory bool   `json:"is-subcategory"`
	Values        struct {
		Values map[string]struct {
			Label string `json:"label"`
		} `json:"values"`
	} `json:"values"`
}

// ---------------------------------------------------------------------------
// API methods
// ---------------------------------------------------------------------------

func (t *speedrunPersonalBests) getUser(username string) (id, name, signup string, err error) {
	resp, err := t.resty.R().
		SetQueryParam("lookup", username).
		Get("https://www.speedrun.com/api/v1/users")
	if err != nil {
		return "", "", "", err
	}
	if resp.StatusCode() != 200 {
		return "", "", "", fmt.Errorf("获取用户失败: %d", resp.StatusCode())
	}
	var ur srcUserResp
	if err := json.Unmarshal(resp.Body(), &ur); err != nil {
		return "", "", "", err
	}
	if len(ur.Data) == 0 {
		return "", "", "", nil
	}
	u := ur.Data[0]
	return u.ID, u.Names.International, u.Signup, nil
}

func (t *speedrunPersonalBests) getUserPBs(userID string) ([]srcPBEntry, error) {
	var all []srcPBEntry
	offset := 0
	for {
		resp, err := t.resty.R().
			SetQueryParams(map[string]string{
				"max":    "200",
				"offset": strconv.Itoa(offset),
				"embed":  "game,category,level,level.variables,category.variables",
			}).
			Get(fmt.Sprintf("https://www.speedrun.com/api/v1/users/%s/personal-bests", userID))
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("获取 PB 失败: %d", resp.StatusCode())
		}
		var pr srcPBResp
		if err := json.Unmarshal(resp.Body(), &pr); err != nil {
			return nil, err
		}
		if len(pr.Data) == 0 {
			break
		}
		all = append(all, pr.Data...)
		hasNext := false
		for _, link := range pr.Pagination.Links {
			if link.Rel == "next" {
				hasNext = true
				break
			}
		}
		if !hasNext {
			break
		}
		offset += 200
	}
	return all, nil
}

// ---------------------------------------------------------------------------
// Time formatting
// ---------------------------------------------------------------------------

func srcFormatTime(seconds float64) string {
	if seconds < 0 {
		return "00:00.000"
	}
	totalMs := int(seconds*1000 + 0.5)
	hours := totalMs / 3_600_000
	minutes := (totalMs % 3_600_000) / 60_000
	secs := (totalMs % 60_000) / 1000
	ms := totalMs % 1000
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	if minutes >= 10 {
		return fmt.Sprintf("%02d:%02d", minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d.%03d", minutes, secs, ms)
}

// ---------------------------------------------------------------------------
// Process PB entry
// ---------------------------------------------------------------------------

type srcPBResult struct {
	rank     int
	game     string
	category string
	timeStr  string
}

func processPBEntry(pb srcPBEntry) (*srcPBResult, bool) {
	gameName := pb.Game.Data.Names.International
	if !strings.Contains(gameName, "Hollow Knight") {
		return nil, false
	}

	var fullName strings.Builder
	var variables []srcVariable
	var categoryBuf json.RawMessage
	if pb.Run.Level != "" {
		categoryBuf = pb.Level
	} else {
		categoryBuf = pb.Category
	}
	var category categoryData
	if err := json.Unmarshal(categoryBuf, &category); err != nil {
		slog.Error("解析 PB 记录失败", "error", err)
		return nil, false
	}
	_, _ = fullName.WriteString(category.Data.Name)
	variables = category.Data.Variables.Data

	for key, value := range pb.Run.Values {
		for _, variable := range variables {
			if variable.ID == key && variable.IsSubcategory {
				if v, ok := variable.Values.Values[value]; ok && v.Label != "" {
					_ = fullName.WriteByte(' ')
					_, _ = fullName.WriteString(v.Label)
				}
				break
			}
		}
	}

	return &srcPBResult{
		rank:     pb.Place,
		game:     gameName,
		category: fullName.String(),
		timeStr:  srcFormatTime(pb.Run.Times.PrimaryT),
	}, true
}

// ---------------------------------------------------------------------------
// Text generation
// ---------------------------------------------------------------------------

func (t *speedrunPersonalBests) getPersonalBestsText(username string) (string, error) {
	userID, userName, signupDate, err := t.getUser(username)
	if err != nil {
		return "", err
	}
	if userID == "" {
		return "❌ 未找到用户", nil
	}

	pbs, err := t.getUserPBs(userID)
	if err != nil {
		return "", err
	}
	if len(pbs) == 0 {
		return "❌ 该用户暂无 PB", nil
	}

	var results []*srcPBResult
	for _, pb := range pbs {
		if r, ok := processPBEntry(pb); ok {
			results = append(results, r)
		}
	}
	if len(results) == 0 {
		return "❌ 记录处理失败", nil
	}

	// sort by rank
	slices.SortFunc(results, func(a, b *srcPBResult) int {
		ra, rb := a.rank, b.rank
		if ra == 0 {
			ra = 999999
		}
		if rb == 0 {
			rb = 999999
		}
		return ra - rb
	})

	var lines []string
	lines = append(lines, "👤 用户: "+userName)
	if len(signupDate) >= 10 {
		lines = append(lines, "📅 注册: "+signupDate[:10])
	}
	lines = append(lines, "")

	limit := min(len(results), 50)
	for _, r := range results[:limit] {
		game := srcTryTranslate(r.game)
		category := srcTryTranslate(r.category)

		rankDisplay := fmt.Sprintf("#%d", r.rank)
		if r.rank == 0 {
			rankDisplay = "N/A"
		}
		if utf8.RuneCountInString(game) > 25 {
			game = string([]rune(game)[:25]) + ".."
		}
		if utf8.RuneCountInString(category) > 60 {
			category = string([]rune(category)[:60]) + ".."
		}
		lines = append(lines, fmt.Sprintf("%s %s %s %s", rankDisplay, game, category, r.timeStr))
	}
	if len(results) > 50 {
		lines = append(lines, "...")
	}

	// QQ 长度限制
	var finalLines []string
	currentLength := 0
	const maxLength = 3800
	for _, line := range lines {
		if currentLength+len(line)+1 > maxLength {
			finalLines = append(finalLines, "")
			finalLines = append(finalLines, "...(内容过长已截断)")
			break
		}
		finalLines = append(finalLines, line)
		currentLength += len(line) + 1
	}

	return strings.Join(finalLines, "\r\n"), nil
}
