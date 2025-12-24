package maplebot

import (
	_ "embed" // 用于嵌入资源文件
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/CuteReimu/YinYangJade/maplebot/scripts"
	. "github.com/CuteReimu/onebot"
	"github.com/tidwall/gjson"
	. "github.com/vicanso/go-charts/v2"
)

type graphData struct {
	CurrentEXP int64  `json:"CurrentEXP"`
	DateLabel  string `json:"DateLabel"`
	Level      int    `json:"Level"`
	ClassID    int    `json:"ClassID"`
}

type findRoleReturnData struct {
	CharacterData *struct {
		CharacterImageURL string      `json:"CharacterImageURL"`
		Image             string      `json:"Image"`
		Class             string      `json:"Class"`
		ClassID           int         `json:"ClassID"`
		EXPPercent        float64     `json:"EXPPercent"`
		GraphData         []graphData `json:"GraphData"`
		LegionLevel       int         `json:"LegionLevel"`
		Level             int         `json:"Level"`
		Name              string      `json:"Name"`
	} `json:"CharacterData"`
}

// FindRoleBackground 在后台预抓取角色数据
func FindRoleBackground() {
	slog.Info("开始角色数据预抓取")
	for _, name := range findRoleData.GetStringMapString("data") {
		output, err := scripts.RunPythonScript("read_player.py", name, "silence")
		if err != nil {
			slog.Error("执行脚本失败", "error", err, "name", name, "output", string(output))
		}
	}
	output, err := scripts.RunPythonScript("scrape.py")
	if err != nil {
		slog.Error("角色数据预抓取失败", "error", err, "output", string(output))
		notifyGroups := config.GetIntSlice("notify_groups")
		notifyQQ := config.GetIntSlice("notify_qq")
		msg := MessageChain{&Text{Text: "角色数据预抓取失败"}}
		for _, qq := range notifyQQ {
			msg = append(msg, &At{QQ: strconv.Itoa(qq)})
		}
		for _, group := range notifyGroups {
			_, _ = bot.SendGroupMessage(int64(group), msg)
		}
		return
	}
	slog.Info("完成角色数据预抓取")
}

func findRole(name string) MessageChain {
	b, err := scripts.RunPythonScript("read_player.py", name)
	if err != nil {
		slog.Error("执行脚本失败", "error", err, "name", name)
	} else {
		var pythonData struct {
			Status  string `json:"status"`
			Profile string `json:"profile"`
			Text    string `json:"text"`
			Chart   string `json:"chart"`
		}
		if err := json.Unmarshal(b, &pythonData); err != nil {
			slog.Error("json解析失败", "error", err, "body", string(b))
		} else if pythonData.Status == "success" {
			var messageChain MessageChain
			if pythonData.Profile != "" {
				messageChain = append(messageChain, &Image{File: "base64://" + pythonData.Profile})
			}
			if pythonData.Text != "" {
				messageChain = append(messageChain, &Text{Text: pythonData.Text})
			}
			if pythonData.Chart != "" {
				messageChain = append(messageChain, &Image{File: "base64://" + pythonData.Chart})
			}
			return messageChain
		}
	}
	var data *findRoleReturnData
	resp, err := restyClient.R().Get("https://api.maplestory.gg/v2/public/character/gms/" + name)
	if err != nil {
		slog.Error("请求失败", "error", err)
	} else {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return MessageChain{&Text{Text: name + "已身死道消"}}
		case http.StatusOK:
			body := resp.Body()
			if err := json.Unmarshal(body, &data); err != nil {
				slog.Error("json解析失败", "error", err, "body", body)
				return MessageChain{&Text{Text: "解析失败"}}
			}
		default:
			slog.Error("请求失败", "status", resp.StatusCode(), "message", gjson.GetBytes(resp.Body(), "message").String())
		}
	}
	if data == nil || data.CharacterData == nil {
		return MessageChain{&Text{Text: "请求失败"}}
	}
	return resolveFindData(data)
}

func resolveFindData(data *findRoleReturnData) MessageChain {
	class := getCharacterClass(data.CharacterData)
	levelExp := fmt.Sprintf("%d (%g%%)", data.CharacterData.Level, data.CharacterData.EXPPercent)
	messageChain := buildCharacterImage(data.CharacterData.Image, data.CharacterData.CharacterImageURL)
	infoText := fmt.Sprintf("角色名：%s\n职业：%s\n等级：%s\n联盟：%d\n",
		data.CharacterData.Name, class, levelExp, data.CharacterData.LegionLevel)

	if len(data.CharacterData.GraphData) == 0 {
		return append(messageChain, &Text{Text: infoText + "近日无经验变化"})
	}

	expValues, levelValues, labels, maybeBurn := processGraphData(data.CharacterData.GraphData)
	if !slices.ContainsFunc(expValues, func(f float64) bool { return f != 0 }) {
		return append(messageChain, &Text{Text: infoText + "近日无经验变化"})
	}

	infoText = addLevelUpPrediction(infoText, levelExp, expValues)
	chart := renderExpChart(expValues, levelValues, labels, maybeBurn)
	if chart != nil {
		return append(messageChain, &Text{Text: infoText}, chart)
	}
	return append(messageChain, &Text{Text: infoText})
}

func getCharacterClass(charData *struct {
	CharacterImageURL string      `json:"CharacterImageURL"`
	Image             string      `json:"Image"`
	Class             string      `json:"Class"`
	ClassID           int         `json:"ClassID"`
	EXPPercent        float64     `json:"EXPPercent"`
	GraphData         []graphData `json:"GraphData"`
	LegionLevel       int         `json:"LegionLevel"`
	Level             int         `json:"Level"`
	Name              string      `json:"Name"`
}) string {
	class := TranslateClassName(charData.Class)
	if len(class) > 0 {
		return class
	}
	class = TranslateClassID(charData.ClassID)
	if len(class) > 0 {
		return class
	}
	for _, d := range slices.Backward(charData.GraphData) {
		if class = TranslateClassID(d.ClassID); len(class) > 0 {
			return class
		}
	}
	return class
}

func buildCharacterImage(img, imgURL string) MessageChain {
	if len(img) > 0 {
		return MessageChain{&Image{File: "base64://" + img}}
	}
	if len(imgURL) > 0 {
		if resp, err := restyClient.R().Get(imgURL); err == nil {
			return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(resp.Body())}}
		}
		slog.Error("请求失败", "url", imgURL)
	}
	return nil
}

func processGraphData(graphData []graphData) (expValues, levelValues []float64, labels []string, maybeBurn bool) {
	levelValues, labels = extractLevelValuesAndLabels(graphData)
	levelValues = fillMissingDates(levelValues, labels, graphData[0].DateLabel)
	labels = labels[1:]
	for i := range labels {
		labels[i] = labels[i][5:]
	}
	levelValues = fixZeroLevels(levelValues)
	maybeBurn = detectBurnEvent(levelValues)
	expValues = calculateExpValues(levelValues, maybeBurn)
	levelValues = adjustForBurn(levelValues[1:], maybeBurn)
	return expValues, levelValues, labels, maybeBurn
}

func extractLevelValuesAndLabels(graphData []graphData) ([]float64, []string) {
	levelValues := make([]float64, 0, 15)
	labels := make([]string, 0, 15)
	for _, d := range graphData {
		labels = append(labels, d.DateLabel)
		levelUpExp := levelExpData.GetFloat64(fmt.Sprintf("data.%d", d.Level))
		var expPercent float64
		if levelUpExp > 0 {
			expPercent = min(float64(d.CurrentEXP)/levelUpExp, 0.999)
		}
		levelValues = append(levelValues, float64(d.Level)+expPercent)
	}
	return levelValues, labels
}

func fillMissingDates(levelValues []float64, labels []string, firstDate string) []float64 {
	if len(labels) >= 15 {
		return levelValues
	}
	t, err := time.Parse("2006-01-02", firstDate)
	for len(labels) < 15 {
		if err != nil {
			labels = append([]string{""}, labels...)
		} else {
			t = t.Add(-24 * time.Hour)
			labels = append([]string{t.Format("2006-01-02")}, labels...)
		}
		levelValues = append(levelValues[:1:1], levelValues...)
	}
	return levelValues
}

func fixZeroLevels(levelValues []float64) []float64 {
	for i := 1; i < len(levelValues); i++ {
		if levelValues[i] == 0 {
			levelValues[i] = levelValues[i-1]
		}
	}
	for i := len(levelValues) - 2; i >= 0; i-- {
		if levelValues[i] == 0 {
			levelValues[i] = levelValues[i+1]
		}
	}
	return levelValues
}

func detectBurnEvent(levelValues []float64) bool {
	return !slices.ContainsFunc(levelValues, func(f float64) bool {
		return int(f) > 220 && int(f) < 260 && int(f)%3 != 2
	})
}

func calculateExpValues(levelValues []float64, maybeBurn bool) []float64 {
	expValues := make([]float64, 0, 14)
	for i := 1; i < len(levelValues); i++ {
		level0, level1 := int(levelValues[i-1]), int(levelValues[i])
		expPercent0, expPercent1 := levelValues[i-1]-float64(level0), levelValues[i]-float64(level1)
		var totalExp float64
		for j := level0; j < level1; j++ {
			if maybeBurn && j > 220 && j < 260 && j%3 != 2 {
				continue
			}
			totalExp += levelExpData.GetFloat64(fmt.Sprintf("data.%d", j))
		}
		totalExp -= expPercent0 * levelExpData.GetFloat64(fmt.Sprintf("data.%d", level0))
		totalExp += expPercent1 * levelExpData.GetFloat64(fmt.Sprintf("data.%d", level1))
		expValues = append(expValues, max(math.Round(totalExp), 0))
	}
	return expValues
}

func adjustForBurn(levelValues []float64, maybeBurn bool) []float64 {
	if !maybeBurn {
		return levelValues
	}
	for i := range levelValues {
		if levelValues[i] < 260 && levelValues[i] >= 221 {
			levelValues[i] += float64((260 - int(levelValues[i])) / 3 * 2)
		} else if levelValues[i] < 221 {
			levelValues[i] += 26
		}
	}
	return levelValues
}

func addLevelUpPrediction(infoText, levelExp string, expValues []float64) string {
	levelStr, _, ok := strings.Cut(levelExp, "(")
	if !ok {
		return infoText
	}
	totalExp := float64(levelExpData.GetInt64("data." + strings.TrimSpace(levelStr)))
	expPercentStart := strings.Index(levelExp, "(")
	expPercentEnd := strings.Index(levelExp, "%")
	if expPercentStart < 0 || expPercentEnd < 0 {
		return infoText
	}
	expPercent, err := strconv.ParseFloat(levelExp[expPercentStart+1:expPercentEnd], 64)
	if err != nil || totalExp <= 0 {
		return infoText
	}
	var sumExp, n float64
	for _, v := range expValues {
		if v != 0 || n > 0 {
			sumExp += v
			n++
		}
	}
	if n == 0 {
		return infoText
	}
	aveExp := sumExp / n
	days := int(math.Ceil((totalExp - totalExp/100.0*expPercent) / aveExp))
	return infoText + fmt.Sprintf("预计还有%d天升级\n", days)
}

func renderExpChart(expValues, levelValues []float64, labels []string, maybeBurn bool) *Image {
	yAxisOptions := calculateYAxisOptions(expValues, levelValues, maybeBurn)
	seriesList := []Series{NewSeriesFromValues(expValues, ChartTypeBar)}
	if len(levelValues) == len(expValues) {
		seriesList = append(seriesList, NewSeriesFromValues(levelValues, ChartTypeLine))
		seriesList[1].AxisIndex = 1
	}
	p, err := Render(
		ChartOption{SeriesList: seriesList},
		ThemeOptionFunc(ThemeDark),
		PaddingOptionFunc(Box{Top: 30, Left: 10, Right: 10, Bottom: 10}),
		XAxisDataOptionFunc(labels),
		YAxisOptionFunc(yAxisOptions...),
		createChartOptionFunc(maybeBurn),
	)
	if err != nil {
		slog.Error("render chart failed", "error", err)
		return nil
	}
	buf, err := p.Bytes()
	if err != nil {
		slog.Error("render chart failed", "error", err)
		return nil
	}
	return &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}
}

func calculateYAxisOptions(expValues, levelValues []float64, maybeBurn bool) []YAxisOption {
	maxValue := max(slices.Max(expValues), 1e11)
	digits := int(math.Floor(math.Log10(maxValue))) + 1
	factor := math.Pow(10, float64(digits-1))
	highest := math.Floor(maxValue / factor)
	if highest < 2 {
		factor /= 5
	} else if highest < 5 {
		factor /= 2
	}
	maxValue = math.Ceil(maxValue/factor) * factor
	divideCount := int(maxValue / factor)
	yAxisOptions := []YAxisOption{{
		Min:         NewFloatPoint(0),
		Max:         &maxValue,
		DivideCount: divideCount,
		Unit:        1,
	}}
	for _, diff := range []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20, 50, 100} {
		minLevel, maxLevel := math.Floor(slices.Min(levelValues)/diff)*diff, math.Ceil(slices.Max(levelValues)/diff)*diff
		remainCount := divideCount - int(math.Round((maxLevel-minLevel)/diff))
		if remainCount >= 0 {
			if remainCount > 0 {
				maxLevel += diff * float64(remainCount-remainCount/2)
				minLevel -= diff * float64(remainCount/2)
			}
			if maybeBurn {
				if minLevel <= 246 {
					maxLevel -= minLevel - 246
					minLevel = 246
				}
			} else if minLevel <= 220 {
				maxLevel -= minLevel - 220
				minLevel = 220
			}
			yAxisOptions = append(yAxisOptions, YAxisOption{
				Min:         NewFloatPoint(minLevel),
				Max:         NewFloatPoint(maxLevel),
				DivideCount: divideCount,
				Unit:        1,
			})
			break
		}
	}
	return yAxisOptions
}

func createChartOptionFunc(maybeBurn bool) func(*ChartOption) {
	return func(opt *ChartOption) {
		opt.XAxis.TextRotation = -math.Pi / 4
		opt.XAxis.LabelOffset = Box{Top: -5, Left: 7}
		opt.XAxis.FontSize = 7.5
		opt.ValueFormatter = func(f float64) string {
			switch {
			case f == 0:
				return "0"
			case f < 260 && maybeBurn:
				f -= float64((260 - int(f)) * 2)
				fallthrough
			case f < 1000.0:
				if f == 220.0 || f == 218.0 {
					return "\u3000"
				}
				return fmt.Sprintf("%g", math.Round(f*1000)/1000)
			case f < 1000000.0:
				return fmt.Sprintf("%gK", math.Round(f)/1000)
			case f < 1000000000.0:
				return fmt.Sprintf("%gM", math.Round(f)/1000000)
			case f < 1000000000000.0:
				return fmt.Sprintf("%gB", math.Round(f)/1000000000)
			default:
				return fmt.Sprintf("%gT", math.Round(f)/1000000000000)
			}
		}
	}
}
