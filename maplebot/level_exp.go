package maplebot

import (
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"log/slog"
	"slices"
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
	format := func(i int64) string {
		f := float64(i)
		switch {
		case f < 1000.0:
			return fmt.Sprintf("%g", f)
		case f < 1000000.0:
			return fmt.Sprintf("%.2fK", f/1000)
		case f < 1000000000.0:
			return fmt.Sprintf("%.2fM", f/1000000)
		case f < 1000000000000.0:
			return fmt.Sprintf("%.2fB", f/1000000000)
		default:
			return fmt.Sprintf("%.2fT", f/1000000000000)
		}
	}
	var accumulate int64
	cur := make([]string, 0, 100)
	acc := make([]string, 0, 100)
	for i := 201; i <= 300; i++ {
		v := levelExpData.GetInt64(fmt.Sprintf("data.%d", i))
		cur = append(cur, format(v))
		acc = append(acc, format(accumulate))
		accumulate += v
	}
	labels := []string{"当前等级", "升级经验", "累计经验"}
	textAligns := []string{AlignRight, AlignRight, AlignRight}
	labels = slices.Concat(labels, labels, labels, labels)
	textAligns = slices.Concat(textAligns, textAligns, textAligns, textAligns)
	values := make([][]string, 0, 25)
	for i := 201; i <= 225; i++ {
		line := make([]string, 0, 12)
		for j := 0; j < 4; j++ {
			level := i + j*25
			line = append(line, strconv.Itoa(level), cur[level-201], acc[level-201])
		}
		values = append(values, line)
	}
	p, err := TableOptionRender(TableChartOption{
		Header:     labels,
		Data:       values,
		Width:      1100,
		TextAligns: textAligns,
		CellStyle: func(cell TableCell) *Style {
			if cell.Column%3 == 0 {
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

func calculateExpDamage(s string) MessageChain {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	switch {
	case i >= 5:
		return MessageChain{&Text{Text: "你比怪物等级高5级以上时，终伤+20%"}}
	case i > 0:
		return MessageChain{&Text{Text: fmt.Sprintf("你比怪物等级高%d级时，终伤+%d%%", i, i*2+10)}}
	case i == 0:
		return MessageChain{&Text{Text: "你和怪物等级相等时，终伤+10%"}}
	case i == -1:
		return MessageChain{&Text{Text: "你比怪物等级低1级时，终伤+5%"}}
	case i == -2:
		return MessageChain{&Text{Text: "你比怪物等级低2级时，终伤不变"}}
	case i == -3:
		return MessageChain{&Text{Text: "你比怪物等级低3级时，终伤-5%"}}
	case i > -40:
		return MessageChain{&Text{Text: fmt.Sprintf("你比怪物等级低%d级时，终伤%g%%", -i, 2.5*float64(i))}}
	default:
		return MessageChain{&Text{Text: "你比怪物等级低40级以上时，终伤-100%"}}
	}
}
