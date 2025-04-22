package imageutil

import (
	"bytes"
	"encoding/base64"
	. "github.com/CuteReimu/onebot"
	"github.com/pkg/errors"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func RemoveBackground(buf []byte, rate int) ([]byte, error) {
	if rate <= 0 || rate >= 100 {
		return nil, errors.New("rate必须在1-99之间")
	}
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		slog.Error("图片解码失败", "error", err)
		return nil, errors.New("只支持gif、jpeg、png格式的图片")
	}
	countCache := make(map[uint8]int)
	totalCount := img.Bounds().Dx() * img.Bounds().Dy()
	for x := range img.Bounds().Dx() {
		for y := range img.Bounds().Dy() {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			v := (r + g + b) / 3
			countCache[uint8(v>>8)]++
		}
	}
	var sum int
	v0 := uint32(255)
	for ; v0 > 0; v0-- {
		sum += countCache[uint8(v0)]
		if sum*100 >= totalCount*rate {
			break
		}
	}
	img0 := image.NewRGBA(img.Bounds())
	for x := range img.Bounds().Dx() {
		for y := range img.Bounds().Dy() {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			v := (r + g + b) / 3
			if v>>8 >= v0 {
				img0.Set(x, y, color.Alpha{})
			} else {
				img0.Set(x, y, c)
			}
		}
	}
	buffer := &bytes.Buffer{}
	err = png.Encode(buffer, img0)
	if err != nil {
		slog.Error("png编码失败", "error", err)
		return nil, errors.New("png编码失败")
	}
	return buffer.Bytes(), nil
}

var B *Bot

func Init(b *Bot) {
	B = b
	B.ListenPrivateMessage(handlePrivateMessage)
}

var alphaImageList = make(map[int64]int)

func handlePrivateMessage(message *PrivateMessage) bool {
	if message.SubType != PrivateMessageFriend {
		return true
	}
	if len(message.Message) == 1 {
		if text, ok := message.Message[0].(*Text); ok {
			if strings.HasPrefix(text.Text, "抠图 ") {
				rate, err := strconv.Atoi(strings.TrimSpace(text.Text[len("抠图"):]))
				if err == nil {
					if rate <= 0 || rate >= 100 {
						sendPrivateMessage(message, &Text{Text: "抠图范围必须在1-99之间"})
					} else {
						alphaImageList[message.Sender.UserId] = rate
						sendPrivateMessage(message, &Text{Text: "请输入要抠图的图片"})
					}
				}
				return true
			}
		}
	}
	if key, ok := alphaImageList[message.Sender.UserId]; ok { // 抠图
		delete(alphaImageList, message.Sender.UserId)
		if len(message.Message) != 1 {
			sendPrivateMessage(message, &Text{Text: "提供的不是一张图片，抠图失败"})
		} else if img, ok := message.Message[0].(*Image); !ok {
			sendPrivateMessage(message, &Text{Text: "提供的不是一张图片，抠图失败"})
		} else if len(img.Url) == 0 {
			sendPrivateMessage(message, &Text{Text: "无法识别图片，抠图失败"})
		} else {
			buf, err := getImage(img)
			if err != nil {
				sendPrivateMessage(message, &Text{Text: "无法识别图片，抠图失败"})
			} else {
				buf, err = RemoveBackground(buf, key)
				if err != nil {
					sendPrivateMessage(message, &Text{Text: err.Error()})
				} else {
					sendPrivateMessage(message, &Image{File: "base64://" + base64.StdEncoding.EncodeToString(buf)})
				}
			}
		}
		return true
	}
	return true
}

func getImage(img *Image) ([]byte, error) {
	u := img.Url
	_, err := url.Parse(u)
	if err != nil {
		slog.Error("userInput is not a valid URL, reject it", "error", err)
		return nil, err
	}
	if err := os.MkdirAll("temp_image", 0755); err != nil {
		slog.Error("mkdir failed", "error", err)
		return nil, err
	}
	defer func() { _ = os.RemoveAll("temp_image") }()
	p := filepath.Join("temp_image", img.File)
	cmd := exec.Command("curl", "-o", p, u)
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Error("cmd.Run() failed", "error", err)
		return nil, err
	} else {
		slog.Debug(string(out))
	}
	buf, err := os.ReadFile(p)
	if err != nil {
		slog.Error("read file failed", "error", err)
		return nil, err
	}
	return buf, nil
}

func sendPrivateMessage(context *PrivateMessage, messages ...SingleMessage) {
	if len(messages) == 0 {
		return
	}
	f := func(messages []SingleMessage) error {
		_, err := B.SendPrivateMessage(context.UserId, messages)
		return err
	}
	if err := f(messages); err != nil {
		slog.Error("send private message failed", "error", err)
		newMessages := make([]SingleMessage, 0, len(messages))
		for _, m := range messages {
			if img, ok := m.(*Image); !ok || !strings.HasPrefix(img.File, "http") {
				newMessages = append(newMessages, m)
			}
		}
		if len(newMessages) != len(messages) && len(newMessages) > 0 {
			if err = f(newMessages); err != nil {
				slog.Error("send group message failed", "error", err)
			}
		}
	}
}
