// 星之力计算器的工具类，相关数据取自:
//
// https://strategywiki.org/wiki/MapleStory/Spell_Trace_and_Star_Force#Meso_Cost

package maplebot

import (
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"log/slog"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
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

func saviorMesoFn(currentStar int) func(int, int) float64 {
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

func saviorCost(currentStar, itemLevel int) float64 {
	return saviorMesoFn(currentStar)(currentStar, itemLevel)
}

func attemptCost(currentStar int, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen, chanceTime bool) float64 {
	multiplier := 1.0
	if boomProtect && !(fiveTenFifteen && currentStar == 15) && !chanceTime && (currentStar == 15 || currentStar == 16) {
		multiplier += 1.0
	}
	if thirtyOff {
		multiplier -= 0.3
	}
	cost := saviorCost(currentStar, itemLevel) * multiplier
	return math.Round(cost)
}

// determineOutcome return either _SUCCESS, _MAINTAIN, _DECREASE, or _BOOM
func determineOutcome(currentStar int, boomProtect, fiveTenFifteen bool) starForceResult {
	if fiveTenFifteen && (currentStar == 5 || currentStar == 10 || currentStar == 15) {
		return _SUCCESS
	}
	var (
		outcome             = rand.Float64()
		probabilitySuccess  = rates[currentStar][_SUCCESS]
		probabilityMaintain = rates[currentStar][_MAINTAIN]
		probabilityDecrease = rates[currentStar][_DECREASE]
		probabilityBoom     = rates[currentStar][_BOOM]
	)
	if boomProtect && currentStar <= 16 { // boom protection enabled
		probabilityDecrease += probabilityBoom
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
func performExperiment(currentStar, desiredStar, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen bool) (float64, int, int) {
	var (
		totalMesos                            float64
		totalBooms, totalCount, decreaseCount int
	)
	for currentStar < desiredStar {
		chanceTime := decreaseCount == 2
		totalMesos += attemptCost(currentStar, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, chanceTime)
		totalCount++
		if chanceTime {
			decreaseCount = 0
			currentStar++
		} else {
			switch determineOutcome(currentStar, boomProtect, fiveTenFifteen) {
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

func calculateBoomCount(content string) MessageChain {
	boomProtect := strings.Contains(content, "保护")
	thirtyOff := strings.Contains(content, "七折") || strings.Contains(content, "超必")
	fiveTenFifteen := strings.Contains(content, "必成") || strings.Contains(content, "超必")
	title := "0-22星爆炸次数"
	if thirtyOff && !fiveTenFifteen {
		title += "(七折)"
	}
	if fiveTenFifteen && !thirtyOff {
		title += "(必成)"
	}
	if thirtyOff && fiveTenFifteen {
		title += "(超必)"
	}
	if boomProtect {
		title += "(保护)"
	}
	booms := make(map[int]int)
	for range 1000 {
		_, b, _ := performExperiment(0, 22, 200, boomProtect, thirtyOff, fiveTenFifteen)
		booms[b]++
	}
	values := make([]float64, 0, len(booms))
	labels := make([]string, 0, len(booms))
	left := 1000
	for k := 0; k <= 10; k++ {
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

func calculateStarForce1(content string) MessageChain {
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
	maxStar := getMaxStar(itemLevel)
	if des > maxStar {
		return MessageChain{&Text{Text: fmt.Sprintf("%d级装备最多升到%d星", itemLevel, maxStar)}}
	}
	if des > 24 {
		return MessageChain{&Text{Text: fmt.Sprintf("还想升%d星？梦里什么都有", des)}}
	}
	boomProtect := strings.Contains(content, "保护")
	thirtyOff := strings.Contains(content, "七折") || strings.Contains(content, "超必")
	fiveTenFifteen := strings.Contains(content, "必成") || strings.Contains(content, "超必")
	var (
		mesos        float64
		booms, count int
	)
	testCount := 1000
	if des >= 24 {
		testCount = 20
	}
	for range testCount {
		m, b, c := performExperiment(cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen)
		mesos += m
		booms += b
		count += c
	}
	data := []any{
		formatInt64(int64(math.Round(mesos / float64(testCount)))),
		strconv.FormatFloat(float64(booms)/float64(testCount), 'f', -1, 64),
		strconv.FormatInt(int64(math.Round(float64(count)/float64(testCount))), 10),
	}
	var activity []string
	if thirtyOff {
		activity = append(activity, "七折活动")
	}
	if fiveTenFifteen {
		activity = append(activity, "5/10/15必成活动")
	}
	var activityStr string
	if len(activity) > 0 {
		activityStr = "在" + strings.Join(activity, "和") + "中"
	}
	s := fmt.Sprintf("%s模拟升星%d级装备", activityStr, itemLevel)
	if boomProtect {
		s += "（点保护）"
	}
	s += fmt.Sprintf("\n共测试了%d次\n", testCount)
	s += fmt.Sprintf("%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次", data...)
	image := drawStarForce(cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, mesos/float64(testCount), testCount)
	if image != nil {
		return MessageChain{&Text{Text: s}, image}
	}
	return MessageChain{&Text{Text: s}}
}

func calculateStarForce2(itemLevel int, thirtyOff, fiveTenFifteen bool) MessageChain {
	if itemLevel < 5 || itemLevel > 300 {
		return MessageChain{&Text{Text: "装备等级不合理"}}
	}
	maxStar := getMaxStar(itemLevel)
	var cur int
	if maxStar > 17 {
		cur = 17
	}
	var (
		des                                = min(maxStar, 22)
		boomProtect                        = itemLevel >= 160
		mesos17, mesos22                   float64
		booms17, count17, booms22, count22 int
	)
	for range 1000 {
		if maxStar > 17 {
			m, b, c := performExperiment(0, 17, itemLevel, boomProtect, thirtyOff, fiveTenFifteen)
			mesos17 += m
			booms17 += b
			count17 += c
		}
		m, b, c := performExperiment(cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen)
		mesos22 += m
		booms22 += b
		count22 += c
	}
	var data17 []any
	if maxStar > 17 {
		data17 = []any{
			formatInt64(int64(math.Round(mesos17 / 1000))),
			strconv.FormatFloat(float64(booms17)/1000.0, 'f', -1, 64),
			strconv.FormatInt(int64(math.Round(float64(count17)/1000.0)), 10),
		}
	}
	data := []any{
		formatInt64(int64(math.Round(mesos22 / 1000))),
		strconv.FormatFloat(float64(booms22)/1000.0, 'f', -1, 64),
		strconv.FormatInt(int64(math.Round(float64(count22)/1000.0)), 10),
	}
	var activity []string
	if thirtyOff {
		activity = append(activity, "七折活动")
	}
	if fiveTenFifteen {
		activity = append(activity, "5/10/15必成活动")
	}
	var activityStr string
	if len(activity) > 0 {
		activityStr = "在" + strings.Join(activity, "和") + "中"
	}
	s := fmt.Sprintf("%s模拟升星%d级装备", activityStr, itemLevel)
	if boomProtect {
		s += "（点保护）"
	}
	s += "\n共测试了1000次\n"
	if maxStar > 17 {
		s += fmt.Sprintf("0-17星，平均花费了%s金币，平均炸了%s次，平均点了%s次\n", data17...)
	}
	s += fmt.Sprintf("%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次", data...)
	image := drawStarForce(0, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, mesos22/1000, 1000)
	if image != nil {
		return MessageChain{&Text{Text: s}, image}
	}
	return MessageChain{&Text{Text: s}}
}

func drawStarForce(cur, des, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen bool, totalMesos float64, testCount int) *Image {
	var labels []string
	var values []float64
	add := func(cur1, des1 int) {
		if des1 == des {
			var value float64
			for _, v := range values {
				value += v
			}
			value = totalMesos - value
			if value > 0 {
				labels = append(labels, fmt.Sprintf("%d-%d", cur1, des1))
				values = append(values, value)
			}
			return
		}
		labels = append(labels, fmt.Sprintf("%d-%d", cur1, des1))
		var a float64
		for range testCount {
			m, _, _ := performExperiment(cur1, des1, itemLevel, boomProtect, thirtyOff, fiveTenFifteen)
			a += m
		}
		values = append(values, a/float64(testCount))
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

func getMaxStar(itemLevel int) int {
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
		return 25
	}
}
