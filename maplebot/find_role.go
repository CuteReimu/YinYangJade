package maplebot

import (
	"bytes"
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/onebot"
	"github.com/tidwall/gjson"
	. "github.com/vicanso/go-charts/v2"
	"log/slog"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	nameRegex   = regexp.MustCompile(`<h3 class="card-title text-nowrap">([A-Za-z0-9 ]+)</h3>`)
	imgRegex    = regexp.MustCompile(`<img src="(.*?)"`)
	levelRegex  = regexp.MustCompile(`<h5 class="card-text">([A-Za-z0-9.% ()]+)</h5>`)
	classRegex  = regexp.MustCompile(`<p class="card-text mb-0">([A-Za-z0-9 ]*?) in`)
	legionRegex = regexp.MustCompile(`Legion Level <span class="char-stat-right">([0-9,]+)</span>`)
	dataRegex   = regexp.MustCompile(`"data":\s*\{`)
)

func findRole(name string) MessageChain {
	resp, err := restyClient.R().Get("https://mapleranks.com/u/" + name)
	if err != nil {
		slog.Error("请求失败", "error", err)
		return nil
	}
	switch resp.StatusCode() {
	case 404:
		return MessageChain{&Text{Text: name + "已身死道消"}}
	case 200:
	default:
		slog.Error("请求失败", "status", resp.StatusCode())
		return nil
	}
	body := resp.Body()
	var (
		rawName     = ""
		class       = ""
		levelExp    = ""
		legionLevel = "0"
		imgUrl      = ""
		expData     gjson.Result
		levelData   gjson.Result
	)
	nameMatch := nameRegex.FindSubmatchIndex(body)
	if nameMatch != nil {
		rawName = string(body[nameMatch[2]:nameMatch[3]])
		body = body[nameMatch[1]:]
	}
	imgMatch := imgRegex.FindSubmatchIndex(body)
	if imgMatch != nil {
		imgUrl = string(body[imgMatch[2]:imgMatch[3]])
	}
	levelMatch := levelRegex.FindSubmatchIndex(body)
	if levelMatch != nil {
		levelExp = string(body[levelMatch[2]:levelMatch[3]])
		if index := strings.Index(levelExp, "Lv."); index != -1 {
			levelExp = levelExp[index+len("Lv."):]
		}
		levelExp = strings.TrimSpace(levelExp)
	}
	classMatch := classRegex.FindSubmatchIndex(body)
	if classMatch != nil {
		class = string(body[classMatch[2]:classMatch[3]])
	}
	legionMatch := legionRegex.FindSubmatchIndex(body)
	if legionMatch != nil {
		legionLevel = string(body[legionMatch[2]:legionMatch[3]])
	}
	body = resp.Body()
	for dataMatch := dataRegex.FindIndex(body); dataMatch != nil; dataMatch = dataRegex.FindIndex(body) {
		body = body[dataMatch[1]:]
		index := bytes.Index(body, []byte("},"))
		rawData := "{" + string(body[:index]) + "}"
		chartData := gjson.Parse(rawData)
		a := chartData.Get("datasets").Array()
		if len(a) > 0 {
			switch a[0].Get("label").String() {
			case "Level":
				levelData = chartData
			case "Exp":
				expData = chartData
			}
		}
	}

	var messageChain MessageChain
	if len(imgUrl) > 0 {
		resp, err := restyClient.R().Get(imgUrl)
		if err != nil {
			slog.Error("请求失败", "error", err)
		} else {
			messageChain = append(messageChain, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(resp.Body())})
		}
	}

	s := fmt.Sprintf("角色名：%s\n职业：%s\n等级：%s\n联盟：%s\n", rawName, class, levelExp, legionLevel)

	expValues := make([]float64, 0, 30)
	levelValues := make([]float64, 0, 30)
	labels := make([]string, 0, 30)
	a := expData.Get("datasets").Array()
	if len(a) > 0 {
		datasets := a[0].Get("data").Array()
		for i, label := range expData.Get("labels").Array() {
			if s := label.String(); len(s) > 0 {
				labels = append(labels, s)
				expValues = append(expValues, float64(datasets[i].Int()))
			}
		}
		maxValue := slices.Max(expValues)
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
		var yAxisOptions []YAxisOption
		yAxisOptions = append(yAxisOptions, YAxisOption{Min: NewFloatPoint(0), Max: &maxValue, DivideCount: divideCount, Unit: 1})
		aLevel := levelData.Get("datasets").Array()
		if len(aLevel) > 0 {
			datasetsLevel := aLevel[0].Get("data").Array()
			for i, label := range levelData.Get("labels").Array() {
				if s := label.String(); len(s) > 0 && i < len(datasetsLevel) {
					levelValues = append(levelValues, datasetsLevel[i].Float())
				}
			}
			for i := range levelValues {
				if levelValues[i] == 0 && i > 0 {
					levelValues[i] = levelValues[i-1]
				}
			}
			for i := len(levelValues) - 2; i >= 0; i-- {
				if levelValues[i] == 0 {
					levelValues[i] = levelValues[i+1]
				}
			}
			for _, diff := range []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 20, 50, 100} {
				minLevel, maxLevel := math.Floor(slices.Min(levelValues)/diff)*diff, math.Ceil(slices.Max(levelValues)/diff)*diff
				remainCount := divideCount - int(math.Round((maxLevel-minLevel)/diff))
				if remainCount >= 0 {
					if remainCount > 0 {
						maxLevel += diff * float64(remainCount-remainCount/2)
						minLevel -= diff * float64(remainCount/2)
					}
					yAxisOptions = append(yAxisOptions, YAxisOption{Min: NewFloatPoint(minLevel), Max: NewFloatPoint(maxLevel), DivideCount: divideCount, Unit: 1})
					break
				}
			}
		}
		if slices.ContainsFunc(expValues, func(f float64) bool { return f != 0 }) {
			if levelEnd := strings.Index(levelExp, "("); levelEnd >= 0 {
				if totalExp := float64(levelExpData.GetInt64("data." + strings.TrimSpace(levelExp[:levelEnd]))); totalExp > 0 {
					if expPercentStart := strings.Index(levelExp, "("); expPercentStart >= 0 {
						if expPercentEnd := strings.Index(levelExp, "%"); expPercentEnd >= 0 {
							if expPercent, err := strconv.ParseFloat(levelExp[expPercentStart+1:expPercentEnd], 64); err == nil {
								var sumExp, n float64
								for _, v := range expValues {
									if v != 0 || n > 0 {
										sumExp += v
										n++
									}
								}
								aveExp := sumExp / float64(len(expValues))
								days := int(math.Ceil((totalExp - totalExp/100.0*expPercent) / aveExp))
								s += fmt.Sprintf("预计还有%d天升级\n", days)
							}
						}
					}
				}
			}
			var seriesList []Series
			seriesList = append(seriesList, NewSeriesFromValues(expValues, ChartTypeBar))
			if len(levelValues) == len(expValues) {
				seriesList = append(seriesList, NewSeriesFromValues(levelValues, ChartTypeLine))
				seriesList[1].AxisIndex = 1
			}
			p, err := Render(
				ChartOption{SeriesList: seriesList},
				ThemeOptionFunc(ThemeDark),
				PaddingOptionFunc(Box{Top: 30, Left: 10, Right: 10, Bottom: 10}),
				XAxisDataOptionFunc(labels),
				//MarkLineOptionFunc(0, SeriesMarkDataTypeAverage),
				YAxisOptionFunc(yAxisOptions...),
				func(opt *ChartOption) {
					opt.XAxis.TextRotation = -math.Pi / 4
					opt.XAxis.LabelOffset = Box{Top: -5, Left: 7}
					opt.XAxis.FontSize = 7.5
					opt.XAxis.FirstAxis = 1
					opt.ValueFormatter = func(f float64) string {
						switch {
						case f < 1000.0:
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
				},
			)
			if err != nil {
				slog.Error("render chart failed", "error", err)
			} else if buf, err := p.Bytes(); err != nil {
				slog.Error("render chart failed", "error", err)
			} else {
				messageChain = append(messageChain, &Text{Text: s}, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)})
			}
		} else {
			messageChain = append(messageChain, &Text{Text: s + "近日无经验变化"})
		}
	} else {
		messageChain = append(messageChain, &Text{Text: s + "近日无经验变化"})
	}
	return messageChain
}
