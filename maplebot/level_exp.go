package maplebot

import (
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"log/slog"
	"math"
	"strconv"
)

func calculateExpBetweenLevel(start, end int64) MessageChain {
	if start < 1 || end > 300 || start >= end {
		return nil
	}
	var exp int64
	for i := start; i < end; i++ {
		exp += int64(levelExpData.GetFloat64(fmt.Sprintf("data.%d", i)))
	}
	var s string
	switch {
	case exp < 1000:
		s = strconv.FormatInt(exp, 10)
	case exp < 1000000:
		s = fmt.Sprintf("%.2fK", float64(exp)/1000)
	case exp < 1000000000:
		s = fmt.Sprintf("%.2fM", float64(exp)/1000000)
	case exp < 1000000000000:
		s = fmt.Sprintf("%.2fB", float64(exp)/1000000000)
	case exp < 1000000000000000:
		s = fmt.Sprintf("%.2fT", float64(exp)/1000000000000)
	default:
		s = fmt.Sprintf("%.2fQ", float64(exp)/1000000000000000)
	}
	return MessageChain{&Text{Text: fmt.Sprintf("从%d级到%d级需要经验：%s", start, end, s)}}
}

func calculateLevelExp() MessageChain {
	labels := make([]string, 0, 100)
	values := [][]float64{make([]float64, 0, 100)}
	for i := 200; i < 300; i++ {
		labels = append(labels, strconv.Itoa(i))
		values[0] = append(values[0], math.Log10(levelExpData.GetFloat64(fmt.Sprintf("data.%d", i))))
	}
	p, err := LineRender(
		values,
		ThemeOptionFunc(ThemeDark),
		HeightOptionFunc(600),
		WidthOptionFunc(1000),
		PaddingOptionFunc(Box{Top: 30, Left: 10, Right: 10, Bottom: 10}),
		XAxisDataOptionFunc(labels),
		YAxisOptionFunc(YAxisOption{Min: NewFloatPoint(9), Max: NewFloatPoint(16), DivideCount: 7, Unit: 1}),
		func(opt *ChartOption) {
			opt.XAxis.TextRotation = -math.Pi / 4
			opt.XAxis.LabelOffset = Box{Top: -5, Left: 7}
			opt.XAxis.FontSize = 7.5
			opt.XAxis.FirstAxis = 1
			opt.XAxis.SplitNumber = 5
			opt.XAxis.FirstAxis = -2
			opt.ValueFormatter = func(f float64) string {
				f = math.Pow(10, f)
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
		return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
	}
	return nil
}
