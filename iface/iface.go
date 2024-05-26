package iface

import (
	. "github.com/CuteReimu/onebot"
)

// CmdHandler 这是聊天指令处理器的接口，当你想要新增自己的聊天指令处理器时，实现这个接口即可
type CmdHandler interface {
	// Name 群友输入聊天指令时，第一个空格前的内容。
	Name() string
	// ShowTips 在【帮助列表】中应该如何显示这个命令。空字符串表示不显示
	ShowTips(groupCode int64, senderId int64) string
	// CheckAuth 如果他有权限执行这个指令，则返回True，否则返回False
	CheckAuth(groupCode int64, senderId int64) bool
	// Execute content参数是除开指令名（第一个空格前的部分）以外剩下的所有内容。返回值是要发送的群聊消息为空就是不发送消息。
	Execute(msg *GroupMessage, content string) MessageChain
}
