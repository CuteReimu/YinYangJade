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
