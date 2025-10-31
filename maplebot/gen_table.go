package maplebot

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"

	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
)

func genTable(s string) (ret MessageChain) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic recovered", "error", r, "stack", string(debug.Stack()))
			ret = MessageChain{&Text{Text: fmt.Sprint(r)}}
		}
	}()
	width := 600
	if strings.HasPrefix(s, "w=") {
		index := strings.Index(s, " ")
		if index >= 0 {
			w, err := strconv.Atoi(s[2:index])
			if err == nil {
				width = w
				s = strings.TrimSpace(s[index:])
			}
		}
	}
	r := bufio.NewReader(bytes.NewReader([]byte(s)))
	var labels []string
	var values [][]string
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return MessageChain{&Text{Text: err.Error()}}
		}
		l := strings.TrimSpace(string(line))
		a := strings.Split(l, ",")
		if labels == nil {
			labels = a
		} else {
			values = append(values, a)
		}
	}
	p, err := TableOptionRender(TableChartOption{
		Header: labels,
		Data:   values,
		Width:  width,
	})
	if err != nil {
		slog.Error("render chart failed", "error", err)
		return MessageChain{&Text{Text: err.Error()}}
	} else if buf, err := p.Bytes(); err != nil {
		slog.Error("render chart failed", "error", err)
		return MessageChain{&Text{Text: err.Error()}}
	} else {
		return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)}}
	}
}
