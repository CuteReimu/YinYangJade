# 阴阳神玉

![](https://img.shields.io/github/languages/top/CuteReimu/YinYangJade "语言")
[![](https://img.shields.io/github/actions/workflow/status/CuteReimu/YinYangJade/golangci-lint.yml?branch=master)](https://github.com/CuteReimu/YinYangJade/actions/workflows/golangci-lint.yml "代码分析")
[![](https://img.shields.io/github/contributors/CuteReimu/YinYangJade)](https://github.com/CuteReimu/YinYangJade/graphs/contributors "贡献者")
[![](https://img.shields.io/github/license/CuteReimu/YinYangJade)](https://github.com/CuteReimu/YinYangJade/blob/master/LICENSE "许可协议")

这是以下几个机器人项目的合集，基于 [onebot-11](https://github.com/botuniverse/onebot-11) 接口编写。

- [东方Project沙包聚集地机器人](tfcc)
- [空洞骑士speedrun推送小助手](hkbot)
- [GMSR群机器人](maplebot)
- [风声群机器人](fengsheng)

## 开始

本项目只含有业务逻辑，不负责QQ机器人的连接与认证、收发消息等功能。

在使用本项目之前，你应该首先自行搭建一个支持 [onebot-11](https://github.com/botuniverse/onebot-11) 接口的QQ机器人。例如：

- [NapCat](https://github.com/NapNeko/NapCatQQ) 基于NTQQ的无头Bot框架
- [OpenShamrock](https://github.com/whitechi73/OpenShamrock) 基于 Lsposed(Non-Riru) 实现 Kritor 标准的 QQ 机器人框架
- [Lagrange](https://github.com/LagrangeDev/Lagrange.Core) 一个基于纯C#的NTQQ协议实现，源自Konata.Core
- [LiteLoaderQQNT](https://github.com/LiteLoaderQQNT/LiteLoaderQQNT) QQNT 插件加载器
- [Gensokyo](https://github.com/Hoshinonyaruko/Gensokyo) 基于qq官方api开发的符合onebot标准的golang实现，轻量、原生跨平台

> [!IMPORTANT]
> 本项目是基于onebot的正向ws接口，因此你需要开启对应机器人项目的ws监听。
>
> 本项目处理消息的格式是消息段数组，因此你需要将onobot中的`event.message_format`配置为`array`。

## 编译

```shell
go build -o YinYangJade
```

## 运行

第一次运行会生成配置文件`config.yaml`，请根据实际情况修改配置文件后重新运行。

```yml
# OneBot的ws的host
host: localhost

# OneBot的ws的port
port: 8080

# 你的机器人的QQ号
qq: 123456789

# 对应OneBot的accessToken
verifykey: ABCDEFGHIJK

# 自动退出除了以下群之外的所有群，为空则是不启用此功能
check_qq_groups:
  - 123456789
  - 987654321
```
