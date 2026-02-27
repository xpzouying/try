# try

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[中文文档](README_CN.md)

> Your experiments deserve a home.

## The Problem

As a developer, you've probably experienced:

- Creating `test`, `test2`, `demo-final-v2` directories when trying out a new library
- Searching everywhere for that Redis test code you wrote last week
- Losing experimental code when `/tmp` gets cleaned up
- Hitting the same problem twice because you can't find your previous solution
- A messy `~/code` directory full of abandoned experiments

## The Solution

**try** gives all your experimental code a home.

```bash
try redis-test
# → Creates ~/tries/2024-01-15-redis-test and cd into it
```

**No paths to remember** - Fuzzy search finds it: `try rds` matches `redis-server`

**No dates to remember** - Auto date-prefix: instantly see when you created it

**No digging through folders** - Recently used experiments appear first

Go rewrite of [tobi/try](https://github.com/tobi/try). Single binary, no dependencies.

## Demo

<table>
<tr>
<td width="33%">

**Create experiment**<br>`try redis-test`

https://github.com/user-attachments/assets/0205df21-459a-4e82-a024-b87e1a3d9982

</td>
<td width="33%">

**Create worktree**<br>`try .`

https://github.com/user-attachments/assets/fdd83db5-075a-4056-b10b-2cf1ad62717f

</td>
<td width="33%">

**Browse experiments**<br>`try`

<img src="https://github.com/user-attachments/assets/e47427f1-f2e7-4e97-8b57-955016ed6d21">

</td>
</tr>
</table>

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

## Review a PR with Worktree

```bash
try . pr458              # Create worktree from current repo, cd into it
gh pr checkout 458       # Checkout PR code (main branch stays untouched)
# ... review, run tests, done? Ctrl-D delete in TUI
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
