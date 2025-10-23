// 星之力计算器的工具类，相关数据取自:
//
// https://strategywiki.org/wiki/MapleStory/Spell_Trace_and_Star_Force#Meso_Cost

package maplebot

import (
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"gonum.org/v1/gonum/mat"
	"log/slog"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
)

func calStarForceCostPerformance() MessageChain {
	experiment := func(cur, des int, thirtyOff, fiveTenFifteen, boomEvent bool) string {
		var mesos float64
		for range 1000 {
			m, _, _ := performExperiment(false, cur, des, 150, false, thirtyOff, fiveTenFifteen, boomEvent)
			mesos += m
		}
		v := 11 + 4*(cur-6)
		if cur == 21 {
			v += 4
		}
		return fmt.Sprintf("%.2fB", mesos/1000/float64(v)*(9*3)/1000000000)
	}
	values := make([][]string, 0, 7)
	for i := 15; i <= 21; i++ {
		v := make([]string, 0, 5)
		v = append(v, fmt.Sprintf("%d★ → %d★", i, i+1))
		v = append(v, experiment(i, i+1, false, false, false))
		v = append(v, experiment(i, i+1, true, false, false))
		v = append(v, experiment(i, i+1, false, true, false))
		v = append(v, experiment(i, i+1, true, true, false))
		v = append(v, experiment(i, i+1, false, false, true))
		values = append(values, v)
	}
	p, err := TableOptionRender(TableChartOption{
		Header: []string{"150装备升星", "无活动", "七折", "必成", "超必", "超爆"},
		Data:   values,
		Width:  700,
		CellStyle: func(cell TableCell) *Style {
			if cell.Column == 0 {
				return &Style{FillColor: drawing.Color{R: 180, G: 180, B: 180, A: 128}}
			}
			return nil
		},
	})
	if err != nil {
		slog.Error("render chart failed", "error", err)
	} else if buf, err := p.Bytes(); err != nil {
		slog.Error("render chart failed", "error", err)
	} else {
		return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
	}
	return nil
}

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

func getProbable(newKms bool, currentStar int, boomProtect, fiveTenFifteen, boomEvent bool) (float64, float64, float64, float64) {
	if fiveTenFifteen && (currentStar == 5 || currentStar == 10 || currentStar == 15) {
		return 1, 0, 0, 0
	}
	rates := rates
	if newKms {
		rates = rates2
	}
	var (
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
	return probabilitySuccess, probabilityMaintain, probabilityDecrease, probabilityBoom
}

// determineOutcome return either _SUCCESS, _MAINTAIN, _DECREASE, or _BOOM
func determineOutcome(newKms bool, currentStar int, boomProtect, fiveTenFifteen, boomEvent bool) starForceResult {
	if fiveTenFifteen && (currentStar == 5 || currentStar == 10 || currentStar == 15) {
		return _SUCCESS
	}
	outcome := rand.Float64()
	probabilitySuccess, probabilityMaintain, probabilityDecrease, probabilityBoom := getProbable(newKms, currentStar, boomProtect, fiveTenFifteen, boomEvent)
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

func perform20To22(itemLevel int, thirtyOff bool) (starForceResult, float64) {
	var (
		currentStar = 20
		totalMesos  float64
	)
	for currentStar < 22 {
		totalMesos += attemptCost(false, currentStar, itemLevel, false, thirtyOff, false, false)
		switch determineOutcome(false, currentStar, false, false, false) {
		case _SUCCESS:
			currentStar++
		case _DECREASE:
			currentStar--
		case _MAINTAIN:
		case _BOOM:
			return _BOOM, totalMesos
		}
	}
	return _SUCCESS, totalMesos
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

func calculateStarForce(newKms bool, currentStar, desiredStar, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen, boomEvent bool) (mesos, boom, count float64) {
	maxStates := 3 * (desiredStar + 1)
	fsm := mat.NewDense(maxStates, maxStates, nil)
	ev := mat.NewVecDense(maxStates, nil)
	for i := range desiredStar {
		success, maintain, decrease, boom := getProbable(newKms, currentStar, boomProtect, fiveTenFifteen, boomEvent)
		fsm.Set(i*3, i*3, 1-maintain)
		fsm.Set(i*3, (i+1)*3, -success)
		if i > 0 {
			fsm.Set(i*3, (i-1)*3+1, -decrease)
		}
		if desiredStar >= 12 {
			fsm.Set(i*3, 12*3, -boom)
		}
		ev.SetVec(i*3, 1)
		if i <= 15 || i == 19 || i > desiredStar-2 {
			fsm.Set(i*3+1, i*3+1, 1)
		} else {
			fsm.Set(i*3+1, i*3, 1-maintain)
			fsm.Set(i*3+1, (i+1)*3, -success)
			fsm.Set(i*3+1, (i-1)*3+2, -decrease)
			if desiredStar >= 12 {
				fsm.Set(i*3, 12*3, -boom)
			}
			ev.SetVec(i*3+1, 1)
		}
		if i < 15 || i == 18 || i == 19 || i >= desiredStar-2 {
			fsm.Set(i*3+2, i*3+2, 1)
		} else {
			fsm.Set(i*3+2, i*3, 1)
			fsm.Set(i*3+2, (i+1)*3, -1)
			ev.SetVec(i*3+2, 1)
		}
	}
	fsm.Set(desiredStar*3, desiredStar*3, 1)
	fsm.Set(desiredStar*3+1, desiredStar*3+1, 1)
	fsm.Set(desiredStar*3+2, desiredStar*3+2, 1)
	for i1 := range maxStates {
		for i2 := i1 + 1; i2 < maxStates; i2++ {
			if func() bool {
				for j := range maxStates {
					if fsm.At(i1, j) != fsm.At(i2, j) {
						return false
					}
				}
				return true
			}() {
				fmt.Printf("发现重复行: %d 和 %d\n", i1, i2)
				for j := range maxStates {
					fmt.Printf("%+.4f ", fsm.At(i1, j))
				}
				fmt.Printf("%+.4f\n", ev.AtVec(i1))
				for j := range maxStates {
					fmt.Printf("%+.4f ", fsm.At(i2, j))
				}
				fmt.Printf("%+.4f\n", ev.AtVec(i2))
			}
		}
	}
	var X mat.VecDense
	err := X.SolveVec(fsm, ev)
	if err != nil {
		fmt.Printf("求解失败: %v\n", err)
		return
	}
	count = X.AtVec(currentStar * 3)
	return
}

func calculate20To22() MessageChain {
	const testCount = 10000
	var (
		totalBooms int
		ss         = make([][]string, 0, 5)
	)
	for _, level := range []int{140, 150, 160, 200, 250} {
		var mesos, mesosThirtyOff float64
		for range testCount {
			b, m := perform20To22(level, false)
			mesos += m
			if b == _BOOM {
				totalBooms++
			}
			b, m = perform20To22(level, true)
			mesosThirtyOff += m
			if b == _BOOM {
				totalBooms++
			}
		}
		ss = append(ss, []string{
			fmt.Sprintf("%d级", level),
			formatInt64(int64(math.Round(mesos / float64(testCount)))),
			formatInt64(int64(math.Round(mesosThirtyOff / float64(testCount)))),
		})
	}
	ret := MessageChain{&Text{Text: fmt.Sprintf("20★ → 22★，有%.2f%%的几率中途爆炸，平均消耗%.1f个备件能成功", float64(totalBooms)/testCount*10, testCount*10/float64(totalBooms))}}
	p, err := TableOptionRender(TableChartOption{
		Width:  300,
		Header: []string{"平均花费", "无活动", "七折"},
		Data:   ss,
	})
	if err != nil {
		slog.Error("render chart failed", "error", err)
	} else if buf, err := p.Bytes(); err != nil {
		slog.Error("render chart failed", "error", err)
	} else {
		ret = append(ret, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)})
	}
	return ret
}

func calculateBoomCount(content string, newKms bool) MessageChain {
	boomProtect := strings.Contains(content, "保护")
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

func calculateStarMarkov(newKms bool, content string) MessageChain {
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
	boomEvent := strings.Contains(content, "超爆")
	var (
		mesos        float64
		booms, count int
	)
	testCount := 1000
	if des >= 24 {
		testCount = 20
	}
	for range testCount {
		m, b, c := performExperiment(newKms, cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
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
		s += "（KMS新规则）"
	}
	s += fmt.Sprintf("\n共测试了%d次\n", testCount)
	s += fmt.Sprintf("%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次", data...)
	image := drawStarForce(newKms, cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent, mesos/float64(testCount), testCount)
	if image != nil {
		return MessageChain{&Text{Text: s}, image}
	}
	return MessageChain{&Text{Text: s}}
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
	boomEvent := strings.Contains(content, "超爆")
	var (
		mesos        float64
		booms, count int
	)
	testCount := 1000
	if des >= 24 {
		testCount = 20
	}
	for range testCount {
		m, b, c := performExperiment(newKms, cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
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
		s += "（KMS新规则）"
	}
	s += fmt.Sprintf("\n共测试了%d次\n", testCount)
	s += fmt.Sprintf("%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次", data...)
	image := drawStarForce(newKms, cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent, mesos/float64(testCount), testCount)
	if image != nil {
		return MessageChain{&Text{Text: s}, image}
	}
	return MessageChain{&Text{Text: s}}
}

func calculateStarForce2(newKms bool, itemLevel int, thirtyOff, fiveTenFifteen, boomEvent bool) MessageChain {
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
			m, b, c := performExperiment(newKms, 0, 17, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
			mesos17 += m
			booms17 += b
			count17 += c
		}
		m, b, c := performExperiment(newKms, cur, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
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
		s += "（KMS新规则）"
	}
	s += "\n共测试了1000次\n"
	if maxStar > 17 {
		s += fmt.Sprintf("0-17星，平均花费了%s金币，平均炸了%s次，平均点了%s次\n", data17...)
	}
	s += fmt.Sprintf("%d-%d星", cur, des)
	s += fmt.Sprintf("，平均花费了%s金币，平均炸了%s次，平均点了%s次", data...)
	image := drawStarForce(newKms, 0, des, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent, mesos22/1000, 1000)
	if image != nil {
		return MessageChain{&Text{Text: s}, image}
	}
	return MessageChain{&Text{Text: s}}
}

func drawStarForce(newKms bool, cur, des, itemLevel int, boomProtect, thirtyOff, fiveTenFifteen, boomEvent bool, totalMesos float64, testCount int) *Image {
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
			m, _, _ := performExperiment(newKms, cur1, des1, itemLevel, boomProtect, thirtyOff, fiveTenFifteen, boomEvent)
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
