# 东方Project沙包聚集地机器人

## 配置文件

第一次运行会自动生成配置文件`config/org.tfcc.bot/TFCCConfig.yml`，如下：

```yaml
bilibili:
  area_v2: 236           # 直播分区，236-主机游戏
  mid: 12345678          # B站ID
  password: 12345678     # 密码
qq:
  super_admin_qq: 12345678  # 主管理员QQ号
  qq_group: # 主要功能的QQ群
    - 12345678
```

修改配置文件后重新启动即可

## 登录B站

第一次运行会提示扫码登录B站，此后会记录Cookies，无需再次登录。
如果提示Cookies超时，或者其他原因需要重新扫码，删除 `data/org.tfcc.bot/BilibiliData.yml` 即可。

## 功能一览

- [x] 管理员、白名单
- [x] B站开播、修改直播标题、查询直播状态
- [ ] 随作品、随机体
- [ ] B站视频解析
- [ ] B站直播解析
- [ ] B站视频推送
- [ ] 投票
- [ ] 查新闻
- [ ] 增加预约功能
- [ ] 查询分数表
- [ ] 打断复读
- [ ] 随符卡
- [ ] rep解析
- [x] 随机骰子、roll
- [ ] 抽签