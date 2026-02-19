# try

A CLI tool to manage experimental project directories. Single binary, no dependencies.

> "Your experiments deserve a home."

Go rewrite of [tobi/try](https://github.com/tobi/try).

## Features

- **Centralized experiments** - All experiments in `~/src/tries` (configurable)
- **Auto-dated directories** - Creates `2024-01-15-projectname` format
- **Fuzzy search** - Interactive selector with smart scoring
- **Time-aware** - Recently accessed directories rank higher
- **Single binary** - No Ruby or other runtime required

## Installation

```bash
# From source
go install github.com/xpzouying/try@latest

# Or build locally
git clone https://github.com/xpzouying/try
cd try
go build -o try .
```

## Setup

Add to your shell config (`~/.zshrc` or `~/.bashrc`):

```bash
eval "$(try init bash)"   # or: try init zsh
```

## Tutorial

### First Run

```bash
# Create your first experiment
$ try redis-test
# Creates ~/src/tries/2024-01-15-redis-test and cd into it

# Start coding...
$ git init && echo "# Redis Test" > README.md
```

### Finding Experiments

```bash
# Open interactive selector
$ try

# Type to fuzzy search, e.g., "red" matches "redis-test"
# Use ↑/↓ to navigate, Enter to select
```

### Daily Workflow

```bash
# Quick jump to existing experiment
$ try redis      # Fuzzy matches "2024-01-15-redis-test"

# Create another experiment
$ try kafka-consumer

# Later, find it again
$ try kafka      # Jumps right in
```

### Organizing Experiments

```bash
# Your tries directory grows over time:
~/src/tries/
├── 2024-01-10-go-generics/
├── 2024-01-12-docker-compose/
├── 2024-01-15-redis-test/
└── 2024-01-15-kafka-consumer/

# Recent directories appear first in selector
# Date prefix keeps things organized chronologically
```

## Usage

```bash
try                  # Interactive selector - browse/search experiments
try redis            # Jump to "redis" experiment or create new
try clone <url>      # Clone repo into dated directory
try .                # Create worktree for current repo
```

### Keyboard Shortcuts (in selector)

| Key | Action |
|-----|--------|
| `↑/↓` or `Ctrl-P/N` | Navigate |
| `Enter` | Select directory (or create if no match) |
| `Ctrl-T` | Create new experiment with current query |
| `Esc` or `Ctrl-C` | Exit |

*Coming soon: Ctrl-D (delete), Ctrl-R (rename), Ctrl-G (graduate)*

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `TRY_PATH` | `~/src/tries` | Root directory for experiments |
| `TRY_PROJECTS` | Parent of TRY_PATH | Where graduated projects go |

## Why Go?

The original `try` is written in Ruby. This rewrite provides:

- **Single binary** - No need to install Ruby
- **Fast startup** - ~5ms vs ~100ms
- **Easy distribution** - Download and run

## License

MIT
