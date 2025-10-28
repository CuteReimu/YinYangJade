package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/CuteReimu/YinYangJade/db"
	"github.com/CuteReimu/YinYangJade/fengsheng"
	"github.com/CuteReimu/YinYangJade/hkbot"
	"github.com/CuteReimu/YinYangJade/imageutil"
	"github.com/CuteReimu/YinYangJade/maplebot"
	"github.com/CuteReimu/YinYangJade/tfcc"
	"github.com/CuteReimu/onebot"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var mainConfig = viper.New()

func init() {
	writerError, err := rotatelogs.New(
		path.Join("logs", "log-%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		slog.Error("unable to write logs", "error", err)
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(writerError, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format("15:04:05.000"))
				}
			case slog.SourceKey:
				if s, ok := a.Value.Any().(*slog.Source); ok {
					if index := strings.LastIndex(s.File, "@"); index >= 0 {
						if index += strings.Index(s.File[index:], string(filepath.Separator)); index >= 0 {
							s.File = s.File[index+1:]
						}
					}
					const projectName = "/YinYangJade/"
					if index := strings.LastIndex(s.File, projectName); index >= 0 {
						s.File = s.File[index+len(projectName):]
					}
				}
			default:
				if e, ok := a.Value.Any().(error); ok {
					a.Value = slog.StringValue(fmt.Sprintf("%+v", e))
				}
			}
			return a
		}})))

	mainConfig.SetConfigName("config")
	mainConfig.SetConfigType("yml")
	mainConfig.AddConfigPath(".")
	mainConfig.SetDefault("host", "localhost")
	mainConfig.SetDefault("port", 8080)
	mainConfig.SetDefault("verifyKey", "ABCDEFGHIJK")
	mainConfig.SetDefault("qq", 123456789)
	mainConfig.SetDefault("check_qq_groups", []int64(nil))
	if err := mainConfig.SafeWriteConfig(); err == nil {
		fmt.Println("Already generated config.yaml. Please modify the config file and restart the program.")
		os.Exit(0)
	}
	if err := mainConfig.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
}

var B *onebot.Bot

func main() {
	db.Init()
	defer db.Stop()
	var err error
	host := mainConfig.GetString("host")
	port := mainConfig.GetInt("port")
	verifyKey := mainConfig.GetString("verifyKey")
	qq := mainConfig.GetInt64("qq")
	B, err = onebot.Connect(host, port, onebot.WsChannelAll, verifyKey, qq, false)
	if err != nil {
		slog.Error("connect failed", "error", err)
		panic(err)
	}
	B.SetLimiter("drop", rate.NewLimiter(rate.Every(3*time.Second), 5))
	tfcc.Init(B)
	fengsheng.Init(B)
	maplebot.Init(B)
	hkbot.Init(B)
	imageutil.Init(B)
	B.ListenFriendRequest(handleNewFriendRequest)
	B.ListenGroupRequest(handleGroupRequest)
	checkQQGroups()
	initCron()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func checkQQGroups() {
	go func() {
		for range time.Tick(30 * time.Second) {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic recover", "error", r)
					}
				}()
				groups := mainConfig.GetIntSlice("check_qq_groups")
				groupList, err := B.GetGroupList()
				if err != nil {
					slog.Error("get group list failed", "error", err)
					return
				}
				for _, group := range groupList {
					if !slices.Contains(groups, int(group.GroupId)) {
						if err = B.SetGroupLeave(group.GroupId, false); err != nil {
							slog.Error("quit group failed", "error", err)
						}
					}
				}
			}()
		}
	}()
}

func handleNewFriendRequest(request *onebot.FriendRequest) bool {
	groupList, err := B.GetGroupList()
	if err != nil {
		slog.Error("获取群列表失败", "error", err)
		err = B.SetFriendAddRequest(request.Flag, false, "")
		if err != nil {
			slog.Error("处理好友请求失败", "error", err)
		}
	} else {
		for _, groupInfo := range groupList {
			memberList, err := B.GetGroupMemberList(groupInfo.GroupId)
			if err != nil {
				slog.Error("获取群成员列表失败", "error", err)
				continue
			}
			for _, member := range memberList {
				if member.UserId == request.UserId {
					err = B.SetFriendAddRequest(request.Flag, true, "")
					if err != nil {
						slog.Error("处理好友请求失败", "error", err)
					}
					return true
				}
			}
		}
		err = B.SetFriendAddRequest(request.Flag, false, "")
		if err != nil {
			slog.Error("处理好友请求失败", "error", err)
		}
	}
	return true
}

func handleGroupRequest(request *onebot.GroupRequest) bool {
	if request.SubType == onebot.GroupRequestInvite {
		var approve bool
		groups := mainConfig.GetIntSlice("check_qq_groups")
		if slices.Contains(groups, int(request.GroupId)) {
			approve = true
		} else {
			approve = false
		}
		err := B.SetGroupAddRequest(request.Flag, request.SubType, approve, "")
		if err != nil {
			slog.Error("处理邀请请求失败", "approve", approve, "error", err)
		} else {
			slog.Info("处理邀请请求成功", "approve", approve, "error", err)
		}
	} else if request.SubType == onebot.GroupRequestAdd {
		if strings.Contains(request.Comment, "管理员你好") {
			err := B.SetGroupAddRequest(request.Flag, request.SubType, false, "")
			if err != nil {
				slog.Error("拒绝申请请求失败", "approve", false, "error", err)
			} else {
				slog.Info("拒绝申请请求成功", "approve", false, "error", err)
			}
		}
	}
	return true
}

func initCron() {
	c := cron.New()
	_ = c.AddFunc("0 0 11 * * *", maplebot.FindRoleBackground)
	_ = c.AddFunc("0 0 17 * * *", maplebot.FindRoleBackground)
	_ = c.AddFunc("0 0 23 * * *", maplebot.FindRoleBackground)
	c.Start()
}
