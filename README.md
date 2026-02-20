# try

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Your experiments deserve a home.

A CLI tool to manage experimental project directories. Single binary, no dependencies.

Go rewrite of [tobi/try](https://github.com/tobi/try).

<!-- TODO: Add demo GIF here -->
<!-- ![Demo](docs/demo.gif) -->

## Install

```bash
# Homebrew (macOS/Linux)
brew install xpzouying/tap/try

# Or via Go
go install github.com/xpzouying/try@latest
```

## Setup

Add to your shell config:

```bash
# zsh
echo 'eval "$(try init zsh)"' >> ~/.zshrc

# bash
echo 'eval "$(try init bash)"' >> ~/.bashrc

# fish
echo 'try init fish | source' >> ~/.config/fish/config.fish
```

Reload and verify:

```bash
source ~/.zshrc      # or restart terminal
type try             # Should show: try is a shell function
```

## Features

- **Centralized experiments** - All experiments in `~/tries` (configurable)
- **Auto-dated directories** - Creates `2024-01-15-projectname` format
- **Fuzzy search** - Interactive selector with smart scoring
- **Time-aware** - Recently accessed directories rank higher
- **Single binary** - No Ruby or other runtime required

## Usage

```bash
try                  # Interactive selector - browse/search experiments
try redis            # Jump to "redis" experiment or create new
try clone <url>      # Clone repo into dated directory
try .                # Create worktree for current repo
```

### Examples

```bash
# Create your first experiment
$ try redis-test
# Creates ~/tries/2024-01-15-redis-test and cd into it

# Later, find it with fuzzy search
$ try redis          # Fuzzy matches "2024-01-15-redis-test"

# Or browse all experiments
$ try
# Type to search, ↑/↓ to navigate, Enter to select
```

### Your tries directory

```
~/tries/
├── 2024-01-10-go-generics/
├── 2024-01-12-docker-compose/
├── 2024-01-15-redis-test/
└── 2024-01-15-kafka-consumer/
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` or `Ctrl-P/N` | Navigate |
| `Enter` | Select directory (or create if no match) |
| `Ctrl-T` | Create new experiment with current query |
| `Esc` or `Ctrl-C` | Exit |

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `TRY_PATH` | `~/tries` | Root directory for experiments |
| `TRY_PROJECTS` | Parent of TRY_PATH | Where graduated projects go |

## Why Go?

The original `try` is written in Ruby. This rewrite provides:

- **Single binary** - No need to install Ruby
- **Fast startup** - ~5ms vs ~100ms
- **Easy distribution** - Download and run

## License

MIT
