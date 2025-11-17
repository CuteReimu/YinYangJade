package maplebot

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/CuteReimu/onebot"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

var ClassNameMap = map[string]string{
	"Hero":        "英雄",
	"Dark Knight": "黑骑士",
	"Paladin":     "圣骑士",

	"Ice/Lightning Archmage": "冰雷魔导师",
	"Arch Mage (I/L)":        "冰雷魔导师",
	"Fire/Poison Archmage":   "火毒魔导师",
	"Arch Mage (F/P)":        "火毒魔导师",
	"Bishop":                 "主教",

	"Shadower":     "侠盗(刀飞)",
	"Night Lord":   "隐士(镖飞)",
	"Blade Master": "暗影双刀",

	"Buccaneer":     "冲锋队长",
	"Corsair":       "船长",
	"Cannon Master": "神炮王",

	"Marksman":   "箭神",
	"Bowmaster":  "神射手",
	"Bow Master": "神射手",
	"Pathfinder": "古迹猎人",

	"Dawn Warrior":    "魂骑士",
	"Blaze Wizard":    "炎术士",
	"Wind Archer":     "风灵使者",
	"Night Walker":    "夜行者",
	"Thunder Breaker": "奇袭者",
	"Mihile":          "米哈尔",

	"Xenon":         "尖兵",
	"Battle Mage":   "幻灵斗师",
	"Wild Hunter":   "豹弩游侠",
	"Mechanic":      "机械师",
	"Demon Slayer":  "恶魔猎手",
	"Demon Avenger": "恶魔复仇者",
	"Blaster":       "爆破手",

	"Aran":     "战神",
	"Evan":     "龙神",
	"Mercedes": "双弩精灵",
	"Phantom":  "幻影",
	"Shade":    "隐月",
	"Luminous": "夜光法师",

	"Kaiser":         "狂龙战士",
	"Kain":           "该隐",
	"Cadena":         "卡德娜",
	"Angelic Buster": "爆莉萌天使",

	"Adele":  "阿黛尔",
	"Illium": "伊利温",
	"Ark":    "亚克",
	"Khali":  "卡莉",

	"Lara":    "菈菈",
	"Hoyoung": "虎影",
	"Ren":     "莲",

	"Hayato": "剑豪",
	"Kanna":  "阴阳师",

	"Zero":    "神之子",
	"Kinesis": "超能力者",

	"Lynn":        "琳恩",
	"Mo Xuan":     "墨玄",
	"Sia Astelle": "施亚",
}

var ClassIDMap = map[int]string{
	1:   "Warrior",
	2:   "Magician",
	3:   "Bowman",
	4:   "Thief",
	5:   "Pirate",
	10:  "Noblesse",
	11:  "Dawn Warrior",
	12:  "Blaze Wizard",
	13:  "Wind Archer",
	14:  "Night Walker",
	15:  "Thunder Breaker",
	20:  "Legend",
	21:  "Aran",
	22:  "Evan",
	23:  "Mercedes",
	24:  "Phantom",
	30:  "Citizen",
	31:  "Demon Slayer",
	32:  "Battle Mage",
	33:  "Wild Hunter",
	35:  "Mechanic",
	113: "Hero",
	123: "Paladin",
	133: "Dark Knight",
	201: "Jett",
	202: "Mihile",
	203: "Luminous",
	204: "Kaiser",
	205: "Angelic Buster",
	206: "Hayato",
	207: "Kanna",
	208: "Xenon",
	209: "Demon Avenger",
	210: "Zero",
	211: "Beast Tamer",
	212: "Shade",
	213: "Fire/Poison",
	214: "Kinesis",
	215: "Blaster",
	216: "Cadena",
	217: "Illium",
	218: "Ark",
	219: "Pathfinder",
	223: "Lara",
	220: "Hoyoung",
	221: "Adele",
	222: "Kain",
	224: "Khali",
	225: "Lynn",
	226: "Mo Xuan",
	227: "Sia Astelle",
}

func init() {
	// 全部转为小写
	classNameMap := make(map[string]string, len(ClassNameMap))
	for k, v := range ClassNameMap {
		classNameMap[strings.ToLower(k)] = v
	}
	ClassNameMap = classNameMap
}

func TranslateClassId(id int) string {
	if name, ok := ClassIDMap[id]; ok {
		return TranslateClassName(name)
	}
	if name, ok := ClassIDMap[id/100]; ok {
		return TranslateClassName(name)
	}
	return ""
}

func TranslateClassName(s string) string {
	if len(s) == 0 {
		return ""
	}
	if name, ok := ClassNameMap[strings.ToLower(s)]; ok {
		return name
	}
	return s
}

func GetClassImage(name string) (image.Image, error) {
	if len(name) == 0 {
		name = "Lynn"
	}
	if name := classImageData.GetString(strings.ToLower(name)); len(name) > 0 {
		img, err := getClassImage(name)
		if err != nil {
			return nil, err
		}
		return resize.Resize(0, 400, img, resize.Lanczos3), nil
	}
	return nil, errors.Errorf("class image not found: %s", name)
}

func GetClassOriginImageBuff(name string) ([]byte, error) {
	if name := classImageData.GetString(strings.ToLower(name)); len(name) > 0 {
		buf, err := os.ReadFile(filepath.Join("class_image", name))
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	return nil, errors.Errorf("class image not found: %s", name)
}

func getClassImage(name string) (image.Image, error) {
	buf, err := os.ReadFile(filepath.Join("class_image", name))
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(buf))
	return img, err
}

func SetClassImage(name string, img *Image) MessageChain {
	if len(name) == 0 || !classImageData.IsSet(name) {
		return MessageChain{&Text{Text: "职业不存在"}}
	}
	u := img.Url
	_, err := url.Parse(u)
	if err != nil {
		slog.Error("userInput is not a valid URL, reject it", "error", err)
		return nil
	}
	if err := os.MkdirAll("class_image", 0755); err != nil {
		slog.Error("mkdir failed", "error", err)
		return MessageChain{&Text{Text: "保存图片失败"}}
	}
	nameLen := len(filepath.Ext(img.File)) + 32
	if len(img.File) > nameLen {
		img.File = img.File[len(img.File)-nameLen:]
	}
	p := filepath.Join("class_image", img.File)
	cmd := exec.Command("curl", "-o", p, u)
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Error("cmd.Run() failed", "error", err)
		return MessageChain{&Text{Text: "保存图片失败"}}
	} else {
		slog.Debug(string(out))
	}
	_, err = getClassImage(img.File)
	if err != nil {
		slog.Error("invalid image", "error", err)
		_ = os.Remove(p)
		return MessageChain{&Text{Text: "不支持的图片格式"}}
	}
	classImageData.Set(name, img.File)
	if err = classImageData.WriteConfig(); err != nil {
		slog.Error("write config failed", "error", err)
	}
	return MessageChain{&Text{Text: "保存图片成功"}}
}

func clearExpiredImages2() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic recovered", "error", err)
		}
	}()
	data := make(map[string]bool)
	for name := range ClassNameMap {
		s := classImageData.GetString(name)
		if len(s) > 0 {
			data[s] = true
		}
	}
	files, err := os.ReadDir("class_image")
	if err != nil {
		slog.Error("read dir failed", "error", err)
	}
	for _, file := range files {
		if !data[file.Name()] {
			if err = os.Remove(filepath.Join("class_image", file.Name())); err != nil {
				slog.Error("remove file failed", "error", err)
			}
		}
	}
}
