# try

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Your experiments deserve a home.

## The Problem

作为开发者，你一定遇到过这些场景：

- 想测试一个新库，随手创建了 `test`、`test2`、`demo-final-v2` 目录
- 上周写的 Redis 测试代码，翻遍了整个磁盘也找不到
- `/tmp` 里的实验代码被系统清理了
- 同一个问题反复踩坑，因为找不到上次的解决方案
- 项目目录越来越乱，`~/code` 里塞满了各种半成品

## The Solution

**try** 让你的所有实验代码都有一个统一的家。

```bash
try redis-test
# → 自动创建 ~/tries/2024-01-15-redis-test 并进入
```

**不用记路径** - 模糊搜索秒找：输入 `try rds` 就能匹配到 `redis-server`

**不用记时间** - 自动日期前缀：一眼看出什么时候创建的

**不用翻目录** - 最近使用的自动排前面

Go rewrite of [tobi/try](https://github.com/tobi/try). Single binary, no dependencies.

## Demo

<table>
<tr>
<td width="50%">

**Create experiment:** `try redis-test`

https://github.com/user-attachments/assets/0205df21-459a-4e82-a024-b87e1a3d9982

</td>
<td width="50%">

**Create worktree:** `try .`

https://github.com/user-attachments/assets/fdd83db5-075a-4056-b10b-2cf1ad62717f

</td>
</tr>
</table>

**Browse experiments:** `try`

<img src="https://github.com/user-attachments/assets/e47427f1-f2e7-4e97-8b57-955016ed6d21" width="600">

## Install

```bash
# Homebrew (macOS/Linux)
brew install xpzouying/tap/try

# Or via Go
go install github.com/xpzouying/try@latest
```

## Setup

Add to your shell config (`~/.zshrc`, `~/.bashrc`, or `~/.config/fish/config.fish`):

```bash
eval "$(try init zsh)"   # or bash/fish
```

## Usage

```bash
try                  # Browse all experiments with fuzzy search
try redis            # Jump to "redis" experiment or create new
try clone <url>      # Clone repo into dated directory
try .                # Create worktree for current repo
```

All experiments are stored in `~/tries/` with auto-dated names:

```
~/tries/
├── 2024-01-10-go-generics/
├── 2024-01-12-docker-compose/
└── 2024-01-15-redis-test/
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Enter` | Select or create |
| `Ctrl-T` | Create new with current query |
| `Esc` | Exit |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `TRY_PATH` | `~/tries` | Experiments directory |

## License

MIT
