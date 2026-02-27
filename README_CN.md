# try

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[English](README.md)

> 让你的实验代码有个家。

## 痛点

作为开发者，你一定遇到过这些场景：

- 想测试一个新库，随手创建了 `test`、`test2`、`demo-final-v2` 目录
- 上周写的 Redis 测试代码，翻遍了整个磁盘也找不到
- `/tmp` 里的实验代码被系统清理了
- 同一个问题反复踩坑，因为找不到上次的解决方案
- 项目目录越来越乱，`~/code` 里塞满了各种半成品

## 解决方案

**try** 让你的所有实验代码都有一个统一的家。

```bash
try redis-test
# → 自动创建 ~/tries/2024-01-15-redis-test 并进入
```

**不用记路径** - 模糊搜索秒找：输入 `try rds` 就能匹配到 `redis-server`

**不用记时间** - 自动日期前缀：一眼看出什么时候创建的

**不用翻目录** - 最近使用的自动排前面

基于 [tobi/try](https://github.com/tobi/try) 的 Go 重写版本，单文件无依赖。

## 演示

<table>
<tr>
<td width="33%">

**创建实验**<br>`try redis-test`

https://github.com/user-attachments/assets/0205df21-459a-4e82-a024-b87e1a3d9982

</td>
<td width="33%">

**创建 worktree**<br>`try .`

https://github.com/user-attachments/assets/fdd83db5-075a-4056-b10b-2cf1ad62717f

</td>
<td width="33%">

**浏览实验**<br>`try`

<img src="https://github.com/user-attachments/assets/e47427f1-f2e7-4e97-8b57-955016ed6d21">

</td>
</tr>
</table>

## 安装

```bash
# Homebrew (macOS/Linux)
brew install xpzouying/tap/try

# 或通过 Go 安装
go install github.com/xpzouying/try@latest
```

## 配置

添加到 shell 配置文件 (`~/.zshrc`、`~/.bashrc` 或 `~/.config/fish/config.fish`)：

```bash
eval "$(try init zsh)"   # 或 bash/fish
```

## 使用

```bash
try                  # 模糊搜索浏览所有实验
try redis            # 跳转到 "redis" 实验或创建新的
try clone <url>      # 克隆仓库到带日期前缀的目录
try .                # 为当前仓库创建 worktree
```

所有实验存储在 `~/tries/`，自动带日期前缀：

```
~/tries/
├── 2024-01-10-go-generics/
├── 2024-01-12-docker-compose/
└── 2024-01-15-redis-test/
```

## 用 Worktree 审查 PR

```bash
try . pr458              # 从当前仓库创建 worktree 并进入
gh pr checkout 458       # checkout PR 代码（main 分支不受影响）
# ... 看代码、跑测试，完了在 TUI 里 Ctrl-D 删除
```

## 快捷键

| 按键 | 功能 |
|------|------|
| `↑/↓` | 上下导航 |
| `Enter` | 选择或创建 |
| `Ctrl-T` | 用当前输入创建新实验 |
| `Esc` | 退出 |

## 配置项

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `TRY_PATH` | `~/tries` | 实验目录 |

## 许可证

MIT
