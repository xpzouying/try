# try

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Your experiments deserve a home. Your brain doesn't work in neat folders.

A CLI tool to manage experimental project directories. Go rewrite of [tobi/try](https://github.com/tobi/try).

## Demo

**Create a new experiment:**

https://github.com/user-attachments/assets/0205df21-459a-4e82-a024-b87e1a3d9982

**Create a worktree for current repo:**

https://github.com/user-attachments/assets/fdd83db5-075a-4056-b10b-2cf1ad62717f

**Browse all experiments:**

![try_history_01](https://github.com/user-attachments/assets/e47427f1-f2e7-4e97-8b57-955016ed6d21)

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
