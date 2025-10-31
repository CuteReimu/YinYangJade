package slicegame

import (
	"bytes"
	"cmp"
	"encoding/base64"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"log/slog"
	"math/rand/v2"
	"slices"
	"strconv"

	"github.com/CuteReimu/goutil"
	"github.com/CuteReimu/neuquant"
	. "github.com/CuteReimu/onebot"
	. "github.com/vicanso/go-charts/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func DoStuff() MessageChain {
	problem := []int{1, 2, 3, 4, 5, 6, 7, 8}
	tryCount := 0
	var reverseCount int
	for {
		tryCount++
		if tryCount > 100 {
			return MessageChain{&Text{Text: "可能出现了意想不到的问题"}}
		}
		rand.Shuffle(len(problem), func(i, j int) {
			problem[i], problem[j] = problem[j], problem[i]
		})
		reverseCount = 0
		for i := range problem {
			for j := range i {
				if problem[j] > problem[i] {
					reverseCount++
				}
			}
		}
		if reverseCount&1 == 0 && dist(problem) >= 10 {
			problem = append(problem, 0)
			break
		}
	}
	openSet := goutil.NewPriorityQueue(nil, func(o1, o2 *Problem) int {
		return cmp.Compare(o1.dist, o2.dist)
	})
	img := new(gif.GIF)
	openSet.Add(&Problem{
		hash:    hash(problem),
		dist:    dist(problem),
		problem: problem,
	})
	closeSet := make(map[int]bool)
	results := make(map[int]*Result)
	results[hash(problem)] = &Result{}
	for openSet.Len() > 0 {
		p := openSet.Poll()
		r0 := results[p.hash]
		if p.hash == 123456780 {
			if err := displayResult(img, p.hash, results); err != nil {
				return MessageChain{&Text{Text: err.Error()}}
			}
			break
		}
		index0 := slices.Index(p.problem, 0)
		for _, idx := range directions[index0] {
			newProblem := slices.Clone(p.problem)
			newProblem[index0], newProblem[idx] = newProblem[idx], newProblem[index0]
			p2 := &Problem{
				hash:    hash(newProblem),
				dist:    dist(newProblem) + r0.dist + 1,
				problem: newProblem,
			}
			if !closeSet[p2.hash] {
				openSet.Add(p2)
			}
			if r, ok := results[p2.hash]; ok && r0.dist+1 >= r.dist {

			} else {
				results[p2.hash] = &Result{
					lastHash: p.hash,
					dist:     r0.dist + 1,
					which:    idx,
					toWhich:  index0,
				}
			}
		}
		closeSet[p.hash] = true
	}
	if len(img.Delay) > 0 {
		img.Delay[len(img.Delay)-1] *= 3
	}
	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, img)
	if err != nil {
		return MessageChain{&Text{Text: err.Error()}}
	}
	return MessageChain{&Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf.Bytes())}}
}

func displayResult(img *gif.GIF, h int, results map[int]*Result) error {
	r := results[h]
	if r.dist != 0 {
		if err := displayResult(img, r.lastHash, results); err != nil {
			return err
		}
		if err := drawImage(img, h); err != nil {
			return err
		}
	} else {
		if err := drawImage(img, h); err != nil {
			return err
		}
	}
	return nil
}

func drawImage(img *gif.GIF, hash int) error {
	problem := make([]string, 0, 9)
	for range 9 {
		if hash%10 != 0 {
			problem = append(problem, strconv.Itoa(hash%10))
		} else {
			problem = append(problem, "")
		}
		hash /= 10
	}
	slices.Reverse(problem)
	p, err := TableOptionRender(TableChartOption{
		Width:           102,
		Header:          problem[:3],
		Data:            [][]string{problem[3:6], problem[6:]},
		HeaderFontColor: Color{A: 255},
		FontColor:       Color{A: 255},
		CellStyle: func(cell TableCell) *Style {
			return &Style{FillColor: drawing.Color{R: 255, G: 255, B: 255, A: 255}}
		},
	})
	if err != nil {
		slog.Error("生成表格失败", "error", err)
		return err
	}
	buf, err := p.Bytes()
	if err != nil {
		slog.Error("生成表格失败", "error", err)
		return err
	}
	img0, err := png.Decode(bytes.NewReader(buf))
	if err != nil {
		slog.Error("解析png失败", "error", err)
		return err
	}
	if img0.Bounds().Dx() != 102 || img0.Bounds().Dy() != 102 {
		img1 := image.NewRGBA(image.Rect(0, 0, 102, 102))
		draw.Draw(img1, img1.Bounds(), image.White, image.Point{}, draw.Src)
		draw.Draw(img1, img0.Bounds(), img0, image.Point{}, draw.Over)
		img0 = img1
	}
	img.Image = append(img.Image, neuquant.Paletted(img0))
	img.Delay = append(img.Delay, 100)
	return nil
}

type Result struct {
	lastHash int
	dist     int
	which    int
	toWhich  int
}

type Problem struct {
	hash, dist int
	problem    []int
}

var directions = map[int][]int{
	0: {1, 3},
	1: {0, 2, 4},
	2: {1, 5},
	3: {0, 4, 6},
	4: {1, 3, 5, 7},
	5: {2, 4, 8},
	6: {3, 7},
	7: {4, 6, 8},
	8: {5, 7},
}

func dist(a []int) int {
	d := 0
	for i, v := range a {
		if v == 0 {
			continue
		}
		v--
		d += max(i%3-v%3, v%3-i%3)
		d += max(i/3-v/3, v/3-i/3)
	}
	return d
}

func hash(a []int) int {
	h := 0
	for _, v := range a {
		h = h*10 + v
	}
	return h
}
