package hkbot

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"

	regexp "github.com/dlclark/regexp2"
)

var translateDict = &trie{}
var regexpSpace = regexp.MustCompile(`(?<![()\[\]{}%'"A-Za-z]) (?![()\[\]{}%'"A-Za-z])`, regexp.None)

//go:embed translate.csv
var transLateData []byte

func init() {
	reader := bufio.NewReader(bytes.NewReader(transLateData))
	for {
		line, _, err := reader.ReadLine()
		if err != nil && err != io.EOF {
			panic(err)
		}
		if len(line) > 0 {
			arr := strings.Split(string(line), ",")
			var key, val string
			key = arr[0]
			if len(arr) >= 2 {
				val = arr[1]
			}
			if !translateDict.Put(key, val) {
				panic(fmt.Sprint("出现重复数据：", string(line)))
			}
		}
		if err == io.EOF {
			break
		}
	}
	translateDict.Put("beat the WR", "打破了世界纪录：")
	translateDict.Put("in Hollow Knight Category Extensions", "在空洞骑士")
	translateDict.Put("in Hollow Knight Category Extensions -", "在空洞骑士")
	translateDict.Put("in Hollow Knight: Silksong Category Extensions", "在丝之歌")
	translateDict.Put("in Hollow Knight: Silksong Category Extensions -", "在丝之歌")
	translateDict.Put("King's Pass: Level", "国王山道")
	translateDict.Put("in Hollow Knight", "在空洞骑士")
	translateDict.Put("in Hollow Knight -", "在空洞骑士")
	translateDict.Put("The new WR is", "新的世界纪录是")
	translateDict.Put("Their time is", "时间是")
	translateDict.Put("Its time is", "时间是")
	translateDict.Put("The time is", "时间是")
	translateDict.Put("His time is", "时间是")
	translateDict.Put("Her time is", "时间是")
	translateDict.Put("got a new top 3 PB", "获得了前三：")
	translateDict.Put("Pantheon of the Master: Level", "大师万神殿")
	translateDict.Put("Pantheon of the Master Level", "大师万神殿")
	translateDict.Put("Pantheon of the Artist: Level", "艺术家万神殿")
	translateDict.Put("Pantheon of the Artist Level", "艺术家万神殿")
	translateDict.Put("Pantheon of the Sage: Level", "贤者万神殿")
	translateDict.Put("Pantheon of the Sage Level", "贤者万神殿")
	translateDict.Put("Pantheon of the Knight: Level", "骑士万神殿")
	translateDict.Put("Pantheon of the Knight Level", "骑士万神殿")
	translateDict.Put("Pantheon of Hallownest: Level", "圣巢万神殿")
	translateDict.Put("Pantheon of Hallownest Level", "圣巢万神殿")
	translateDict.Put("White Palace: Level", "白色宫殿")
	translateDict.Put("White Palace Level", "白色宫殿")
	translateDict.Put("Path of Pain: Level", "苦痛之路")
	translateDict.Put("Path of Pain Level", "苦痛之路")
	translateDict.Put("Trial of the Warrior: Level", "勇士的试炼")
	translateDict.Put("Trial of the Warrior Level", "勇士的试炼")
	translateDict.Put("Trial of the Conqueror: Level", "征服者的试炼")
	translateDict.Put("Trial of the Conqueror Level", "征服者的试炼")
	translateDict.Put("Trial of the Fool: Level", "愚人的试炼")
	translateDict.Put("Trial of the Fool Level", "愚人的试炼")
	translateDict.Put("NMG", "无主要邪道")
	translateDict.Put("NMG.", "无主要邪道.")
	translateDict.Put("- NMG", "- 无主要邪道")
	translateDict.Put("- NMG.", "- 无主要邪道.")
	translateDict.Put("Console Runs", "主机速通")
	translateDict.Put("Any Bindings", "任意锁")
	translateDict.Put("Abyss Climb: Level", "深渊攀爬")
	translateDict.Put("Abyss Climb Level", "深渊攀爬")
	translateDict.Put("King's Pass: Level", "国王山道")
	translateDict.Put("King's Pass Level", "国王山道")
	translateDict.Put("NG", "新存档")
	translateDict.Put("NG+", "无需新存档")
	translateDict.Put("NMG NG+", "无主要邪道无需新存档")
	translateDict.Put("Warpless", "禁SL")
	translateDict.Put("Hollow Knight: Silksong", "丝之歌")
	translateDict.Put("Judgement", "末日裁决者")
	translateDict.Put("Twisted", "畸芽")
	translateDict.Put("Tools", "工具")
	translateDict.Put("Awoo", "全跳蚤")
	translateDict.Put("Individual Level", "")
	translateDict.Put("Moss Grotto", "苔穴")
	translateDict.Put("Dapper Slapper", "衣冠楚楚的掌掴者")
	translateDict.Put("Wishes", "祈愿")
	translateDict.Put("Slab%", "监狱钟道%")
	translateDict.Put("Jailed", "被锁住")
	translateDict.Put("Bellbeast", "钟道兽")
}

func translate(s string) string {
	s = translateDict.ReplaceAll(s)
	s, err := regexpSpace.Replace(s, "", -1, -1)
	if err != nil {
		panic(err)
	}
	return s
}

type trieNode struct {
	child  map[rune]*trieNode
	value  string
	exists bool
}

type trie struct {
	root trieNode
}

func (t *trie) Put(key, value string) bool {
	node := &t.root
	for _, c := range strings.ToLower(key) {
		var n *trieNode
		if node.child == nil {
			node.child = make(map[rune]*trieNode)
		}
		n = node.child[c]
		if n != nil {
			node = n
		} else {
			newNode := &trieNode{}
			node.child[c] = newNode
			node = newNode
		}
	}
	notExists := !node.exists
	node.exists = true
	node.value = value
	return notExists
}

func (t *trie) ReplaceAll(str string) string {
	s := []rune(str)
	var s2 []rune
	for len(s) > 0 {
		if !(len(s2) == 0 || symbols[s2[len(s2)-1]]) {
			s2 = append(s2, s[0])
			s = s[1:]
			continue
		}
		key, value := t.getLongest(string(s))
		if len(key) > 0 {
			s2 = append(s2, []rune(value)...)
			s = s[len([]rune(key)):]
		} else {
			s2 = append(s2, s[0])
			s = s[1:]
		}
	}
	return string(s2)
}

func (t *trie) getLongest(s string) (key, value string) {
	var node, node2 *trieNode
	var key2 string
	node = &t.root
	r := []rune(strings.ToLower(s))
	for idx, c := range r {
		if node.child != nil {
			if n, ok := node.child[c]; ok {
				key += string(c)
				node = n
				if node.exists && (idx+1 >= len(s) || symbols[r[idx+1]]) {
					node2 = node
					key2 = key
				}
				continue
			}
		}
		break
	}
	if node2 != nil {
		return key2, node2.value
	}
	return "", ""
}

var symbols = map[rune]bool{
	' ':  true,
	'(':  true,
	')':  true,
	'[':  true,
	']':  true,
	'-':  true,
	'{':  true,
	'}':  true,
	'%':  true,
	'\'': true,
	'"':  true,
	',':  true,
	'.':  true,
}
