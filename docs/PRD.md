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
- [x] TUI reads/writes to /dev/tty (shell wrapper compatibility)
- [x] Help/version output to stderr (avoid eval issues)
- [x] Global flag check before routing (Ruby-style)

### Phase 2.5: TUI Polish ✅
- [x] Header with title and separator lines
- [x] "Search:" prompt label
- [x] Emoji icons (📁 folder, 📂 create)
- [x] Date prefix dimmed in directory names
- [x] Match score displayed alongside age
- [x] "Create new" as navigable list item
- [x] Hide create option when exact match exists
- [x] Vibrant color scheme (matching Ruby)

### Phase 3: Full Features (Current)
- [ ] `try clone <url>` - Git clone to tries directory
- [ ] `try .` / `try ./path` - Create worktree for current repo
- [ ] Auto-detect git URL and clone
- [ ] Ctrl-D delete with confirmation
- [ ] Ctrl-R rename directory
- [ ] Ctrl-G graduate to projects directory

### Phase 4: Polish
- [ ] Unit tests for fuzzy, entry, shell
- [ ] Integration tests
- [ ] Cross-platform build (Linux, macOS)
- [ ] Release automation (goreleaser)
- [ ] Homebrew formula

## Command Reference

| Command | Status | Description |
|---------|--------|-------------|
| `try` | ✅ | Open interactive selector |
| `try <query>` | ✅ | Search with initial query |
| `try init [shell]` | ✅ | Output shell wrapper function |
| `try -h` / `try help` | ✅ | Show help |
| `try version` | ✅ | Show version |
| `try clone <url>` | ❌ | Clone git repo to tries dir |
| `try .` | ❌ | Create worktree for current repo |
| `try <git-url>` | ❌ | Auto-detect and clone |

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
3. **TUI via /dev/tty** - Bypass stdout capture in shell wrapper
4. **Help/version to stderr** - Prevent accidental eval
5. **internal/ packages** - Prevent external imports
6. **Minimal abstraction** - Flat, readable code
