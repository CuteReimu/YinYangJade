# 风声群机器人

## 配置

第一次启动后，会在 `config/com.fengsheng.bot` 下生成相关配置文件。

* 配置文件 `FengshengConfig.yml` ：

```yaml
# QQ相关配置
qq:
  # 管理员QQ号
  super_admin_qq: 12345678
  # 生效的QQ群
  qq_group: 
    - 12345678
# 对应风声服务端的GM入口
fengshengUrl: 'http://127.0.0.1:9094'
# 图片超时时间（单位：小时）
image_expire_hours: 24
```

根据自己的需要修改配置文件后，重启即可。