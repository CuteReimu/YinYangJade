// 星之力计算器的工具类，相关数据取自:
//
// https://strategywiki.org/wiki/MapleStory/Spell_Trace_and_Star_Force#Meso_Cost

package maplebot

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/CuteReimu/YinYangJade/maplebot/scripts"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
)

func makeMesoFn(divisor int, currentStarExp0 ...float64) func(int, int) float64 {
	currentStarExp := 2.7
	if len(currentStarExp0) > 0 {
		currentStarExp = currentStarExp0[0]
	}
	return func(currentStar, itemLevel int) float64 {
		itemLevel3 := float64(itemLevel * itemLevel * itemLevel)
		exp := math.Pow(float64(currentStar+1), currentStarExp)
		return 100*math.Round(itemLevel3*exp/float64(divisor)) + 10
	}
}

func preSaviorMesoFn(currentStar int) func(int, int) float64 {
	switch {
	case currentStar >= 15:
		return makeMesoFn(20000)
	case currentStar >= 10:
		return makeMesoFn(40000)
	default:
		return makeMesoFn(5000, 1.0)
	}
}

func saviorMesoFn(newKms bool, currentStar int) func(int, int) float64 {
	if newKms {
		switch currentStar {
		case 17:
			return makeMesoFn(15000)
		case 18:
			return makeMesoFn(7000)
		case 19:
			return makeMesoFn(4500)
		case 21:
			return makeMesoFn(12500)
		}
	}
	switch currentStar {
	case 11:
		return makeMesoFn(22000)
	case 12:
		return makeMesoFn(15000)
	case 13:
		return makeMesoFn(11000)
	case 14:
		return makeMesoFn(7500)
	default:
		return preSaviorMesoFn(currentStar)
	}
}

func saviorCost(newKms bool, currentStar, itemLevel int) float64 {
	return saviorMesoFn(newKms, currentStar)(currentStar, itemLevel)
}

func attemptCost(newKms bool, currentStar int, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen, chanceTime bool) float64 {
	multiplier := 1.0
	if boomProtect && !(fiveTenFifteen && currentStar == 15) && !chanceTime && (currentStar == 15 || currentStar == 16) {
		if newKms {
			multiplier += 2.0
		} else {
			multiplier += 1.0
		}
	}
	if thirtyOff {
		multiplier -= 0.3
	}
	cost := saviorCost(newKms, currentStar, itemLevel) * multiplier
	return math.Round(cost)
}

// determineOutcome return either _SUCCESS, _MAINTAIN, _DECREASE, or _BOOM
func determineOutcome(newKms bool, currentStar int, boomProtect, fiveTenFifteen, boomEvent bool) starForceResult {
	if fiveTenFifteen && (currentStar == 5 || currentStar == 10 || currentStar == 15) {
		return _SUCCESS
	}
	rates := rates
	if newKms {
		rates = rates2
	}
	var (
		outcome             = rand.Float64()
		probabilitySuccess  = rates[currentStar][_SUCCESS]
		probabilityMaintain = rates[currentStar][_MAINTAIN]
		probabilityDecrease = rates[currentStar][_DECREASE]
		probabilityBoom     = rates[currentStar][_BOOM]
	)
	if boomEvent && currentStar <= 21 {
		probabilityMaintain += probabilityBoom * 0.3
		probabilityBoom *= 0.7
	}
	if boomProtect && (newKms && currentStar <= 17 || currentStar <= 16) { // boom protection enabled
		if probabilityDecrease > 0 {
			probabilityDecrease += probabilityBoom
		} else {
			probabilityMaintain += probabilityBoom
		}
		probabilityBoom = 0.0
	}
	// star catch adjustment
	probabilitySuccess *= 1.05
	leftOver := 1 - probabilitySuccess
	if probabilityDecrease == 0.0 {
		probabilityMaintain *= leftOver / (probabilityMaintain + probabilityBoom)
		probabilityBoom = leftOver - probabilityMaintain
	} else {
		probabilityDecrease *= leftOver / (probabilityDecrease + probabilityBoom)
		probabilityBoom = leftOver - probabilityDecrease
	}
	if outcome < probabilitySuccess {
		return _SUCCESS
	} else if outcome < probabilitySuccess+probabilityMaintain {
		return _MAINTAIN
	} else if outcome < probabilitySuccess+probabilityMaintain+probabilityDecrease {
		return _DECREASE
	} else if outcome < probabilitySuccess+probabilityMaintain+probabilityDecrease+probabilityBoom {
		return _BOOM
	}
	slog.Error("Case not caputured")
	return _SUCCESS
}

// performExperiment return (totalMesos, totalBooms, totalCount)
func performExperiment(newKms bool, currentStar, desiredStar, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen, boomEvent bool) (float64, int, int) {
	var (
		totalMesos                            float64
		totalBooms, totalCount, decreaseCount int
	)
	for currentStar < desiredStar {
		chanceTime := !newKms && decreaseCount == 2
		totalMesos += attemptCost(newKms, currentStar, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, chanceTime)
		totalCount++
		if chanceTime {
			decreaseCount = 0
			currentStar++
		} else {
			switch determineOutcome(newKms, currentStar, boomProtect, fiveTenFifteen, boomEvent) {
			case _SUCCESS:
				decreaseCount = 0
				currentStar++
			case _DECREASE:
				decreaseCount++
				currentStar--
			case _MAINTAIN:
				decreaseCount = 0
			case _BOOM:
				decreaseCount = 0
				currentStar = 12
				totalBooms++
			}
		}
	}
	return totalMesos, totalBooms, totalCount
}

func formatInt64(i int64) string {
	if i < 1000000000000 {
		return fmt.Sprintf("%.2fB", float64(i)/1000000000.0)
	}
	return fmt.Sprintf("%.2fT", float64(i)/1000000000000.0)
}

func calculateBoomCount(content string, newKms bool) MessageChain {
	boomProtect := strings.Contains(content, "保护") && !strings.Contains(content, "不保护")
	thirtyOff := strings.Contains(content, "七折") || strings.Contains(content, "超必")
	fiveTenFifteen := strings.Contains(content, "必成") || strings.Contains(content, "超必")
	boomEvent := strings.Contains(content, "超爆")
	title := "旧0-22星爆炸次数"
	if newKms {
		title = "新0-22星爆炸次数"
	}
	if thirtyOff && !fiveTenFifteen {
		title += "(七折)"
	}
	if fiveTenFifteen && !thirtyOff {
		title += "(必成)"
	}
	if thirtyOff && fiveTenFifteen {
		title += "(超必)"
	}
	if boomEvent {
		title += "(超爆)"
	}
	if boomProtect {
		title += "(保护)"
	}
	booms := make(map[int]int)
	for range 1000 {
		_, b, _ := performExperiment(newKms, 0, 22, 200, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
		booms[b]++
	}
	values := make([]float64, 0, len(booms))
	labels := make([]string, 0, len(booms))
	left := 1000
	for k := range 11 {
		labels = append(labels, fmt.Sprintf("%d次", k))
		values = append(values, float64(booms[k]))
		left -= booms[k]
	}
	if left > 0 {
		labels = append(labels, "超过10次")
		values = append(values, float64(left))
	}
	p, err := PieRender(
		values,
		TitleOptionFunc(TitleOption{
			Text: title,
			Left: PositionRight,
		}),
		LegendOptionFunc(LegendOption{
			Data: labels,
			Show: FalseFlag(),
		}),
		PieSeriesShowLabel(),
	)
	if err != nil {
		slog.Error("render chart failed", "error", err)
	} else if buf, err := p.Bytes(); err != nil {
		slog.Error("render chart failed", "error", err)
	} else {
		return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
	}
	return nil
}

var (
	regRexpStarForcePythonResult1 = regexp.MustCompile(`Cost Mean: <(.+?)>`)
	regRexpStarForcePythonResult2 = regexp.MustCompile(`Boom mean: <(.+?)>`)
	regRexpStarForcePythonResult3 = regexp.MustCompile(`Tap mean: <(.+?)>`)
	regRexpStarForcePythonResult4 = regexp.MustCompile(`Chance of no boom: <(.+?)%>`)
	regRexpStarForcePythonResult5 = regexp.MustCompile(`Midway costs: <(.*?)>`)
)

func parseFloat(s string) (float64, error) {
	multiplier := 1.0
	switch s[len(s)-1] {
	case 'r', 'R':
		multiplier = 1e18
	case 'q', 'Q':
		multiplier = 1e15
	case 't', 'T':
		multiplier = 1e12
	case 'b', 'B':
		multiplier = 1e9
	case 'm', 'M':
		multiplier = 1e6
	case 'k', 'K':
		multiplier = 1e3
	}
	if multiplier != 1.0 {
		s = s[:len(s)-1]
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v * multiplier, nil
}

func pythonStarForce(newKms bool, itemLevel, cur, des int, boomProtect, thirtyOff, fiveTenFifteen, boomEvent bool) (float64, float64, float64, float64, []float64, error) {
	result, err := scripts.RunPythonScript(
		"sf_cal_shell.py",
		strconv.Itoa(itemLevel),
		strconv.Itoa(cur),
		strconv.Itoa(des),
		strconv.FormatBool(boomProtect),
		"true",
		"false",
		strconv.FormatBool(newKms),
		strconv.FormatBool(thirtyOff),
		strconv.FormatBool(fiveTenFifteen),
		"false",
		strconv.FormatBool(boomEvent),
	)
	if err != nil {
		slog.Error("计算失败", "err", err, "result", string(result))
		return 0, 0, 0, 0, nil, errors.New("计算失败")
	}

	match1 := regRexpStarForcePythonResult1.FindSubmatch(result)
	if len(match1) == 0 {
		slog.Error("regexp star force python failed", "result", result)
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	mesos, err := parseFloat(string(match1[1]))
	if err != nil {
		slog.Error("parse mesos failed", "error", err, "value", string(match1[1]))
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	match2 := regRexpStarForcePythonResult2.FindSubmatch(result)
	if len(match2) == 0 {
		slog.Error("regexp star force python failed", "result", result)
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	booms, err := parseFloat(string(match2[1]))
	if err != nil {
		slog.Error("parse booms failed", "error", err, "value", string(match2[1]))
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	match3 := regRexpStarForcePythonResult3.FindSubmatch(result)
	if len(match3) == 0 {
		slog.Error("regexp star force python failed", "result", result)
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	count, err := parseFloat(string(match3[1]))
	if err != nil {
		slog.Error("parse count failed", "error", err, "value", string(match3[1]))
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	match4 := regRexpStarForcePythonResult4.FindSubmatch(result)
	if len(match4) == 0 {
		slog.Error("regexp star force python failed", "result", result)
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	noBoom, err := parseFloat(string(match4[1]))
	if err != nil {
		slog.Error("parse noBoom failed", "error", err, "value", string(match3[1]))
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	match5 := regRexpStarForcePythonResult5.FindSubmatch(result)
	if len(match5) == 0 {
		slog.Error("regexp star force python failed", "result", result)
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	var midway []float64
	err = json.Unmarshal(fmt.Appendf(nil, "[%s]", match5[1]), &midway)
	if err != nil {
		slog.Error("parse midway failed", "error", err, "value", string(match3[1]))
		return 0, 0, 0, 0, nil, errors.New("计算结果解析失败")
	}

	return mesos, booms, count, noBoom, midway, nil
}

func calculateStarForce1(newKms bool, content string) MessageChain {
	arr := strings.Split(content, " ")
	if len(arr) < 3 {
		return nil
	}
	itemLevel, err := strconv.Atoi(arr[0])
	if err != nil {
		return nil
	}
	if itemLevel < 5 || itemLevel > 300 {
		return MessageChain{&Text{Text: "装备等级不合理"}}
	}
	cur, err := strconv.Atoi(arr[1])
	if err != nil {
		return nil
	}
	if cur < 0 {
		return MessageChain{&Text{Text: "当前星数不合理"}}
	}
	des, err := strconv.Atoi(arr[2])
	if err != nil {
		return nil
	}
	if des <= cur {
		return MessageChain{&Text{Text: "目标星数必须大于当前星数"}}
	}
	maxStar := getMaxStar(newKms, itemLevel)
	if des > maxStar {
		return MessageChain{&Text{Text: fmt.Sprintf("%d级装备最多升到%d星", itemLevel, maxStar)}}
	}
	boomProtect := strings.Contains(content, "保护") && !strings.Contains(content, "不保护")
	thirtyOff := strings.Contains(content, "七折") || strings.Contains(content, "超必")
	fiveTenFifteen := strings.Contains(content, "必成") || strings.Contains(content, "超必")
	boomEvent := strings.Contains(content, "超爆")

	mesos, booms, count, noBoom, midway, err := pythonStarForce(newKms, itemLevel, cur, des, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
	if err != nil {
		return MessageChain{&Text{Text: err.Error()}}
	}

	data := []any{
		formatInt64(int64(mesos)),
		strconv.FormatFloat(booms, 'f', 2, 64),
		strconv.FormatInt(int64(math.Round(count)), 10),
		strconv.FormatFloat(noBoom, 'f', 2, 64),
	}
	var activity []string
	if thirtyOff {
		activity = append(activity, "七折活动")
	}
	if fiveTenFifteen {
		activity = append(activity, "5/10/15必成活动")
	}
	if boomEvent {
		activity = append(activity, "超爆活动")
	}
	var activityStr string
	if len(activity) > 0 {
		activityStr = "在" + strings.Join(activity, "和") + "中"
	}
	s := fmt.Sprintf("%s模拟升星%d级装备", activityStr, itemLevel)
	if boomProtect {
		s += "（点保护）"
	}
	if newKms {
		s += "（GMS新规）"
	} else {
		s += "（GMS旧规）"
	}
	s += fmt.Sprintf("\n%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次，有%s%%的概率一次都不炸", data...)
	if des > cur+1 && des > 12 {
		image := drawStarForce(cur, des, append(midway, mesos))
		if image != nil {
			return MessageChain{&Text{Text: s}, image}
		}
	}
	return MessageChain{&Text{Text: s}}
}

func calculateStarForce2(newKms bool, itemLevel int, thirtyOff, fiveTenFifteen, boomEvent bool) MessageChain {
	if itemLevel < 5 || itemLevel > 300 {
		return MessageChain{&Text{Text: "装备等级不合理"}}
	}
	maxStar := getMaxStar(newKms, itemLevel)
	var cur int
	if maxStar > 17 {
		cur = 17
	}
	var (
		des         = min(maxStar, 22)
		boomProtect = itemLevel >= 160
	)
	var data17 []any
	if maxStar > 17 {
		mesos, booms, count, noBoom, _, err := pythonStarForce(newKms, itemLevel, 0, 17, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
		if err != nil {
			return MessageChain{&Text{Text: err.Error()}}
		}
		data17 = []any{
			formatInt64(int64(mesos)),
			strconv.FormatFloat(booms, 'f', 2, 64),
			strconv.FormatInt(int64(math.Round(count)), 10),
			strconv.FormatFloat(noBoom, 'f', 2, 64),
		}
	}
	mesos, booms, count, noBoom, midway, err := pythonStarForce(newKms, itemLevel, cur, des, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
	if err != nil {
		return MessageChain{&Text{Text: err.Error()}}
	}
	data := []any{
		formatInt64(int64(mesos)),
		strconv.FormatFloat(booms, 'f', 2, 64),
		strconv.FormatInt(int64(math.Round(count)), 10),
		strconv.FormatFloat(noBoom, 'f', 2, 64),
	}
	var activity []string
	if thirtyOff {
		activity = append(activity, "七折活动")
	}
	if fiveTenFifteen {
		activity = append(activity, "5/10/15必成活动")
	}
	if boomEvent {
		activity = append(activity, "超爆活动")
	}
	var activityStr string
	if len(activity) > 0 {
		activityStr = "在" + strings.Join(activity, "和") + "中"
	}
	s := fmt.Sprintf("%s模拟升星%d级装备", activityStr, itemLevel)
	if boomProtect {
		s += "（点保护）"
	}
	if newKms {
		s += "（GMS新规）"
	} else {
		s += "（GMS旧规）"
	}
	if maxStar > 17 {
		s += fmt.Sprintf("\n0-17星，平均花费了%s金币，平均炸了%s次，平均点了%s次，有%s%%的概率一次都不炸", data17...)
	}
	s += fmt.Sprintf("\n%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次，有%s%%的概率一次都不炸", data...)
	image := drawStarForce(cur, des, append(midway, mesos))
	if image != nil {
		return MessageChain{&Text{Text: s}, image}
	}
	return MessageChain{&Text{Text: s}}
}

func drawStarForce(cur, des int, midway []float64) *Image {
	var labels []string
	var values []float64
	add := func(cur1, des1 int) {
		labels = append(labels, fmt.Sprintf("%d-%d", cur1, des1))
		index1 := des1 - cur - 1
		mesos := midway[index1]
		index0 := cur1 - cur - 1
		if index0 >= 0 {
			mesos -= midway[index0]
		}
		values = append(values, mesos)
	}
	if des <= 17 {
		if cur < 12 {
			add(cur, 12)
		}
		for i := max(cur, 12); i < des; i++ {
			add(i, i+1)
		}
	} else {
		if cur < 15 {
			add(cur, 15)
		}
		for i := max(cur, 15); i < des; i++ {
			add(i, i+1)
		}
	}
	if len(values) <= 1 {
		return nil
	}
	maxValue := slices.Max(values)
	var (
		exp     = "T"
		divisor = 1000000000000.0
	)
	if maxValue < 1000000000000 {
		exp = "B"
		divisor = 1000000000
	}
	for i := range values {
		values[i] /= divisor
	}
	p, err := PieRender(
		values,
		PaddingOptionFunc(Box{Top: 50}),
		LegendOptionFunc(LegendOption{
			Show: FalseFlag(),
			Data: labels,
		}),
		func(opt *ChartOption) {
			for index := range opt.SeriesList {
				opt.SeriesList[index].Label.Show = true
				opt.SeriesList[index].Label.Formatter = "{b}: {c}" + exp
			}
		},
	)
	if err != nil {
		slog.Error("render chart failed", "error", err)
		return nil
	} else if buf, err := p.Bytes(); err != nil {
		slog.Error("render chart failed", "error", err)
		return nil
	} else {
		return &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}
	}
}

type starForceResult int

const (
	_SUCCESS starForceResult = iota
	_MAINTAIN
	_DECREASE
	_BOOM
)

// current_star => (success, maintain, decrease, boom)
var rates = [][]float64{
	{0.95, 0.05, 0.0, 0.0},
	{0.9, 0.1, 0.0, 0.0},
	{0.85, 0.15, 0.0, 0.0},
	{0.85, 0.15, 0.0, 0.0},
	{0.80, 0.2, 0.0, 0.0},
	{0.75, 0.25, 0.0, 0.0},
	{0.7, 0.3, 0.0, 0.0},
	{0.65, 0.35, 0.0, 0.0},
	{0.6, 0.4, 0.0, 0.0},
	{0.55, 0.45, 0.0, 0.0},
	{0.5, 0.5, 0.0, 0.0},
	{0.45, 0.55, 0.0, 0.0},
	{0.4, 0.6, 0.0, 0.0},
	{0.35, 0.65, 0.0, 0.0},
	{0.3, 0.7, 0.0, 0.0},
	{0.3, 0.679, 0.0, 0.021},
	{0.3, 0.0, 0.679, 0.021},
	{0.3, 0.0, 0.679, 0.021},
	{0.3, 0.0, 0.672, 0.028},
	{0.3, 0.0, 0.672, 0.028},
	{0.3, 0.63, 0.0, 0.07},
	{0.3, 0.0, 0.63, 0.07},
	{0.03, 0.0, 0.776, 0.194},
	{0.02, 0.0, 0.686, 0.294},
	{0.01, 0.0, 0.594, 0.396},
}

// current_star => (success, maintain, decrease, boom)
var rates2 = [][]float64{
	{0.95, 0.05, 0.0, 0.0},
	{0.9, 0.1, 0.0, 0.0},
	{0.85, 0.15, 0.0, 0.0},
	{0.85, 0.15, 0.0, 0.0},
	{0.80, 0.2, 0.0, 0.0},
	{0.75, 0.25, 0.0, 0.0},
	{0.7, 0.3, 0.0, 0.0},
	{0.65, 0.35, 0.0, 0.0},
	{0.6, 0.4, 0.0, 0.0},
	{0.55, 0.45, 0.0, 0.0},
	{0.5, 0.5, 0.0, 0.0},
	{0.45, 0.55, 0.0, 0.0},
	{0.4, 0.6, 0.0, 0.0},
	{0.35, 0.65, 0.0, 0.0},
	{0.3, 0.7, 0.0, 0.0},
	{0.3, 0.679, 0.0, 0.021},
	{0.3, 0.679, 0.0, 0.021},
	{0.15, 0.782, 0.0, 0.068},
	{0.15, 0.782, 0.0, 0.068},
	{0.15, 0.765, 0.0, 0.085},
	{0.3, 0.595, 0.0, 0.105},
	{0.15, 0.7225, 0.0, 0.1275},
	{0.15, 0.68, 0.0, 0.17},
	{0.1, 0.72, 0.0, 0.18},
	{0.1, 0.72, 0.0, 0.18},
	{0.1, 0.72, 0.0, 0.18},
	{0.07, 0.744, 0.0, 0.186},
	{0.05, 0.76, 0.0, 0.19},
	{0.03, 0.776, 0.0, 0.194},
	{0.01, 0.792, 0.0, 0.198},
}

func getMaxStar(newKms bool, itemLevel int) int {
	switch {
	case itemLevel < 95:
		return 5
	case itemLevel < 108:
		return 8
	case itemLevel < 118:
		return 10
	case itemLevel < 128:
		return 15
	case itemLevel < 138:
		return 20
	default:
		if newKms {
			return 30
		}
		return 25
	}
}
