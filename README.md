# 阴阳神玉

![](https://img.shields.io/github/languages/top/CuteReimu/YinYangJade "语言")
[![](https://img.shields.io/github/actions/workflow/status/CuteReimu/YinYangJade/golangci-lint.yml?branch=master)](https://github.com/CuteReimu/YinYangJade/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/CuteReimu/YinYangJade)](https://github.com/CuteReimu/YinYangJade/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/CuteReimu/YinYangJade)](https://github.com/CuteReimu/YinYangJade/blob/master/LICENSE "许可协议")

这是以下几个机器人项目的合集，采用[mirai](https://github.com/mamoe/mirai)框架的[mirai-api-http](https://github.com/project-mirai/mirai-api-http)接口。

- [东方Project沙包聚集地机器人](tfcc)
- [空洞骑士群机器人](https://github.com/CuteReimu/hollow-knight-speedrun-bot)
- [GMSR群机器人](maplebot)
- [风声群机器人](fengsheng)

> [!NOTE]
> 为什么要修改成这种形式？
> 
> 考虑到原先写成Kotlin插件的形式，每次更新都需要重启[mirai](https://github.com/mamoe/mirai)并重新登录。
> 而重新登录很不稳定，因此改为使用[mirai-api-http](https://github.com/project-mirai/mirai-api-http)插件，把所有功能都独立成单独的进程，通过ws接口进行通信。
> 这样每次更新代码时，只需要重启机器人的进程，而不需重启[mirai](https://github.com/mamoe/mirai)和重新登录了。

## 开始

在使用本项目之前，你应该知道如何使用[mirai](https://github.com/mamoe/mirai)进行登录，并安装[mirai-api-http](https://github.com/project-mirai/mirai-api-http)插件。

请多参阅mirai-api-http的[文档](https://docs.mirai.mamoe.net/mirai-api-http/api/API.html)

本项目使用ws接口，因此你需要修改mirai的配置文件`config/net.mamoe.mirai-api-http/setting.yml`，开启ws监听。

```yaml
adapters:
  - ws
verifyKey: ABCDEFGHIJK
adapterSettings:
  ws:
    ## websocket server 监听的本地地址
    ## 一般为 localhost 即可, 如果多网卡等情况，自定设置
    host: localhost

    ## websocket server 监听的端口
    ## 与 http server 可以重复, 由于协议与路径不同, 不会产生冲突
    port: 8080

    ## 就填-1
    reservedSyncId: -1
```

## 编译

```shell
go build -o YinYangJade
```

## 运行

第一次运行会生成配置文件`config.yaml`，请根据实际情况修改配置文件后重新运行。

```yml
# 和上面那个ws的host保持一致
host: localhost

# 和上面那个ws的port保持一致
port: 8080

# 你的机器人的QQ号
qq: 123456789

# 和上面的那个verifyKey保持一致
verifykey: ABCDEFGHIJK
```
