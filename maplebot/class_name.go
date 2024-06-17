package maplebot

import (
	"strings"
)

var ClassNameMap = map[string]string{
	"Hero":                   "英雄",
	"Dark Knight":            "黑骑士",
	"Paladin":                "圣骑士",
	"Ice/Lightning Archmage": "冰雷魔导师",
	"Fire/Poison Archmage":   "火毒魔导师",
	"Bishop":                 "主教",
	"Shadower":               "侠盗(刀飞)",
	"Night Lord":             "隐士(镖飞)",
	"Blade Master":           "暗影双刀",
	"Buccaneer":              "冲锋队长",
	"Corsair":                "船长",
	"Cannoneer":              "火炮手",
	"Marksman":               "箭神",
	"Bowmaster":              "神射手",
	"Pathfinder":             "古迹猎人",
	"Dawn Warrior":           "魂骑士",
	"Blaze Wizard":           "炎术士",
	"Wind Archer":            "风灵使者",
	"Night Walker":           "夜行者",
	"Thunder Breaker":        "奇袭者",
	"Mihile":                 "米哈尔",
	"Xenon":                  "尖兵",
	"Battle Mage":            "幻灵斗师",
	"Wild Hunter":            "豹弩游侠",
	"Mechanic":               "机械师",
	"Demon Slayer":           "恶魔猎手",
	"Demon Avenger":          "恶魔复仇者",
	"Blaster":                "爆破手",
	"Aran":                   "战神",
	"Evan":                   "龙神",
	"Mercedes":               "双弩精灵",
	"Phantom":                "幻影",
	"Shade":                  "隐月",
	"Luminous":               "夜光法师",
	"Kaiser":                 "狂龙战士",
	"Kain":                   "炼狱黑客",
	"Cadena":                 "魔链影士",
	"Angelic Buster":         "爆莉萌天使",
	"Adele":                  "御剑骑士",
	"llium":                  "圣晶使徒",
	"Ark":                    "影魂异人",
	"Khali":                  "飞刃沙士",
	"Lara":                   "元素师",
	"Hoyoung":                "虎影",
	"Hayato":                 "剑豪",
	"Kanna":                  "阴阳师",
	"ZERO":                   "神之子",
	"Lynn":                   "琳恩",
	"Kinesis":                "超能力者",
}

func init() {
	// 全部转为小写
	classNameMap := make(map[string]string, len(ClassNameMap))
	for k, v := range ClassNameMap {
		classNameMap[strings.ToLower(k)] = v
	}
	ClassNameMap = classNameMap
}

func TranslateClassName(s string) string {
	if name, ok := ClassNameMap[strings.ToLower(s)]; ok {
		return name
	}
	return s
}
