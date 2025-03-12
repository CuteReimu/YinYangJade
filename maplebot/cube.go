package maplebot

import (
	"cmp"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
)

type nameAndLevel struct {
	name  string
	level int
}

var nameMap = map[string]nameAndLevel{
	"戒指": {name: "accessory", level: 150},
	"项链": {name: "accessory", level: 150},
	"耳环": {name: "accessory", level: 150},
	"首饰": {name: "accessory", level: 150},
	"腰带": {name: "belt", level: 150},
	"副手": {name: "secondary", level: 140},
	"上衣": {name: "top", level: 150},
	"下衣": {name: "bottom", level: 150},
	"披风": {name: "cape", level: 200},
	"纹章": {name: "emblem", level: 100},
	"手套": {name: "gloves", level: 200},
	"帽子": {name: "hat", level: 150},
	"心脏": {name: "heart", level: 100},
	"套服": {name: "overall", level: 200},
	"鞋子": {name: "shoes", level: 200},
	"护肩": {name: "shoulder", level: 200},
	"武器": {name: "weapon", level: 200},
}

var statMap = map[string]string{
	"percStat":           "%s%%+属性",
	"lineStat":           "%s条属性",
	"percAtt":            "%s%%+攻",
	"lineAtt":            "%s条攻",
	"percBoss":           "%s%%+BD",
	"lineBoss":           "%s条BD",
	"lineIed":            "%s条无视",
	"lineCritDamage":     "%s条爆伤",
	"lineMeso":           "%s条钱",
	"lineDrop":           "%s条爆",
	"lineMesoOrDrop":     "%s条钱爆",
	"secCooldown":        "%s秒CD",
	"lineAttOrBoss":      "%s条攻或BD",
	"lineAttOrBossOrIed": "总计%s条有用属性",
	"lineBossOrIed":      "%s条BD或无视",
}

var (
	defaultSelections    = slices.Clip([]string{"percStat+18", "percStat+21", "percStat+24", "percStat+27", "percStat+30", "percStat+33", "percStat+36"})
	defaultSelections160 = append(defaultSelections, "percStat+39")
	accessorySelections  = append(defaultSelections,
		"lineMeso+1", "lineDrop+1", "lineMesoOrDrop+1",
		"lineMeso+2", "lineDrop+2", "lineMesoOrDrop+2",
		"lineMeso+3", "lineDrop+3", "lineMeso+1&lineStat+1", "lineDrop+1&lineStat+1", "lineMesoOrDrop+1&lineStat+1",
	)
	accessorySelections160 = append(defaultSelections160,
		"lineMeso+1", "lineDrop+1", "lineMesoOrDrop+1",
		"lineMeso+2", "lineDrop+2", "lineMesoOrDrop+2",
		"lineMeso+3", "lineDrop+3", "lineMeso+1&lineStat+1", "lineDrop+1&lineStat+1", "lineMesoOrDrop+1&lineStat+1",
	)
	hatSelections = append(defaultSelections,
		"secCooldown+2", "secCooldown+3", "secCooldown+4", "secCooldown+5", "secCooldown+6",
		"secCooldown+2&lineStat+2",
		"secCooldown+2&lineStat+1", "secCooldown+3&lineStat+1", "secCooldown+4&lineStat+1",
	)
	hatSelections160 = append(defaultSelections160,
		"secCooldown+2", "secCooldown+3", "secCooldown+4", "secCooldown+5", "secCooldown+6",
		"secCooldown+2&lineStat+2",
		"secCooldown+2&lineStat+1", "secCooldown+3&lineStat+1", "secCooldown+4&lineStat+1",
	)
	gloveSelections = append(defaultSelections,
		"lineCritDamage+1", "lineCritDamage+2", "lineCritDamage+3",
		"lineCritDamage+1&lineStat+1", "lineCritDamage+1&lineStat+2", "lineCritDamage+2&lineStat+1",
	)
	gloveSelections160 = append(defaultSelections160,
		"lineCritDamage+1", "lineCritDamage+2", "lineCritDamage+3",
		"lineCritDamage+1&lineStat+1", "lineCritDamage+1&lineStat+2", "lineCritDamage+2&lineStat+1",
	)
	wsSelections = []string{
		"percAtt+18", "percAtt+21", "percAtt+24", "percAtt+30", "percAtt+33", "percAtt+36",
		"lineIed+1&percAtt+18", "lineIed+1&percAtt+21", "lineIed+1&percAtt+24",
		"lineAttOrBossOrIed+1", "lineAttOrBossOrIed+2", "lineAttOrBossOrIed+3",
		"lineAtt+1&lineAttOrBossOrIed+2", "lineAtt+1&lineAttOrBossOrIed+3", "lineAtt+2&lineAttOrBossOrIed+3",
		"lineAtt+1&lineBoss+1", "lineAtt+1&lineBoss+2", "lineAtt+2&lineBoss+1",
		"percAtt+21&percBoss+30", "percAtt+21&percBoss+35", "percAtt+21&percBoss+40",
		"percAtt+24&percBoss+30",
		"lineAttOrBoss+1", "lineAttOrBoss+2", "lineAttOrBoss+3",
	}
	wsSelections160 = []string{
		"percAtt+20", "percAtt+23", "percAtt+26", "percAtt+33", "percAtt+36", "percAtt+39",
		"lineIed+1&percAtt+20", "lineIed+1&percAtt+23", "lineIed+1&percAtt+26",
		"lineAttOrBossOrIed+1", "lineAttOrBossOrIed+2", "lineAttOrBossOrIed+3",
		"lineAtt+1&lineAttOrBossOrIed+2", "lineAtt+1&lineAttOrBossOrIed+3", "lineAtt+2&lineAttOrBossOrIed+3",
		"lineAtt+1&lineBoss+1", "lineAtt+1&lineBoss+2", "lineAtt+2&lineBoss+1",
		"percAtt+23&percBoss+30", "percAtt+23&percBoss+35", "percAtt+23&percBoss+40",
		"percAtt+26&percBoss+30",
		"lineAttOrBoss+1", "lineAttOrBoss+2", "lineAttOrBoss+3",
	}
	eSelections = []string{
		"percAtt+18", "percAtt+21", "percAtt+24", "percAtt+30", "percAtt+33", "percAtt+36",
		"lineIed+1&percAtt+18", "lineIed+1&percAtt+21", "lineIed+1&percAtt+24",
		"lineAttOrBossOrIed+1", "lineAttOrBossOrIed+2", "lineAttOrBossOrIed+3",
		"lineAtt+1&lineAttOrBossOrIed+2", "lineAtt+1&lineAttOrBossOrIed+3", "lineAtt+2&lineAttOrBossOrIed+3",
	}
	eSelections160 = []string{
		"percAtt+20", "percAtt+23", "percAtt+26", "percAtt+33", "percAtt+36", "percAtt+39",
		"lineIed+1&percAtt+20", "lineIed+1&percAtt+23", "lineIed+1&percAtt+26",
		"lineAttOrBossOrIed+1", "lineAttOrBossOrIed+2", "lineAttOrBossOrIed+3",
		"lineAtt+1&lineAttOrBossOrIed+2", "lineAtt+1&lineAttOrBossOrIed+3", "lineAtt+2&lineAttOrBossOrIed+3",
	}
)

func getSelection(name string, itemLevel int) []string {
	switch name {
	case "emblem":
		if itemLevel < 160 {
			return eSelections
		} else {
			return eSelections160
		}
	case "weapon", "secondary":
		if itemLevel < 160 {
			return wsSelections
		} else {
			return wsSelections160
		}
	case "accessory":
		if itemLevel < 160 {
			return accessorySelections
		} else {
			return accessorySelections160
		}
	case "hat":
		if itemLevel < 160 {
			return hatSelections
		} else {
			return hatSelections160
		}
	case "gloves":
		if itemLevel < 160 {
			return gloveSelections
		} else {
			return gloveSelections160
		}
	default:
		if itemLevel < 160 {
			return defaultSelections
		} else {
			return defaultSelections160
		}
	}
}

func calculateCubeAll() MessageChain {
	names := make([]string, 0, 10)
	for s, nameLevel := range nameMap {
		if slices.Contains([]string{"weapon", "secondary", "emblem", "overall"}, nameLevel.name) {
			continue
		}
		if slices.Contains([]string{"戒指", "项链", "耳环"}, s) {
			continue
		}
		names = append(names, s)
	}
	selections := defaultSelections160
	ss := make([][]string, len(names))
	styles := make(map[string][]*Style, len(names))
	costs := make(map[string]int64, len(names))
	for i, s := range names {
		nameLevel := nameMap[s]
		ss[i] = append(ss[i], fmt.Sprintf("%d%s", nameLevel.level, s))
		for _, it := range selections {
			if nameLevel.level < 160 && it == "percStat+39" {
				ss[i] = append(ss[i], "")
				continue
			}
			_, red := runCalculator(nameLevel.name, "red", 3, nameLevel.level, 3, it)
			_, black := runCalculator(nameLevel.name, "black", 3, nameLevel.level, 3, it)
			var (
				cost int64
			)
			if red <= black {
				styles[ss[i][0]] = append(styles[ss[i][0]], &Style{FillColor: drawing.Color{R: 0, G: 132, B: 199, A: 96}})
				cost = red
			} else {
				styles[ss[i][0]] = append(styles[ss[i][0]], &Style{FillColor: drawing.Color{R: 147, G: 21, B: 152, A: 96}})
				cost = black
			}
			ss[i] = append(ss[i], formatInt64(cost))
			if it == "percStat+27" {
				costs[s] = cost
			}
		}
	}
	slices.SortFunc(ss, func(a, b []string) int {
		aName, bName := a[0][3:], b[0][3:]
		if nameMap[aName].level != nameMap[bName].level {
			return nameMap[aName].level - nameMap[bName].level
		}
		return cmp.Compare(costs[aName], costs[bName])
	})
	header := make([]string, 0, len(selections)+1)
	header = append(header, "部位")
	for _, it := range selections {
		var target string
		for _, stat := range strings.Split(it, "&") {
			arr := strings.Split(stat, "+")
			target += fmt.Sprintf(statMap[arr[0]], arr[1])
		}
		header = append(header, target)
	}
	p, err := TableOptionRender(TableChartOption{
		Width:           770,
		Header:          header,
		Data:            ss,
		HeaderFontColor: Color{R: 35, G: 35, B: 35, A: 255},
		CellStyle: func(cell TableCell) *Style {
			if cell.Column > 0 && cell.Row > 0 && len(cell.Text) > 0 {
				return styles[ss[cell.Row-1][0]][cell.Column-1]
			}
			return nil
		},
	})
	if err != nil {
		slog.Error("生成表格失败", "error", err)
		return nil
	}
	buf, err := p.Bytes()
	if err != nil {
		slog.Error("生成表格失败", "error", err)
		return nil
	}
	return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
}

func calculateCube(s string) MessageChain {
	sAndLevel := strings.Split(s, " ")
	if len(sAndLevel) >= 2 {
		s = sAndLevel[0]
	}
	nameLevel, ok := nameMap[s]
	if !ok {
		return nil
	}
	if len(sAndLevel) >= 2 {
		level, err := strconv.Atoi(sAndLevel[1])
		if err != nil {
			return nil
		}
		if level < 71 {
			return MessageChain{&Text{Text: "不能低于71级"}}
		} else if level > 300 {
			return MessageChain{&Text{Text: "不能高于300级"}}
		}
		nameLevel.level = level
	}
	s = strings.TrimLeft(s, "0123456789")
	selections := getSelection(nameLevel.name, nameLevel.level)
	styles := make([]*Style, 0, len(selections)+1)
	_, eToLR := runCalculator(nameLevel.name, "red", 1, nameLevel.level, 3, "")
	_, eToLB := runCalculator(nameLevel.name, "black", 1, nameLevel.level, 3, "")
	var eToLCost int64
	if eToLR < eToLB {
		styles = append(styles, &Style{FillColor: drawing.Color{R: 0, G: 132, B: 199, A: 96}})
		eToLCost = eToLR
	} else {
		styles = append(styles, &Style{FillColor: drawing.Color{R: 147, G: 21, B: 152, A: 96}})
		eToLCost = eToLB
	}
	ss := make([][]string, 0, len(selections)+1)
	ss = append(ss, []string{"紫洗绿", formatInt64(eToLCost)})
	for _, it := range selections {
		_, red := runCalculator(nameLevel.name, "red", 3, nameLevel.level, 3, it)
		_, black := runCalculator(nameLevel.name, "black", 3, nameLevel.level, 3, it)
		var cost int64
		if red <= black {
			styles = append(styles, &Style{FillColor: drawing.Color{R: 0, G: 132, B: 199, A: 96}})
			cost = red
		} else {
			styles = append(styles, &Style{FillColor: drawing.Color{R: 147, G: 21, B: 152, A: 96}})
			cost = black
		}
		var target string
		for _, stat := range strings.Split(it, "&") {
			arr := strings.Split(stat, "+")
			target += fmt.Sprintf(statMap[arr[0]], arr[1])
		}
		ss = append(ss, []string{target, formatInt64(cost)})
	}
	p, err := TableOptionRender(TableChartOption{
		Width:           400,
		Header:          []string{fmt.Sprintf("%d级%s", nameLevel.level, s), "（底色表示魔方颜色）"},
		Data:            ss,
		HeaderFontColor: Color{R: 35, G: 35, B: 35, A: 255},
		CellStyle: func(cell TableCell) *Style {
			if cell.Column > 0 && cell.Row > 0 && len(cell.Text) > 0 {
				return styles[cell.Row-1]
			}
			return nil
		},
	})
	if err != nil {
		slog.Error("生成表格失败", "error", err)
		return nil
	}
	buf, err := p.Bytes()
	if err != nil {
		slog.Error("生成表格失败", "error", err)
		return nil
	}
	return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
}

var (
	//go:embed cubeRates.json
	cubeRatesJson []byte
	cubeRates     map[string]any
)

func init() {
	if err := json.Unmarshal(cubeRatesJson, &cubeRates); err != nil {
		panic(err)
	}
}

var cubeCost = map[string]int64{
	"red":    12000000,
	"black":  22000000,
	"master": 7500000,
}

func getRevealCostConstant(itemLevel int) float64 {
	switch {
	case itemLevel < 30:
		return 0.0
	case itemLevel <= 70:
		return 0.5
	case itemLevel <= 120:
		return 2.5
	default:
		return 20.0
	}
}

func cubingCost(cubeType string, itemLevel int, totalCubeCount float64) float64 {
	cubeCost := float64(cubeCost[cubeType])
	revealCostConst := getRevealCostConstant(itemLevel)
	revealPotentialCost := revealCostConst * float64(itemLevel*itemLevel)
	return cubeCost*totalCubeCount + totalCubeCount*revealPotentialCost
}

func getTierCosts(currentTier, desireTier int, cubeType string) DistrQuantileResult {
	var mean, median, seventyFifth, eightyFifth, ninetyFifth float64
	for i := currentTier; i < desireTier; i++ {
		p := tierRates[cubeType][i]
		stats := getDistrQuantile(p)
		mean += stats.mean
		median += stats.median
		seventyFifth += stats.seventyFifth
		eightyFifth += stats.eightyFifth
		ninetyFifth += stats.ninetyFifth
	}
	return DistrQuantileResult{mean: mean, median: median, seventyFifth: seventyFifth, eightyFifth: eightyFifth, ninetyFifth: ninetyFifth}
}

// Nexon rates: https://maplestory.nexon.com/Guide/OtherProbability/cube/strange
// GMS community calculated rates: https://docs.google.com/spreadsheets/d/1od_hep5Y6x2ljfrh4M8zj5RwlpgYDRn5uTymx4iLPyw/pubhtml#
// Nexon rates used when they match close enough to ours.
var tierRates = map[string]map[int]float64{
	"occult": {0: 0.009901},
	// Community rates are notably higher than nexon rates here. Assuming GMS is different and using those instead.
	"master": {0: 0.1184, 1: 0.0381},
	// Community rates are notably higher than nexon rates here. Assuming GMS is different and using those instead.
	// The sample size isn't great, but anecdotes from people in twitch chats align with the community data.
	// That being said, take meister tier up rates with a grain of salt.
	"meister": {0: 0.1163, 1: 0.0879, 2: 0.0459},
	// Community rates notably higher than KMS rates, using them.
	"red": {0: 0.14, 1: 0.06, 2: 0.025},
	// Community rates notably higher than KMS rates, using them.
	"black": {0: 0.17, 1: 0.11, 2: 0.05},
}

type DistrQuantileResult struct {
	mean, median, seventyFifth, eightyFifth, ninetyFifth float64
}

func getDistrQuantile(p float64) DistrQuantileResult {
	mean := 1 / p
	median := math.Log(1-0.5) / math.Log(1-p)
	seventyFifth := math.Log(1-0.75) / math.Log(1-p)
	eightyFifth := math.Log(1-0.85) / math.Log(1-p)
	ninetyFifth := math.Log(1-0.95) / math.Log(1-p)
	return DistrQuantileResult{mean: mean, median: median, seventyFifth: seventyFifth, eightyFifth: eightyFifth, ninetyFifth: ninetyFifth}
}

func runCalculator(itemType, cubeType string, currentTier, itemLevel, desiredTier int, desiredStat string) (int64, int64) {
	anyStats := len(desiredStat) == 0
	probabilityInputObject := translateInputToObject(desiredStat)
	p := 1.0
	if !anyStats {
		p = getProbability(desiredTier, probabilityInputObject, itemType, cubeType, itemLevel)
	}
	tierUp := getTierCosts(currentTier, desiredTier, cubeType)
	stats := DistrQuantileResult{}
	if !anyStats {
		stats = getDistrQuantile(p)
	}

	mean := stats.mean + tierUp.mean
	meanCost := cubingCost(cubeType, itemLevel, mean)
	return int64(math.Round(mean)), int64(math.Round(meanCost))
}

func getProbability(desiredTier int, probabilityInput map[string]int, itemType string, cubeType string, itemLevel int) float64 {
	// convert parts of input for easier mapping to keys in cubeRates
	tier := string(tierForNumber(desiredTier))
	itemLabel := itemType
	if itemType == "accessory" {
		itemLabel = "ring"
	} else if itemType == "badge" {
		itemLabel = "heart"
	}

	// get the cubing data for this input criteria from cubeRates (which is based on json data)
	rawCubedata := [][]any{
		cubeRates["lvl120to200"].(map[string]any)[itemLabel].(map[string]any)[cubeType].(map[string]any)[tier].(map[string]any)["first_line"].([]any),
		cubeRates["lvl120to200"].(map[string]any)[itemLabel].(map[string]any)[cubeType].(map[string]any)[tier].(map[string]any)["second_line"].([]any),
		cubeRates["lvl120to200"].(map[string]any)[itemLabel].(map[string]any)[cubeType].(map[string]any)[tier].(map[string]any)["third_line"].([]any),
	}

	// make adjustments to stat values if needed (for items lvl 160 or above)
	cubeData := convertCubeDataForLevel(rawCubedata, itemLevel)

	// generate consolidated version of cubing data that group any lines not relevant to the calculation into a single
	// Junk entry
	usefulCategories := getUsefulCategories(probabilityInput)
	consolidatedCubeData := [][][]any{
		getConsolidatedRates(cubeData[0], usefulCategories),
		getConsolidatedRates(cubeData[1], usefulCategories),
		getConsolidatedRates(cubeData[2], usefulCategories),
	}

	// loop through all possible outcomes for 1st, 2nd, and 3rd line using consolidated cube data
	// sum up the rate of outcomes that satisfied the input to determine final probability
	var totalChance = 0.0
	for _, line1 := range consolidatedCubeData[0] {
		for _, line2 := range consolidatedCubeData[1] {
			for _, line3 := range consolidatedCubeData[2] {
				// check if this outcome meets our needs
				outcome := [][]any{line1, line2, line3}
				ok := true
				for field, input := range probabilityInput {
					if !OUTCOME_MATCH_FUNCTION_MAP[field](outcome, input) {
						ok = false
					}
				}
				if ok { // calculate chance of this outcome occurring
					totalChance += calculateRate(outcome, consolidatedCubeData)
				}
			}
		}
	}
	return totalChance / 100.0
}

type Tier string

const (
	Rare      Tier = "rare"
	Epic      Tier = "epic"
	Unique    Tier = "unique"
	Legendary Tier = "legendary"
)

func tierForNumber(n int) Tier {
	switch n {
	case 0:
		return Rare
	case 1:
		return Epic
	case 2:
		return Unique
	case 3:
		return Legendary
	default:
		panic("unknown tier: " + strconv.Itoa(n))
	}
}

// translateInputToObject 将“lineMeso+1&lineStat+1”格式的string输入转化为map[string]int
func translateInputToObject(input string) map[string]int {
	output := make(map[string]int)
	if len(input) > 0 {
		vals := strings.Split(input, "&")
		for _, val := range vals {
			arr := strings.Split(val, "+")
			v, _ := strconv.Atoi(arr[1])
			output[arr[0]] += +v
		}
	}
	return output
}

// calculateRate 计算概率
func calculateRate(outcome [][]any, filteredRates [][][]any) float64 {
	// 对特殊的潜能条目进行调整。
	//
	// 参考： https://maplestory.nexon.com/Guide/OtherProbability/cube/strange
	getAdjustedRate := func(currentLine []any, previousLines, currentPool [][]any) float64 {
		currentCategory := currentLine[0].(string)
		currentRate := currentLine[2].(float64)

		// the first line will never have its rates adjusted
		if len(previousLines) == 0 {
			return currentRate
		}

		// determine special categories that we've reached the limit on in previous lines which need to be removed from
		// the current pool
		var (
			toBeRemoved           []string
			prevSpecialLinesCount = make(map[string]int)
		)
		for _, a := range previousLines {
			cat := a[0].(string)
			if _, ok := MAX_CATEGORY_COUNT[cat]; ok {
				prevSpecialLinesCount[cat]++
			}
		}

		// populate the list of special lines to be removed from the current pool
		// exit early with rate of 0 if this set of lines is not valid (exceeds max category count)
		for spCat, count := range prevSpecialLinesCount {
			if count > MAX_CATEGORY_COUNT[spCat] || spCat == currentCategory && (count+1) > MAX_CATEGORY_COUNT[spCat] {
				return 0.0
			} else if count == MAX_CATEGORY_COUNT[spCat] {
				toBeRemoved = append(toBeRemoved, spCat)
			}
		}
		var (
			// deduct total rate for each item that is removed from the pool
			adjustedTotal = 100.0
			// avoid doing math operations if the rate is not changing (due to floating point issues)
			adjustedFlag = false
		)
		for _, a := range currentPool {
			cat := a[0].(string)
			rate := a[2].(float64)
			if slices.Contains(toBeRemoved, cat) {
				adjustedTotal -= rate
				adjustedFlag = true
			}
		}
		if adjustedFlag {
			return currentRate / adjustedTotal * 100
		}
		return currentRate
	}

	adjustedRates := []float64{
		getAdjustedRate(outcome[0], nil, filteredRates[0]),
		getAdjustedRate(outcome[1], [][]any{outcome[0]}, filteredRates[1]),
		getAdjustedRate(outcome[2], [][]any{outcome[0], outcome[1]}, filteredRates[2]),
	}

	// calculate probability for this specific set of lines to occur
	var chance = 100.0
	for _, rate := range adjustedRates {
		chance *= rate / 100
	}

	return chance
}

// 计算联合概率
func getConsolidatedRates(ratesList []any, usefulCategories []string) [][]any {
	var (
		consolidatedRates [][]any
		junkRate          float64
		junkCategories    []any
	)

	for _, e := range ratesList {
		var (
			item     = e.([]any)
			category = item[0].(string)
			val      = item[1]
			rate     = item[2].(float64)
		)

		if slices.Contains(usefulCategories, category) || MAX_CATEGORY_COUNT[category] > 0 {
			consolidatedRates = append(consolidatedRates, item)
		} else if category == CATEGORY_JUNK {
			// using concat here since "Junk" is already a category that exists in the json data.
			// we're expanding it here with additional "contextual junk" based on the user input, so we want to preserve
			// the old list of junk categories too
			junkRate += rate
			for _, v := range val.([]any) {
				junkCategories = append(junkCategories, v.(string))
			}
		} else {
			junkRate += rate
			junkCategories = append(junkCategories, fmt.Sprintf("%s (%v)", category, val))
		}
	}

	consolidatedRates = append(consolidatedRates, []any{CATEGORY_JUNK, junkCategories, junkRate})
	return consolidatedRates
}

// getUsefulCategories 筛选有用的属性
func getUsefulCategories(probabilityInput map[string]int) []string {
	var usefulCategories []string
	for field, val := range INPUT_CATEGORY_MAP {
		input, ok := probabilityInput[field]
		if !ok {
			continue
		}
		if input > 0 {
			for _, v := range val {
				if !slices.Contains(usefulCategories, v) {
					usefulCategories = append(usefulCategories, v)
				}
			}
		}
	}
	return usefulCategories
}

// convertCubeDataForLevel 针对160+的装备，12%会变成13%
func convertCubeDataForLevel(cubeData [][]any, itemLevel int) [][]any {
	// don't need to make adjustments to items lvl <160
	if itemLevel < 160 {
		return cubeData
	}

	affectedCategories := []string{
		CATEGORY_STR_PERC, CATEGORY_LUK_PERC, CATEGORY_DEX_PERC, CATEGORY_INT_PERC,
		CATEGORY_ALLSTATS_PERC, CATEGORY_ATT_PERC, CATEGORY_MATT_PERC,
	}

	f := func(cubeDataLine []any) []any {
		var ret []any
		for _, e := range cubeDataLine {
			var (
				arr         = e.([]any)
				cat         = arr[0].(string)
				val         = arr[1]
				rate        = arr[2].(float64)
				adjustedVal = val
			)

			// adjust the value if this is an affected category
			for _, affectedCategory := range affectedCategories {
				if affectedCategory == cat {
					adjustedVal = val.(float64) + 1
					break
				}
			}
			ret = append(ret, any([]any{cat, adjustedVal, rate}))
		}
		return ret
	}
	return [][]any{f(cubeData[0]), f(cubeData[1]), f(cubeData[2])}
}

/**
 * calculateTotal 计算属性的总值
 *
 * calcVal false-只算条数，true-算值之和
 */
func calculateTotal(outcome [][]any, desiredCategory string, calcVal bool) int {
	if calcVal {
		var actualVal int
		for _, a := range outcome {
			category := a[0].(string)
			val := a[1]
			if category == desiredCategory {
				actualVal += int(val.(float64))
			}
		}
		return actualVal
	} else {
		var count int
		for _, a := range outcome {
			if a[0].(string) == desiredCategory {
				count++
			}
		}
		return count
	}
}

// OUTCOME_MATCH_FUNCTION_MAP 判断是否满足条件的函数
var OUTCOME_MATCH_FUNCTION_MAP = map[string]func([][]any, int) bool{
	"percStat": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_STR_PERC, true)+
			calculateTotal(outcome, CATEGORY_ALLSTATS_PERC, true) >= requiredVal
	},
	"lineStat": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_STR_PERC, false)+
			calculateTotal(outcome, CATEGORY_ALLSTATS_PERC, false) >= requiredVal
	},
	"percAllStat": func(outcome [][]any, requiredVal int) bool {
		var actualVal float64
		for _, a := range outcome {
			category := a[0].(string)
			{
				val := a[1].(float64)
				switch category { // （尖兵）力量、敏捷、运气都算作1/3全属性
				case CATEGORY_ALLSTATS_PERC:
					actualVal += val
				case CATEGORY_STR_PERC, CATEGORY_DEX_PERC, CATEGORY_LUK_PERC:
					actualVal += val / 3
				}
			}
		}
		return actualVal >= float64(requiredVal)
	},
	"lineAllStat": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_ALLSTATS_PERC, false) >= requiredVal
	},
	"percHp": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_MAXHP_PERC, true) >= requiredVal
	},
	"lineHp": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_MAXHP_PERC, false) >= requiredVal
	},
	"percAtt": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_ATT_PERC, true) >= requiredVal
	},
	"lineAtt": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_ATT_PERC, false) >= requiredVal
	},
	"percBoss": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_BOSSDMG_PERC, true) >= requiredVal
	},
	"lineBoss": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_BOSSDMG_PERC, false) >= requiredVal
	},
	"lineIed": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_IED_PERC, false) >= requiredVal
	},
	"lineCritDamage": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_CRITDMG_PERC, false) >= requiredVal
	},
	"lineMeso": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_MESO_PERC, false) >= requiredVal
	},
	"lineDrop": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_DROP_PERC, false) >= requiredVal
	},
	"lineMesoOrDrop": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_MESO_PERC, false)+calculateTotal(outcome, CATEGORY_DROP_PERC, false) >= requiredVal
	},
	"secCooldown": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_CDR_TIME, true) >= requiredVal
	},
	"lineAutoSteal": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_AUTOSTEAL_PERC, false) >= requiredVal
	},
	"lineAttOrBoss": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_ATT_PERC, false)+
			calculateTotal(outcome, CATEGORY_BOSSDMG_PERC, false) >= requiredVal
	},
	"lineAttOrBossOrIed": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_ATT_PERC, false)+
			calculateTotal(outcome, CATEGORY_BOSSDMG_PERC, false)+
			calculateTotal(outcome, CATEGORY_IED_PERC, false) >= requiredVal
	},
	"lineBossOrIed": func(outcome [][]any, requiredVal int) bool {
		return calculateTotal(outcome, CATEGORY_BOSSDMG_PERC, false)+
			calculateTotal(outcome, CATEGORY_IED_PERC, false) >= requiredVal
	},
}

const (
	CATEGORY_STR_PERC       = "STR %"
	CATEGORY_DEX_PERC       = "DEX %"
	CATEGORY_INT_PERC       = "INT %"
	CATEGORY_LUK_PERC       = "LUK %"
	CATEGORY_MAXHP_PERC     = "Max HP %"
	CATEGORY_MAXMP_PERC     = "Max MP %"
	CATEGORY_ALLSTATS_PERC  = "All Stats %"
	CATEGORY_ATT_PERC       = "ATT %"
	CATEGORY_MATT_PERC      = "MATT %"
	CATEGORY_BOSSDMG_PERC   = "Boss Damage"
	CATEGORY_IED_PERC       = "Ignore Enemy Defense %"
	CATEGORY_MESO_PERC      = "Meso Amount %"
	CATEGORY_DROP_PERC      = "Item Drop Rate %"
	CATEGORY_AUTOSTEAL_PERC = "Chance to auto steal %"
	CATEGORY_CRITDMG_PERC   = "Critical Damage %"
	CATEGORY_CDR_TIME       = "Skill Cooldown Reduction"
	CATEGORY_JUNK           = "Junk"
)

// only used for special line probability adjustment calculation
const (
	CATEGORY_DECENT_SKILL    = "Decent Skill"
	CATEGORY_INVINCIBLE_PERC = "Chance of being invincible for seconds when hit"
	CATEGORY_INVINCIBLE_TIME = "Increase invincibility time after being hit"
	CATEGORY_IGNOREDMG_PERC  = "Chance to ignore % damage when hit"
)

var (
	// INPUT_CATEGORY_MAP 有用的属性，例如：
	//
	// 如果需要力量%，则力量%和全属性%都是有用的属性
	//
	// 如果需要全属性%（尖兵），则除了全属性%以外，力量%|敏捷%|运气%都有用（算作1/3全属性%）
	INPUT_CATEGORY_MAP = map[string][]string{
		"percStat":           {CATEGORY_STR_PERC, CATEGORY_ALLSTATS_PERC},
		"lineStat":           {CATEGORY_STR_PERC, CATEGORY_ALLSTATS_PERC},
		"percAllStat":        {CATEGORY_ALLSTATS_PERC, CATEGORY_STR_PERC, CATEGORY_DEX_PERC, CATEGORY_LUK_PERC},
		"lineAllStat":        {CATEGORY_ALLSTATS_PERC},
		"percHp":             {CATEGORY_MAXHP_PERC},
		"lineHp":             {CATEGORY_MAXHP_PERC},
		"percAtt":            {CATEGORY_ATT_PERC},
		"lineAtt":            {CATEGORY_ATT_PERC},
		"percBoss":           {CATEGORY_BOSSDMG_PERC},
		"lineBoss":           {CATEGORY_BOSSDMG_PERC},
		"lineIed":            {CATEGORY_IED_PERC},
		"lineCritDamage":     {CATEGORY_CRITDMG_PERC},
		"lineMeso":           {CATEGORY_MESO_PERC},
		"lineDrop":           {CATEGORY_DROP_PERC},
		"lineMesoOrDrop":     {CATEGORY_DROP_PERC, CATEGORY_MESO_PERC},
		"secCooldown":        {CATEGORY_CDR_TIME},
		"lineAutoSteal":      {CATEGORY_AUTOSTEAL_PERC},
		"lineAttOrBoss":      {CATEGORY_ATT_PERC, CATEGORY_BOSSDMG_PERC},
		"lineAttOrBossOrIed": {CATEGORY_ATT_PERC, CATEGORY_BOSSDMG_PERC, CATEGORY_IED_PERC},
	}
	// MAX_CATEGORY_COUNT 以下属性最多出现1条或者2条
	MAX_CATEGORY_COUNT = map[string]int{
		CATEGORY_DECENT_SKILL:    1,
		CATEGORY_INVINCIBLE_TIME: 1,
		CATEGORY_IED_PERC:        3,
		CATEGORY_BOSSDMG_PERC:    3,
		CATEGORY_DROP_PERC:       3,
		CATEGORY_IGNOREDMG_PERC:  2,
		CATEGORY_INVINCIBLE_PERC: 2,
	}
)
