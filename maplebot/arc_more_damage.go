// Package maplebot 实现冒险岛GMSR群机器人功能
package maplebot

import (
	"encoding/base64"
	"log/slog"
	"strconv"

	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
)

// BossArcData 表示boss所需的ARC数据
type BossArcData struct {
	Name    string
	NeedArc int
}

var bossArc = []*BossArcData{
	{Name: "Lucid", NeedArc: 360},
	{Name: "Will", NeedArc: 760},
	{Name: "Gloom", NeedArc: 730},
	{Name: "Hilla", NeedArc: 900},
	{Name: "Darknell", NeedArc: 850},
	{Name: "BlackMage", NeedArc: 1320},
}

// GetMoreDamageArc 生成更多伤害所需ARC的图表
func GetMoreDamageArc() (ret MessageChain) {
	data := make([][]string, 0, len(bossArc))
	for _, v := range bossArc {
		data = append(data, []string{
			v.Name,
			strconv.Itoa(v.NeedArc),
			strconv.Itoa(v.NeedArc * 11 / 10),
			strconv.Itoa(v.NeedArc * 13 / 10),
			strconv.Itoa(v.NeedArc * 15 / 10),
		})
	}
	p, err := TableOptionRender(TableChartOption{
		Header:     []string{"Boss", "100%", "110%", "130%", "150%"},
		Data:       data,
		Width:      480,
		TextAligns: []string{AlignLeft, AlignLeft, AlignLeft, AlignLeft, AlignLeft},
	})
	if err != nil {
		slog.Error("render chart failed", "error", err)
		return nil
	}
	buf, err := p.Bytes()
	if err != nil {
		slog.Error("render chart failed", "error", err)
		return nil
	}
	ret = append(ret, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)})
	return ret
}
