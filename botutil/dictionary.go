package botutil

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/CuteReimu/YinYangJade/iface"
	. "github.com/CuteReimu/onebot"
	"github.com/spf13/viper"
)

// DealKey 处理词条的key，转换中文数字为阿拉伯数字
func DealKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "零", "0")
	s = strings.ReplaceAll(s, "一", "1")
	s = strings.ReplaceAll(s, "二", "2")
	s = strings.ReplaceAll(s, "三", "3")
	s = strings.ReplaceAll(s, "四", "4")
	s = strings.ReplaceAll(s, "五", "5")
	s = strings.ReplaceAll(s, "六", "6")
	s = strings.ReplaceAll(s, "七", "7")
	s = strings.ReplaceAll(s, "八", "8")
	s = strings.ReplaceAll(s, "九", "9")
	return strings.ToLower(s)
}

// SaveImage 保存消息链中的图片到本地
func SaveImage(message MessageChain, imageDir string, bot *Bot) (MessageChain, error) {
	for _, m := range message {
		if img, ok := m.(*Image); ok && len(img.Url) > 0 {
			u := img.Url
			_, err := url.Parse(u)
			if err != nil {
				slog.Error("userInput is not a valid URL, reject it", "error", err)
				return message, err
			}
			if err := os.MkdirAll(imageDir, 0755); err != nil {
				slog.Error("mkdir failed", "error", err)
				return message, errors.New("保存图片失败")
			}
			nameLen := len(filepath.Ext(img.File)) + 32
			if len(img.File) > nameLen {
				img.File = img.File[len(img.File)-nameLen:]
			}
			p := filepath.Join(imageDir, img.File)
			abs, err := filepath.Abs(p)
			if err != nil {
				slog.Error("filepath.Abs() failed", "error", err)
				return message, errors.New("保存图片失败")
			}
			cmd := exec.Command("curl", "-o", p, u)
			out, err := cmd.CombinedOutput()
			if err != nil {
				slog.Error("cmd.Run() failed", "error", err, "p", p, "u", u)
				return message, errors.New("保存图片失败")
			}
			slog.Debug(string(out))
			img.File = "file://" + abs
			img.Url = ""
		} else if forward, ok := m.(*Forward); ok {
			msgs, err := bot.GetForwardMessage(forward.Id)
			if err != nil {
				slog.Error("获取转发消息失败", "forwardId", forward.Id)
				return message, errors.New("获取转发消息失败")
			}
			var ret MessageChain
			for _, msg := range msgs {
				if m, ok := msg.(*GroupMessage); ok {
					newMsg, err := SaveImage(m.Message, imageDir, bot)
					if err != nil {
						slog.Error("嵌套saveImage失败", "error", err)
					}
					ret = append(ret, &Node{
						UserId:   strconv.FormatInt(m.UserId, 10),
						Nickname: m.Sender.Nickname,
						Content:  newMsg,
					})
				}
			}
			return ret, nil
		}
	}
	return message, nil
}

// ClearExpiredImages 清理过期的图片文件
func ClearExpiredImages(qunDb *viper.Viper, imageDir string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered", "error", err)
		}
	}()
	data := qunDb.GetStringMapString("data")
	data2 := make(map[string]bool)
	var f func(ms MessageChain)
	f = func(ms MessageChain) {
		for _, m := range ms {
			if img, ok := m.(*Image); ok && len(img.File) > 0 && strings.HasPrefix(img.File, "file://") {
				data2[filepath.Base(img.File[len("file://"):])] = true
			} else if node, ok := m.(*Node); ok {
				f(node.Content)
			}
		}
	}
	for _, v := range data {
		var ms MessageChain
		if err := json.Unmarshal([]byte(v), &ms); err != nil {
			slog.Error("json unmarshal failed", "error", err)
			continue
		}
		f(ms)
	}
	files, err := os.ReadDir(imageDir)
	if err != nil {
		slog.Error("read dir failed", "error", err)
		return
	}
	for _, file := range files {
		if !data2[file.Name()] {
			if err = os.Remove(filepath.Join(imageDir, file.Name())); err != nil {
				slog.Error("remove file failed", "error", err)
			}
		}
	}
}

// DictionaryCommand 词条命令结构
type DictionaryCommand struct {
	Name      string
	Tips      string
	CheckPerm bool
	AuthCheck func(int64) bool // 权限检查函数，由各模块提供
}

func (d *DictionaryCommand) GetName() string {
	return d.Name
}

func (d *DictionaryCommand) ShowTips(int64, int64) string {
	return d.Tips
}

func (d *DictionaryCommand) CheckAuth(_ int64, senderID int64) bool {
	return !d.CheckPerm || d.AuthCheck(senderID)
}

func (d *DictionaryCommand) Execute(_ *GroupMessage, content string) MessageChain {
	if len(strings.TrimSpace(content)) == 0 {
		return MessageChain{&Text{Text: "命令格式：\n" + d.Tips}}
	}
	return nil
}

// DealAddDict 处理添加词条
func DealAddDict(
	message *GroupMessage,
	key string,
	qunDb *viper.Viper,
	cmdMap map[string]iface.CmdHandler,
	addDbQQList map[int64]string,
	sendFunc func(*GroupMessage, ...SingleMessage),
) {
	if strings.Contains(key, ".") {
		sendFunc(message, &Text{Text: "词条名称中不能包含 . 符号"})
		return
	}
	if _, ok := cmdMap[key]; ok {
		sendFunc(message, &Text{Text: "不能用" + key + "作为词条"})
		return
	}
	if len(key) > 0 {
		m := qunDb.GetStringMapString("data")
		if _, ok := m[key]; ok {
			sendFunc(message, &Text{Text: "词条已存在"})
		} else {
			sendFunc(message, &Text{Text: "请输入要添加的内容"})
			addDbQQList[message.Sender.UserId] = key
		}
	}
}

// DealModifyDict 处理修改词条
func DealModifyDict(
	message *GroupMessage,
	key string,
	qunDb *viper.Viper,
	addDbQQList map[int64]string,
	sendFunc func(*GroupMessage, ...SingleMessage),
) {
	m := qunDb.GetStringMapString("data")
	if _, ok := m[key]; !ok {
		sendFunc(message, &Text{Text: "词条不存在"})
	} else {
		sendFunc(message, &Text{Text: "请输入要修改的内容"})
		addDbQQList[message.Sender.UserId] = key
	}
}

// DealRemoveDict 处理删除词条（通用版本，不包含特殊逻辑）
func DealRemoveDict(message *GroupMessage, key string, qunDb *viper.Viper, sendFunc func(*GroupMessage, ...SingleMessage)) bool {
	m := qunDb.GetStringMapString("data")
	if _, ok := m[key]; !ok {
		sendFunc(message, &Text{Text: "词条不存在"})
		return false
	}
	delete(m, key)
	qunDb.Set("data", m)
	if err := qunDb.WriteConfig(); err != nil {
		slog.Error("write data failed", "error", err)
	}
	sendFunc(message, &Text{Text: "删除词条成功"})
	return true
}

// DealSearchDict 处理搜索词条
func DealSearchDict(message *GroupMessage, key string, qunDb *viper.Viper, sendFunc func(*GroupMessage, ...SingleMessage)) {
	var res []string
	m := qunDb.GetStringMapString("data")
	for k := range m {
		if strings.Contains(k, key) {
			res = append(res, k)
		}
	}
	if len(res) > 0 {
		slices.Sort(res)
		num := len(res)
		if num > 10 {
			res = res[:10]
			res[9] += fmt.Sprintf("\n等%d个词条", num)
		}
		for i := range res {
			res[i] = fmt.Sprintf("%d. %s", i+1, res[i])
		}
		sendFunc(message, &Text{Text: "搜索到以下词条：\n" + strings.Join(res, "\n")})
	} else {
		sendFunc(message, &Text{Text: "搜索不到词条(" + key + ")"})
	}
}

// FillSpecificMessage 填充消息链中的文件路径图片为base64
func FillSpecificMessage(messages []SingleMessage) {
	for _, m := range messages {
		if img, ok := m.(*Image); ok && len(img.File) > 0 {
			if strings.HasPrefix(img.File, "file://") {
				fileName := img.File[len("file://"):]
				buf, err := os.ReadFile(fileName)
				if err != nil {
					slog.Error("read file failed", "error", err)
					continue
				}
				img.File = "base64://" + base64.StdEncoding.EncodeToString(buf)
			}
		} else if node, ok := m.(*Node); ok {
			FillSpecificMessage(node.Content)
		}
	}
}
