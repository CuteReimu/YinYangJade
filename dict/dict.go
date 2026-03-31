// Package dict 提供基于 TF-IDF 的文本分词、词频统计和相似度计算功能。
package dict

import (
	"cmp"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"time"
	"unicode"

	"github.com/CuteReimu/YinYangJade/db"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-ego/gse"
	"github.com/pkg/errors"
)

var seg gse.Segmenter

func init() {
	if err := seg.LoadDictEmbed(); err != nil {
		slog.Error("load gse dict failed", "error", err)
	}
}

// cutWords 对文本进行分词，并过滤掉纯标点/空白词
func cutWords(text string) []string {
	raw := seg.Cut(text, true)
	result := make([]string, 0, len(raw))
	for _, w := range raw {
		// 过滤全部由标点或空白组成的词
		allPunct := true
		for _, r := range w {
			if !unicode.IsPunct(r) && !unicode.IsSymbol(r) && !unicode.IsSpace(r) {
				allPunct = false
				break
			}
		}
		if allPunct {
			continue
		}
		result = append(result, w)
	}
	return result
}

// getInt 从 BadgerDB 事务中读取一个整数值，key 不存在时返回 0
func getInt(txn *badger.Txn, key string) (int64, error) {
	item, err := txn.Get([]byte(key))
	if errors.Is(err, badger.ErrKeyNotFound) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	buf, err := item.ValueCopy(nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(buf), 10, 64)
}

// setInt 在 BadgerDB 事务中写入一个整数值
func setInt(txn *badger.Txn, key string, val int64) error {
	return txn.Set([]byte(key), []byte(strconv.FormatInt(val, 10)))
}

const (
	keyTFPrefix = "dict:tf:"
	keyDFPrefix = "dict:df:"
	keyN        = "dict:n"
)

// AddIntoDict 当收到用户文字聊天时，调用此函数。
//
// 此函数会统计词频，得到每个词的权重
func AddIntoDict(text string) {
	words := cutWords(text)
	if len(words) == 0 {
		return
	}

	// 统计本文档中每个词的出现次数，用于去重计算 df
	localCount := make(map[string]int64)
	for _, w := range words {
		localCount[w]++
	}

	// 使用 badger 事务原子更新，遇到事务冲突时重试
	i := 0
	for {
		err := db.DB.Update(func(txn *badger.Txn) error {
			// 更新总文档数
			n, err := getInt(txn, keyN)
			if err != nil {
				return err
			}
			if err := setInt(txn, keyN, n+1); err != nil {
				return err
			}

			for w, cnt := range localCount {
				// 更新 tf（全局词频）
				tfKey := keyTFPrefix + w
				tf, err := getInt(txn, tfKey)
				if err != nil {
					return err
				}
				if err := setInt(txn, tfKey, tf+cnt); err != nil {
					return err
				}

				// 更新 df（文档频率），每个词每篇文档只计一次
				dfKey := keyDFPrefix + w
				df, err := getInt(txn, dfKey)
				if err != nil {
					return err
				}
				if err := setInt(txn, dfKey, df+1); err != nil {
					return err
				}
			}
			return nil
		})
		if errors.Is(err, badger.ErrConflict) {
			i++
			if i < 5 {
				time.Sleep(200 * time.Millisecond) // 等待一段时间后重试
				continue
			}
		}
		if err != nil {
			slog.Error("AddIntoDict failed", "error", err)
		}
		return
	}
}

// getWordWeight 获取一个词在指定文档中的 TF-IDF 权重
//
// 影响一个词(Term)在一篇文档中的重要性主要有两个因素：
//   - 词频率（Term Frequency，简称tf）：即此Term在此文档中出现了多少次，越大说明越重要。
//   - 文档频率(Document Frequency，简称df)：即有多少文档包含此Term，越大说明越不重要。
//
// 公式：W = tf × log(n/df)
func getWordWeight(word string, docWordCount map[string]int64, n int64) float64 {
	tf := docWordCount[word]
	if tf == 0 {
		return 0
	}

	// 从语料库中读取文档频率，如果查不到说明该词从未在语料中出现过，
	// 使用 0.5 做平滑，使罕见词获得更高的权重。
	var df float64
	dfStr, ok := db.Get(keyDFPrefix + word)
	if ok {
		parsed, err := strconv.ParseInt(dfStr, 10, 64)
		if err == nil && parsed > 0 {
			df = float64(parsed)
		}
	}

	return float64(tf) * math.Log(float64(max(n, 1))/max(df, 0.1))
}

// GetTextRelativity 获取两个文本的相似度
//
// 使用 TF-IDF 向量的余弦相似度：cos(θ) = (A·B) / (|A|×|B|)
func GetTextRelativity(text1, text2 string) float64 {
	words1 := cutWords(text1)
	words2 := cutWords(text2)
	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	// 统计每篇文档中的词频
	count1 := make(map[string]int64)
	for _, w := range words1 {
		count1[w]++
	}
	count2 := make(map[string]int64)
	for _, w := range words2 {
		count2[w]++
	}

	// 读取总文档数
	nStr, ok := db.Get(keyN)
	if !ok {
		nStr = "0"
	}
	n, _ := strconv.ParseInt(nStr, 10, 64)
	if n == 0 {
		// 没有文档数据时，回退到简单的词频向量余弦相似度
		n = 1
	}

	// 收集所有词构建词汇表
	vocab := make(map[string]struct{})
	for w := range count1 {
		vocab[w] = struct{}{}
	}
	for w := range count2 {
		vocab[w] = struct{}{}
	}

	// 计算向量点积和模
	var dot, norm1, norm2 float64
	for w := range vocab {
		w1 := getWordWeight(w, count1, n)
		w2 := getWordWeight(w, count2, n)
		dot += w1 * w2
		norm1 += w1 * w1
		norm2 += w2 * w2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dot / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

func GetFamiliarValue(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}

	type pair struct {
		k string
		v float64
	}

	var cache []pair
	for k := range m {
		v := GetTextRelativity(k, key)
		if v >= 0.65 {
			cache = append(cache, pair{k: k, v: v})
		}
	}
	slices.SortFunc(cache, func(a, b pair) int {
		return cmp.Compare(b.v, a.v)
	})
	if len(cache) > 0 && cache[0].v >= 0.9 || len(cache) > 1 && cache[1].v >= 0.8 || len(cache) > 2 {
		return m[cache[0].k]
	}
	return ""
}
