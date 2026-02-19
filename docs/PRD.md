# try - Product Requirements Document

Go rewrite of [tobi/try](https://github.com/tobi/try).

## Goal

Single binary CLI tool to manage experimental project directories. No runtime dependencies.

## Development Phases

### Phase 1: Core Foundation ✅
- [x] Go module setup
- [x] CLI entry point with command routing
- [x] `try init` - Shell wrapper generation (bash/zsh/fish)
- [x] Entry package for directory scanning
- [x] Fuzzy matching algorithm
- [x] Basic scoring (recency)

### Phase 2: Interactive Selector ✅
- [x] Bubbletea TUI model
- [x] Directory listing with fuzzy search
- [x] Keyboard navigation (arrows, Ctrl-P/N)
- [x] Enter to select, Esc to cancel
- [x] Create new experiment (Ctrl-T or direct input)
- [x] Output `cd` command for shell eval

### Phase 3: Full Features
- [ ] `try clone <url>` - Git clone integration
- [ ] `try worktree` - Git worktree support
- [ ] Ctrl-D delete with confirmation
- [ ] Ctrl-R rename
- [ ] Ctrl-G graduate to projects

### Phase 4: Polish
- [ ] Unit tests for fuzzy, entry, shell
- [ ] Integration tests
- [ ] Cross-platform build
- [ ] Release automation

## Architecture

```
try/
├── main.go              # CLI entry, command routing
├── internal/
│   ├── selector/        # Bubbletea TUI
│   ├── fuzzy/           # Fuzzy matching
│   ├── entry/           # Directory entry
│   └── shell/           # Shell integration
└── docs/
    └── PRD.md           # This file
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `bubbletea` | TUI framework |
| `lipgloss` | Styling |
| `go-runewidth` | Unicode width |

## Key Design Decisions

1. **No CLI framework** - Standard `flag` is sufficient
2. **Output commands to stdout** - Shell wrapper evals them
3. **internal/ packages** - Prevent external imports
4. **Minimal abstraction** - Flat, readable code
