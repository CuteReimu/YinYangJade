# Copilot 代码助手使用说明

**重要提示**: 本项目的维护者主要是简体中文使用者。在新建Issue、新建PR、review PR、进行回复等所有交互时，请**全部使用简体中文**。

## 项目概述

阴阳神玉（YinYangJade）是一个基于 [onebot-11](https://github.com/botuniverse/onebot-11) 接口的QQ机器人项目合集，使用 Go 语言开发。项目包含以下四个独立的机器人模块：

1. **tfcc** - 东方Project沙包聚集地机器人
2. **hkbot** - 空洞骑士speedrun推送小助手  
3. **maplebot** - GMSR群机器人（冒险岛相关）
4. **fengsheng** - 风声群机器人

## 仓库信息

- **代码规模**: 约45个Go文件，总计约8700行代码
- **主要语言**: Go
- **Go版本要求**: 1.24.0+ （go.mod中定义）
- **项目类型**: QQ机器人应用，基于WebSocket通信
- **依赖管理**: Go Modules

## 目录结构

```
/home/runner/work/YinYangJade/YinYangJade/
├── .github/
│   ├── workflows/
│   │   ├── golangci-lint.yml  # 代码质量检查和构建
│   │   └── gofmt.yml          # 自动格式化代码
│   └── dependabot.yml         # 依赖更新配置
├── db/                         # 数据库封装（使用BadgerDB）
├── fengsheng/                  # 风声机器人模块
├── hkbot/                      # 空洞骑士机器人模块
├── iface/                      # 通用接口定义
├── imageutil/                  # 图像处理工具
├── maplebot/                   # 冒险岛机器人模块
│   └── scripts/               # python脚本
├── slicegame/                  # 滑块游戏，一个小算法，与主功能无关
├── tfcc/                       # 东方Project机器人模块
├── main.go                     # 程序入口
├── go.mod                      # Go模块依赖
├── go.sum                      # 依赖校验和
├── .golangci.yml              # golangci-lint配置
├── .gitignore                 # Git忽略文件配置
└── README.md                   # 项目说明文档
```

## 核心架构

### 主程序入口 (main.go)
- 初始化数据库（BadgerDB，存储在 `assets/database`）
- 读取主配置文件 `config.yaml`
- 连接OneBot WebSocket服务器
- 初始化四个子机器人模块
- 处理好友请求和群组请求
- 定时任务调度（使用cron）

### 配置文件结构
- **主配置**: `config.yaml`（根目录）
- **子模块配置目录**: 
  - `config/org.tfcc.bot/TFCCConfig.yml`
  - `config/net.cutereimu.hkbot/HKConfig.yml`
  - `config/net.cutereimu.maplebots/Config.yml`
  - `config/com.fengsheng.bot/FengshengConfig.yml`
- **数据目录**: 各模块在 `data/` 下存储运行时数据
- **日志目录**: `logs/` 保存7天的日志文件

### 关键依赖
- `github.com/CuteReimu/onebot` - OneBot协议实现
- `github.com/dgraph-io/badger/v4` - 嵌入式KV数据库
- `github.com/spf13/viper` - 配置管理
- `github.com/go-rod/rod` - 浏览器自动化（fengsheng模块使用）
- `github.com/CuteReimu/bilibili/v2` - B站API封装
- `github.com/vicanso/go-charts/v2` - 图表生成

## 构建和验证流程

### 环境准备

**始终**在进行任何构建或测试操作前，确保Go环境正确：

```bash
go version  # 应该显示 go1.24.x 或更高版本
```

### 构建步骤

**重要**: 构建命令应该按以下顺序执行：

1. **下载依赖**（如果是首次构建或go.mod有变化）：
   ```bash
   go mod download
   ```

2. **构建所有包**（验证代码可编译）：
   ```bash
   go build -v ./...
   ```
   - 耗时：首次约30-60秒（需下载依赖），后续约5-10秒
   - 输出：会显示编译的所有包

3. **生成可执行文件**：
   ```bash
   go build -o YinYangJade .
   ```
   - 生成约33MB的二进制文件
   - 注意：此文件已加入.gitignore，不应提交到仓库

### 代码质量检查

**必须**在提交代码前运行以下检查：

1. **代码格式检查**：
   ```bash
   gofmt -l .
   ```
   - 如果有输出，说明有文件格式不符合标准
   - 运行 `gofmt -s -l -w .` 自动修复格式问题
   - 注意：master分支推送时，gofmt工作流会自动修复并提交

2. **Linting**（最重要的检查）：
   ```bash
   golangci-lint run --timeout 3m
   ```
   - **必须**通过此检查才能合并PR
   - 超时时间设置为3分钟（与CI配置一致）
   - 配置文件：`.golangci.yml`
   - 耗时：首次约30-60秒，后续约10-20秒

### 测试

**注意**: 本项目当前**没有测试文件**。运行 `go test ./...` 会显示所有包都没有测试文件，这是正常现象。

### CI/CD流程

#### golangci-lint工作流（PR和master推送触发）
1. 检出代码
2. 设置Go环境（使用stable版本）
3. 运行golangci-lint（latest版本）
4. 构建项目 `go build -v ./...`

#### gofmt工作流（仅master推送触发）
1. 检出代码
2. 运行gofmt检查和自动修复
3. 如果有格式变更，自动提交并推送

**关键点**：
- golangci-lint检查**必须**通过，否则PR会被拒绝
- 格式问题会在master分支自动修复，但最好在PR前就修复

## 常见构建问题和解决方案

### 问题1: go.mod exclude导致的依赖问题
**现象**: 某些依赖版本无法下载  
**原因**: go.mod中有exclude配置，排除了某些breaking change版本：
```go.mod
exclude (
	github.com/ysmood/fetchup v0.4.0
	github.com/ysmood/fetchup v0.5.0
	github.com/ysmood/fetchup v0.5.1
	github.com/ysmood/fetchup v0.5.2
)
```
**解决**: 不要修改这些exclude配置，它们是有意为之

### 问题2: 初次运行程序退出
**现象**: 运行 `./YinYangJade` 后立即退出，提示修改config.yaml  
**原因**: 这是正常行为，程序会生成默认配置文件  
**解决**: 
1. 修改生成的 `config.yaml` 配置文件
2. 重新运行程序

### 问题3: 数据库相关错误
**现象**: 启动时出现database相关错误  
**解决**: 
- 确保 `assets/database` 目录存在且可写
- BadgerDB会自动创建和管理数据库文件
- 如需清理，删除 `assets/` 目录即可重新初始化

### 问题4: 中文字体问题（仅Linux）
**现象**: maplebot的"洗魔方"功能显示中文乱码  
**解决**: 
```bash
# 将黑体文件 simhei.ttf 放入 /usr/share/fonts
sudo cp simhei.ttf /usr/share/fonts/
# 刷新字体缓存
fc-cache
# 验证
fc-list :lang=zh | grep 黑体
```

## 代码变更指南

### 新增功能时的注意事项

1. **确定目标模块**: 根据功能属性选择合适的子模块目录（tfcc/hkbot/maplebot/fengsheng）

2. **实现CmdHandler接口**（如果是新增命令）:
   - 位置：`iface/iface.go`
   - 必须实现：Name(), ShowTips(), CheckAuth(), Execute()

3. **配置管理**: 
   - 使用viper管理配置
   - 配置文件放在 `config/<module-name>/`
   - 运行时数据放在 `data/<module-name>/`

4. **数据库访问**:
   - 通过 `db.DB` 访问BadgerDB实例
   - 参考 `db/map.go` 中的使用示例

5. **日志记录**:
   - 使用 `log/slog` 包
   - 日志会自动写入 `logs/log-YYYY-MM-DD.log`
   - 保留7天

### 修改代码时的最佳实践

1. **始终**先运行 `go build -v ./...` 确保代码可编译
2. **始终**在提交前运行 `golangci-lint run --timeout 3m`
3. **始终**使用 `gofmt` 格式化代码
4. **不要**修改go.mod中的exclude配置
5. **不要**提交 `.yml` 配置文件（除了.golangci.yml）
6. **不要**提交 `assets/`、`logs/`、`data/` 目录
7. **不要**提交编译后的二进制文件

### 依赖更新

- dependabot会自动检查go依赖更新（每天运行）
- 手动更新: `go get -u ./...` 然后 `go mod tidy`
- 更新后**必须**运行完整的构建和lint检查

## 验证变更的完整流程

在提交PR之前，按此顺序执行：

```bash
# 1. 格式化代码
gofmt -s -l -w .

# 2. 整理依赖（如果修改了import）
go mod tidy

# 3. 构建检查
go build -v ./...

# 4. Lint检查（最重要）
golangci-lint run --timeout 3m

# 5. 生成二进制验证（可选）
go build -o YinYangJade .

# 6. 清理构建产物
rm -f YinYangJade
```

如果以上所有步骤都通过，代码就可以提交了。

## 项目特殊说明

### OneBot协议配置要求
本项目使用OneBot的**正向WebSocket**接口，配置时注意：
- 必须开启OneBot的ws监听
- `event.message_format` 必须配置为 `array`（消息段数组格式）

### 配置文件模式
- 首次运行会生成默认配置文件
- 所有 `.yml` 文件都在 `.gitignore` 中（除了.golangci.yml）
- 修改配置后需要重启程序

### 模块独立性
四个机器人模块相对独立，可以根据需要启用或禁用特定功能。每个模块都有自己的：
- 配置文件目录
- 数据存储目录  
- README说明文档

## 关键文件列表

**根目录文件**:
- `main.go` - 程序入口，约212行
- `go.mod` - 依赖定义
- `.golangci.yml` - Lint配置
- `.gitignore` - Git忽略配置
- `README.md` - 项目说明

**配置和数据** (运行时生成，不提交):
- `config.yaml` - 主配置
- `config/` - 各模块配置
- `data/` - 运行时数据
- `logs/` - 日志文件
- `assets/` - 资源和数据库

## 工作流程总结

**信任这份说明**: 以上所有步骤都已经过验证。只有在说明不完整或发现错误时才需要进行额外搜索。遵循此说明可以避免90%以上的常见问题和构建失败。
