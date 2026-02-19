# CLAUDE.md

## Project Overview

**try** - CLI tool to manage experimental project directories. Go rewrite of [tobi/try](https://github.com/tobi/try).

See [docs/PRD.md](docs/PRD.md) for development roadmap.

## Commands

```bash
make build              # Build binary
make test               # Run tests
make install            # Install to $GOPATH/bin
./try init bash         # Output shell wrapper
./try exec              # Launch interactive selector
```

## Structure

```
main.go           # CLI entry
internal/
  selector/       # Bubbletea TUI
  fuzzy/          # Fuzzy matching
  entry/          # Directory entry
  shell/          # Shell wrapper
```

## Code Style

- No comments for obvious code
- Table-driven tests with `t.Run()`
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Early returns, avoid nesting

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `TRY_PATH` | `~/src/tries` | Experiments directory |
| `TRY_PROJECTS` | Parent of TRY_PATH | Graduate destination |
