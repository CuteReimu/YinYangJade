# 空洞骑士speedrun推送小助手

## 配置文件：

第一次运行会自动生成配置文件`config/net.cutereimu.hkbot/HKConfig.yml`，如下：

```yaml
# 是否启用推送
enable: true
# speedrun推送间隔
speedrun_push_delay: 300
# speedrun推送QQ群
speedrun_push_qq_group:
  - 12345678
# speedrun的API Key
speedrun_api_key: xxxxxxxx
qq:
  # 管理员QQ
  super_admin_qq: 12345678
```

修改配置文件后重新启动即可