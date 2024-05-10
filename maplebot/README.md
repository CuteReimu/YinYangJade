# GMSR群机器人

## 配置

第一次启动后，会在 `config/net.cutereimu.maplebots` 下生成相关配置文件。

* 配置文件 `Config.yml` ：

```yaml
# 生效的QQ群
qq_groups:
  - 12345678
# 管理员QQ号
admin: 12345678
# 图片超时时间（单位：小时）
image_expire_hours: 72
```

根据自己的需要修改配置文件后，重启即可。

> [!IMPORTANT]
> 想要正确使用词条功能，必须让本项目的运行目录相对于 Mirai 的运行目录为 `../YinYangJade`，即：
> 
> ```console
> $ tree -C -L 2
> .
> ├── mirai
> │   ├── libs
> │   ├── mcl
> │   ├── mcl.jar
> │   ├── plugins
> ├── YinYangJade
> │   ├── config.yml
> │   ├── YinYangJade
> ```