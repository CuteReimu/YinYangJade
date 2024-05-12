package maplebot

import (
	"bytes"
	"encoding/base64"
	"fmt"
	. "github.com/CuteReimu/mirai-sdk-http"
	"github.com/tidwall/gjson"
	. "github.com/vicanso/go-charts/v2"
	"log/slog"
	"regexp"
	"slices"
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
		return MessageChain{&Plain{Text: name + "已身死道消"}}
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
		chartData   gjson.Result
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
		chartData = gjson.Parse(rawData)
		a := chartData.Get("datasets").Array()
		if len(a) > 0 && a[0].Get("label").String() != "Exp" {
			chartData = gjson.Result{}
		} else {
			break
		}
	}

	var messageChain MessageChain
	if len(imgUrl) > 0 {
		messageChain = append(messageChain, &Image{Url: imgUrl})
	}

	s := fmt.Sprintf("角色名：%s\n职业：%s\n等级：%s\n联盟：%s\n", rawName, class, levelExp, legionLevel)

	values := make([]float64, 0, 14)
	labels := make([]string, 0, 14)
	a := chartData.Get("datasets").Array()
	if len(a) > 0 {
		datasets := a[0].Get("data").Array()
		for i, label := range chartData.Get("labels").Array() {
			if s := label.String(); len(s) > 0 {
				labels = append(labels, s)
				values = append(values, float64(datasets[i].Int()))
			}
		}
		if trimLen := len(labels) - 14; trimLen > 0 {
			labels = labels[trimLen:]
			values = values[trimLen:]
		}
		if slices.ContainsFunc(values, func(f float64) bool { return f != 0 }) {
			p, err := HorizontalBarRender([][]float64{values}, YAxisDataOptionFunc(labels))
			if err != nil {
				slog.Error("render chart failed", "error", err)
			} else if buf, err := p.Bytes(); err != nil {
				slog.Error("render chart failed", "error", err)
			} else {
				messageChain = append(messageChain, &Plain{Text: s}, &Image{Base64: base64.StdEncoding.EncodeToString(buf)})
			}
		} else {
			messageChain = append(messageChain, &Plain{Text: s + "近日无经验变化"})
		}
	} else {
		messageChain = append(messageChain, &Plain{Text: s + "近日无经验变化"})
	}
	return messageChain
}