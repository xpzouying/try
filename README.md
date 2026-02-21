# try

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Your experiments deserve a home.

Ever find yourself with directories like `test`, `test2`, `new-test`, `actually-working-test` scattered across your filesystem? Lost experimental code to `/tmp`? Can't remember where you saved that Redis test from last month?

**try** solves this by:
- **Centralizing experiments** in `~/tries` with auto-dated names (`2024-01-15-redis-test`)
- **Fuzzy search** to quickly find any experiment (`rds` matches `redis-server`)
- **Time-aware ranking** - recently used experiments appear first

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
