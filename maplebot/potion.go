package maplebot

import (
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"log/slog"
	"strconv"
)

func calculatePotion() MessageChain {
	potions := map[string]int64{
		"敲头":   levelExpData.GetInt64("data.199"),
		"209药": levelExpData.GetInt64("data.209"),
		"219药": levelExpData.GetInt64("data.219"),
		"229药": levelExpData.GetInt64("data.229"),
		"239药": levelExpData.GetInt64("data.239"),
		"249药": levelExpData.GetInt64("data.249"),
	}
	keys := []string{"敲头", "209药", "219药", "229药", "239药", "249药"}
	ss := make([][]string, 0, 51)
	for i := 210; i <= 260; i++ {
		values := make([]string, 0, len(keys)+1)
		values = append(values, strconv.Itoa(i))
		for _, key := range keys {
			value := potions[key]
			need := levelExpData.GetInt64(fmt.Sprintf("data.%d", i))
			v := min(100.0, float64(value)/float64(need)*100)
			values = append(values, fmt.Sprintf("%.2f%%", v))
		}
		ss = append(ss, values)
	}
	p, err := TableOptionRender(TableChartOption{
		Width:           600,
		Header:          append([]string{"等级"}, keys...),
		Data:            ss,
		HeaderFontColor: Color{R: 35, G: 35, B: 35, A: 255},
		CellStyle: func(cell TableCell) *Style {
			if cell.Column > 1 && cell.Row > 0 && len(cell.Text) > 0 {
				v, _ := strconv.ParseFloat(cell.Text[:len(cell.Text)-1], 64)
				if v >= 50 {
					text := ss[cell.Row-1][cell.Column-1]
					v2, _ := strconv.ParseFloat(text[:len(text)-1], 64)
					if v > v2*2 {
						return &Style{FillColor: drawing.Color{R: 255, G: 130, B: 171, A: 128}}
					}
				} else if cell.Column == len(keys) {
					return &Style{FillColor: drawing.Color{R: 255, G: 130, B: 171, A: 128}}
				}
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
