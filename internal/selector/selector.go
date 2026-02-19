package selector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xpzouying/try/internal/entry"
	"github.com/xpzouying/try/internal/fuzzy"
)

// Result represents the outcome of the selector.
type Result struct {
	Action string // "cd", "mkdir", "cancel"
	Path   string
}

// Run launches the interactive selector and returns the result.
func Run(initialQuery string) (*Result, error) {
	entries, err := entry.LoadEntries(entry.TriesPath())
	if err != nil {
		return nil, fmt.Errorf("load entries: %w", err)
	}

	// Open /dev/tty directly for TUI input/output.
	// This is necessary because shell wrapper captures stdout with $(...),
	// so we need to bypass stdout and write directly to the terminal.
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open tty: %w", err)
	}
	defer tty.Close()

	m := newModel(entries, initialQuery)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("run selector: %w", err)
	}

	return finalModel.(model).result, nil
}

type model struct {
	entries  []*entry.Entry // All entries
	filtered []filteredEntry
	query    string
	cursor   int
	width    int
	height   int
	result   *Result
}

type filteredEntry struct {
	entry     *entry.Entry
	positions []int // Matched positions for highlighting
}

func newModel(entries []*entry.Entry, query string) model {
	m := model{
		entries: entries,
		query:   query,
		width:   80,
		height:  24,
	}
	m.filter()
	return m
}

func (m *model) filter() {
	if m.query == "" {
		m.filtered = make([]filteredEntry, len(m.entries))
		for i, e := range m.entries {
			m.filtered[i] = filteredEntry{entry: e}
		}
		return
	}

	names := make([]string, len(m.entries))
	for i, e := range m.entries {
		names[i] = e.Name
	}

	matches := fuzzy.Search(m.query, names)
	m.filtered = make([]filteredEntry, 0, len(matches))
	for _, match := range matches {
		m.filtered = append(m.filtered, filteredEntry{
			entry:     m.entries[match.StartIndex],
			positions: match.Positions,
		})
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("try")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.result = &Result{Action: "cancel"}
		return m, tea.Quit

	case tea.KeyEnter:
		return m.selectCurrent()

	case tea.KeyUp, tea.KeyCtrlP:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyDown, tea.KeyCtrlN:
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyCtrlT:
		// Create new with current query
		return m.createNew()

	case tea.KeyBackspace:
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.filter()
		}
		return m, nil

	case tea.KeyRunes:
		m.query += string(msg.Runes)
		m.filter()
		return m, nil
	}

	return m, nil
}

func (m model) selectCurrent() (tea.Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		// No matches, create new
		return m.createNew()
	}

	selected := m.filtered[m.cursor].entry
	m.result = &Result{
		Action: "cd",
		Path:   selected.Path,
	}
	return m, tea.Quit
}

func (m model) createNew() (tea.Model, tea.Cmd) {
	if m.query == "" {
		m.result = &Result{Action: "cancel"}
		return m, tea.Quit
	}

	name := fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02"), m.query)
	path := filepath.Join(entry.TriesPath(), name)

	m.result = &Result{
		Action: "mkdir",
		Path:   path,
	}
	return m, tea.Quit
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("208"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("229"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func (m model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("try"))
	b.WriteString(" ")
	b.WriteString(promptStyle.Render("> "))
	b.WriteString(m.query)
	b.WriteString("█\n\n")

	// List
	visibleCount := m.height - 5 // Reserve for header and footer
	if visibleCount < 1 {
		visibleCount = 10
	}

	start := 0
	if m.cursor >= visibleCount {
		start = m.cursor - visibleCount + 1
	}
	end := min(start+visibleCount, len(m.filtered))

	for i := start; i < end; i++ {
		fe := m.filtered[i]
		line := m.renderEntry(fe, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Fill empty lines
	for i := len(m.filtered); i < visibleCount; i++ {
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	if len(m.filtered) == 0 && m.query != "" {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  Press Enter to create: %s-%s", time.Now().Format("2006-01-02"), m.query)))
	} else {
		b.WriteString(helpStyle.Render("  ↑/↓ navigate • Enter select • Ctrl-T new • Esc quit"))
	}

	return b.String()
}

func (m model) renderEntry(fe filteredEntry, selected bool) string {
	name := fe.entry.Name

	// Highlight matched positions
	if len(fe.positions) > 0 {
		posSet := make(map[int]bool)
		for _, p := range fe.positions {
			posSet[p] = true
		}

		var result strings.Builder
		for i, r := range name {
			s := string(r)
			if posSet[i] {
				s = matchStyle.Render(s)
			}
			result.WriteString(s)
		}
		name = result.String()
	}

	// Format age
	age := formatAge(fe.entry.ModTime)

	line := fmt.Sprintf("  %s  %s", name, dimStyle.Render(age))

	if selected {
		line = selectedStyle.Render("▸ " + name + "  " + dimStyle.Render(age))
	}

	return line
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return t.Format("2006-01-02")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// EnsureTriesDir creates the tries directory if it doesn't exist.
func EnsureTriesDir() error {
	path := entry.TriesPath()
	return os.MkdirAll(path, 0755)
}
