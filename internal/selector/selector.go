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
	entries      []*entry.Entry
	filtered     []filteredEntry
	query        string
	cursor       int
	width        int
	height       int
	result       *Result
	now          time.Time
	showCreate   bool // Whether to show "Create new" option
}

type filteredEntry struct {
	entry     *entry.Entry
	score     float64
	positions []int
}

func newModel(entries []*entry.Entry, query string) model {
	m := model{
		entries:    entries,
		query:      query,
		width:      80,
		height:     24,
		now:        time.Now(),
		showCreate: true,
	}
	m.filter()
	return m
}

func (m *model) filter() {
	if m.query == "" {
		m.filtered = make([]filteredEntry, len(m.entries))
		for i, e := range m.entries {
			m.filtered[i] = filteredEntry{
				entry: e,
				score: e.Score(m.now),
			}
		}
		m.showCreate = false
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
			score:     match.Score,
			positions: match.Positions,
		})
	}

	m.showCreate = true

	// Ensure cursor is in valid range
	maxCursor := len(m.filtered)
	if m.showCreate {
		maxCursor++
	}
	if m.cursor >= maxCursor {
		m.cursor = max(0, maxCursor-1)
	}
}

func (m model) totalItems() int {
	count := len(m.filtered)
	if m.showCreate {
		count++
	}
	return count
}

func (m model) isCreateSelected() bool {
	return m.showCreate && m.cursor == len(m.filtered)
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
		maxCursor := m.totalItems() - 1
		if m.cursor < maxCursor {
			m.cursor++
		}
		return m, nil

	case tea.KeyCtrlT:
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
	if m.isCreateSelected() {
		return m.createNew()
	}

	if len(m.filtered) == 0 {
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

	name := fmt.Sprintf("%s-%s", m.now.Format("2006-01-02"), m.query)
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
			Foreground(lipgloss.Color("212"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236"))

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	createStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))
)

func (m model) View() string {
	var b strings.Builder

	// Header
	b.WriteString("  ")
	b.WriteString(titleStyle.Render("🏠 Try"))
	b.WriteString(titleStyle.Render(" - Experiment Directory"))
	b.WriteString("\n")

	// Separator
	b.WriteString(m.separator())
	b.WriteString("\n")

	// Search input
	b.WriteString("  ")
	b.WriteString(promptStyle.Render("Search: "))
	b.WriteString(inputStyle.Render(m.query))
	b.WriteString(cursorStyle.Render("█"))
	b.WriteString("\n")

	// Separator
	b.WriteString(m.separator())
	b.WriteString("\n")

	// List
	visibleCount := m.height - 8 // Reserve for header, separators, footer
	if visibleCount < 3 {
		visibleCount = 3
	}

	totalItems := m.totalItems()
	start := 0
	if m.cursor >= visibleCount {
		start = m.cursor - visibleCount + 1
	}
	end := min(start+visibleCount, totalItems)

	// Render directory entries
	for i := start; i < end; i++ {
		if i < len(m.filtered) {
			b.WriteString(m.renderEntry(i, i == m.cursor))
		} else if m.showCreate {
			b.WriteString(m.renderCreateOption(i == m.cursor))
		}
		b.WriteString("\n")
	}

	// Fill empty lines
	for i := totalItems; i < visibleCount && i < visibleCount; i++ {
		b.WriteString("\n")
	}

	// Separator
	b.WriteString(m.separator())
	b.WriteString("\n")

	// Footer
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("↑/↓ Navigate  Enter Select  Ctrl-T New  Esc Cancel"))

	return b.String()
}

func (m model) separator() string {
	width := m.width - 2
	if width < 10 {
		width = 78
	}
	return "  " + separatorStyle.Render(strings.Repeat("─", width))
}

func (m model) renderEntry(idx int, selected bool) string {
	fe := m.filtered[idx]
	var line strings.Builder

	// Selection indicator
	if selected {
		line.WriteString(arrowStyle.Render("→ "))
	} else {
		line.WriteString("  ")
	}

	// Folder emoji
	line.WriteString(folderStyle.Render("📁 "))

	// Directory name with date dimmed
	name := fe.entry.Name
	if fe.entry.HasDate && len(name) > 11 {
		datePart := name[:11]  // "2024-01-15-"
		namePart := name[11:]  // rest

		// Render date part dimmed
		line.WriteString(dateStyle.Render(datePart))

		// Render name part with highlights
		if len(fe.positions) > 0 {
			line.WriteString(m.highlightName(namePart, fe.positions, 11))
		} else {
			line.WriteString(nameStyle.Render(namePart))
		}
	} else {
		// No date prefix, render with highlights
		if len(fe.positions) > 0 {
			line.WriteString(m.highlightName(name, fe.positions, 0))
		} else {
			line.WriteString(nameStyle.Render(name))
		}
	}

	// Metadata (right side)
	age := formatAge(fe.entry.ModTime)
	scoreStr := fmt.Sprintf("%.1f", fe.score)
	meta := fmt.Sprintf("  %s, %s", age, scoreStr)
	line.WriteString(metaStyle.Render(meta))

	// Apply selected background
	result := line.String()
	if selected {
		result = selectedStyle.Render(result)
	}

	return result
}

func (m model) highlightName(name string, positions []int, offset int) string {
	posSet := make(map[int]bool)
	for _, p := range positions {
		adjusted := p - offset
		if adjusted >= 0 {
			posSet[adjusted] = true
		}
	}

	var result strings.Builder
	for i, r := range name {
		s := string(r)
		if posSet[i] {
			result.WriteString(matchStyle.Render(s))
		} else {
			result.WriteString(nameStyle.Render(s))
		}
	}
	return result.String()
}

func (m model) renderCreateOption(selected bool) string {
	var line strings.Builder

	if selected {
		line.WriteString(arrowStyle.Render("→ "))
	} else {
		line.WriteString("  ")
	}

	line.WriteString(createStyle.Render("📂 "))

	datePrefix := m.now.Format("2006-01-02")
	if m.query == "" {
		line.WriteString(createStyle.Render(fmt.Sprintf("Create new: %s-", datePrefix)))
	} else {
		line.WriteString(createStyle.Render(fmt.Sprintf("Create new: %s-%s", datePrefix, m.query)))
	}

	result := line.String()
	if selected {
		result = selectedStyle.Render(result)
	}

	return result
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	default:
		return t.Format("Jan 02")
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
