// Package iface 定义了聊天指令处理器的接口
package iface

import (
	. "github.com/CuteReimu/onebot"
)

// CmdHandler 这是聊天指令处理器的接口，当你想要新增自己的聊天指令处理器时，实现这个接口即可
type CmdHandler interface {
	// Name 群友输入聊天指令时，第一个空格前的内容。
	Name() string
	// ShowTips 在【帮助列表】中应该如何显示这个命令。空字符串表示不显示
	ShowTips(groupCode int64, senderID int64) string
	// CheckAuth 如果他有权限执行这个指令，则返回True，否则返回False
	CheckAuth(groupCode int64, senderID int64) bool
	// Execute content参数是除开指令名（第一个空格前的部分）以外剩下的所有内容。返回值是要发送的群聊消息为空就是不发送消息。
	Execute(msg *GroupMessage, content string) MessageChain
}

// SimpleCmdHandler 这是一个简单的CmdHandler实现，可以直接使用它来快速创建一个命令处理器
type SimpleCmdHandler struct {
	HandlerName string
	HandlerTips string
	Handler     func(string) MessageChain
}

func (s *SimpleCmdHandler) Name() string {
	return s.HandlerName
}

func (s *SimpleCmdHandler) ShowTips(int64, int64) string {
	return s.HandlerTips
}

func (*SimpleCmdHandler) CheckAuth(int64, int64) bool {
	return true
}

func (s *SimpleCmdHandler) Execute(_ *GroupMessage, content string) MessageChain {
	return s.Handler(content)
}
