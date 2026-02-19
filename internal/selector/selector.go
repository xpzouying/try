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
	Action   string // "cd", "mkdir", "graduate", "delete", "cancel"
	Path     string
	DestPath string // For graduate: destination path
	BaseName string // For graduate/delete: original directory name
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

// UI mode
type mode int

const (
	modeList mode = iota
	modeGraduate
	modeDelete
)

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

	// Dialog mode
	mode         mode
	dialogInput  string // Input buffer for dialog
	dialogCursor int    // Cursor position in dialog input
	dialogError  string // Error message to display
	dialogEntry  *entry.Entry // Entry being operated on
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

	// Check if exact name already exists (don't show create option if so)
	newName := fmt.Sprintf("%s-%s", m.now.Format("2006-01-02"), m.query)
	m.showCreate = true
	for _, fe := range m.filtered {
		if fe.entry.Name == newName {
			m.showCreate = false
			break
		}
	}

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
		switch m.mode {
		case modeGraduate:
			return m.handleGraduateKey(msg)
		case modeDelete:
			return m.handleDeleteKey(msg)
		default:
			return m.handleKey(msg)
		}
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

	case tea.KeyCtrlG:
		return m.enterGraduateMode()

	case tea.KeyCtrlD:
		return m.enterDeleteMode()

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

func (m model) enterGraduateMode() (tea.Model, tea.Cmd) {
	// Can only graduate an existing directory
	if m.isCreateSelected() || len(m.filtered) == 0 {
		return m, nil
	}

	selected := m.filtered[m.cursor].entry

	// Default destination: projects_dir/basename
	destPath := filepath.Join(entry.ProjectsPath(), selected.BaseName)

	m.mode = modeGraduate
	m.dialogEntry = selected
	m.dialogInput = destPath
	m.dialogCursor = len(destPath)
	m.dialogError = ""

	return m, nil
}

func (m model) handleGraduateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		// Cancel graduate mode
		m.mode = modeList
		m.dialogError = ""
		return m, nil

	case tea.KeyEnter:
		// Confirm graduate
		return m.confirmGraduate()

	case tea.KeyBackspace:
		if m.dialogCursor > 0 {
			m.dialogInput = m.dialogInput[:m.dialogCursor-1] + m.dialogInput[m.dialogCursor:]
			m.dialogCursor--
			m.dialogError = ""
		}
		return m, nil

	case tea.KeyLeft, tea.KeyCtrlB:
		if m.dialogCursor > 0 {
			m.dialogCursor--
		}
		return m, nil

	case tea.KeyRight, tea.KeyCtrlF:
		if m.dialogCursor < len(m.dialogInput) {
			m.dialogCursor++
		}
		return m, nil

	case tea.KeyCtrlA:
		m.dialogCursor = 0
		return m, nil

	case tea.KeyCtrlE:
		m.dialogCursor = len(m.dialogInput)
		return m, nil

	case tea.KeyCtrlK:
		m.dialogInput = m.dialogInput[:m.dialogCursor]
		m.dialogError = ""
		return m, nil

	case tea.KeyCtrlW:
		// Delete word backward
		if m.dialogCursor > 0 {
			newPos := wordBoundaryBackward(m.dialogInput, m.dialogCursor)
			m.dialogInput = m.dialogInput[:newPos] + m.dialogInput[m.dialogCursor:]
			m.dialogCursor = newPos
			m.dialogError = ""
		}
		return m, nil

	case tea.KeyRunes:
		ch := string(msg.Runes)
		m.dialogInput = m.dialogInput[:m.dialogCursor] + ch + m.dialogInput[m.dialogCursor:]
		m.dialogCursor += len(ch)
		m.dialogError = ""
		return m, nil
	}

	return m, nil
}

func (m model) confirmGraduate() (tea.Model, tea.Cmd) {
	dest := strings.TrimSpace(m.dialogInput)

	if dest == "" {
		m.dialogError = "Destination cannot be empty"
		return m, nil
	}

	// Expand home directory
	if strings.HasPrefix(dest, "~") {
		home, _ := os.UserHomeDir()
		dest = filepath.Join(home, dest[1:])
	}

	// Check if destination already exists
	if _, err := os.Stat(dest); err == nil {
		m.dialogError = "Destination already exists"
		return m, nil
	}

	// Check if parent directory exists
	parent := filepath.Dir(dest)
	if _, err := os.Stat(parent); os.IsNotExist(err) {
		m.dialogError = "Parent directory does not exist"
		return m, nil
	}

	m.result = &Result{
		Action:   "graduate",
		Path:     m.dialogEntry.Path,
		DestPath: dest,
		BaseName: m.dialogEntry.Name,
	}
	return m, tea.Quit
}

func wordBoundaryBackward(s string, pos int) int {
	if pos <= 0 {
		return 0
	}
	// Skip trailing spaces
	i := pos - 1
	for i > 0 && s[i] == ' ' {
		i--
	}
	// Skip non-spaces (word characters)
	for i > 0 && s[i-1] != ' ' && s[i-1] != '/' {
		i--
	}
	return i
}

func (m model) enterDeleteMode() (tea.Model, tea.Cmd) {
	// Can only delete an existing directory
	if m.isCreateSelected() || len(m.filtered) == 0 {
		return m, nil
	}

	selected := m.filtered[m.cursor].entry

	m.mode = modeDelete
	m.dialogEntry = selected
	m.dialogInput = ""
	m.dialogCursor = 0
	m.dialogError = ""

	return m, nil
}

func (m model) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		// Cancel delete mode
		m.mode = modeList
		m.dialogError = ""
		return m, nil

	case tea.KeyEnter:
		// Confirm delete
		return m.confirmDelete()

	case tea.KeyBackspace:
		if m.dialogCursor > 0 {
			m.dialogInput = m.dialogInput[:m.dialogCursor-1] + m.dialogInput[m.dialogCursor:]
			m.dialogCursor--
			m.dialogError = ""
		}
		return m, nil

	case tea.KeyRunes:
		ch := string(msg.Runes)
		m.dialogInput = m.dialogInput[:m.dialogCursor] + ch + m.dialogInput[m.dialogCursor:]
		m.dialogCursor += len(ch)
		m.dialogError = ""
		return m, nil
	}

	return m, nil
}

func (m model) confirmDelete() (tea.Model, tea.Cmd) {
	if strings.ToUpper(strings.TrimSpace(m.dialogInput)) != "YES" {
		m.dialogError = "Type YES to confirm deletion"
		return m, nil
	}

	m.result = &Result{
		Action:   "delete",
		Path:     m.dialogEntry.Path,
		BaseName: m.dialogEntry.Name,
	}
	return m, tea.Quit
}

// Styles - using vibrant colors similar to Ruby version
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("114")) // Green (like Ruby HEADER)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("238"))

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange (like Ruby ACCENT)
			Bold(true)

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")) // Yellow folder

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")) // Muted (like Ruby MUTED)

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")). // Bright yellow (like Ruby MATCH)
			Bold(true)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	createStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")). // Green
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	graduateStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("208")) // Orange for graduate

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red for errors

	deleteStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")) // Red for delete
)

func (m model) View() string {
	switch m.mode {
	case modeGraduate:
		return m.viewGraduateDialog()
	case modeDelete:
		return m.viewDeleteDialog()
	}

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
	b.WriteString(helpStyle.Render("↑/↓ Navigate  Enter Select  Ctrl-T New  Ctrl-G Graduate  Ctrl-D Delete  Esc Cancel"))

	return b.String()
}

func (m model) viewGraduateDialog() string {
	var b strings.Builder

	// Header
	b.WriteString("  ")
	b.WriteString(graduateStyle.Render("🚀 Graduate"))
	b.WriteString(titleStyle.Render(" - Promote to Project"))
	b.WriteString("\n")

	// Separator
	b.WriteString(m.separator())
	b.WriteString("\n\n")

	// Source directory
	b.WriteString("  ")
	b.WriteString(folderStyle.Render("📁 "))
	b.WriteString(nameStyle.Render(m.dialogEntry.Name))
	b.WriteString("\n\n")

	// Destination hint
	envHint := "$TRY_PROJECTS"
	if os.Getenv("TRY_PROJECTS") == "" {
		envHint = "parent of $TRY_PATH"
	}
	projectsDir := entry.ProjectsPath()
	b.WriteString("  ")
	b.WriteString(metaStyle.Render(fmt.Sprintf("Destination (%s: %s)", envHint, projectsDir)))
	b.WriteString("\n\n")

	// Input field
	b.WriteString("  ")
	b.WriteString(promptStyle.Render("Move to: "))
	// Render input with cursor
	if m.dialogCursor >= len(m.dialogInput) {
		b.WriteString(inputStyle.Render(m.dialogInput))
		b.WriteString(cursorStyle.Render("█"))
	} else {
		b.WriteString(inputStyle.Render(m.dialogInput[:m.dialogCursor]))
		b.WriteString(cursorStyle.Render(string(m.dialogInput[m.dialogCursor])))
		b.WriteString(inputStyle.Render(m.dialogInput[m.dialogCursor+1:]))
	}
	b.WriteString("\n\n")

	// Symlink hint
	b.WriteString("  ")
	b.WriteString(metaStyle.Render("A symlink will be left in the tries directory"))
	b.WriteString("\n")

	// Error message
	if m.dialogError != "" {
		b.WriteString("\n  ")
		b.WriteString(errorStyle.Render("⚠ " + m.dialogError))
		b.WriteString("\n")
	}

	// Separator
	b.WriteString("\n")
	b.WriteString(m.separator())
	b.WriteString("\n")

	// Footer
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("Enter Confirm  Esc Cancel"))

	return b.String()
}

func (m model) viewDeleteDialog() string {
	var b strings.Builder

	// Header
	b.WriteString("  ")
	b.WriteString(deleteStyle.Render("🗑️  Delete"))
	b.WriteString(titleStyle.Render(" - Remove Directory"))
	b.WriteString("\n")

	// Separator
	b.WriteString(m.separator())
	b.WriteString("\n\n")

	// Directory to delete
	b.WriteString("  ")
	b.WriteString(deleteStyle.Render("📁 "))
	b.WriteString(nameStyle.Render(m.dialogEntry.Name))
	b.WriteString("\n\n")

	// Warning
	b.WriteString("  ")
	b.WriteString(errorStyle.Render("⚠ This will permanently delete the directory and all its contents!"))
	b.WriteString("\n\n")

	// Input field
	b.WriteString("  ")
	b.WriteString(promptStyle.Render("Type YES to confirm: "))
	b.WriteString(inputStyle.Render(m.dialogInput))
	b.WriteString(cursorStyle.Render("█"))
	b.WriteString("\n")

	// Error message
	if m.dialogError != "" {
		b.WriteString("\n  ")
		b.WriteString(errorStyle.Render("⚠ " + m.dialogError))
		b.WriteString("\n")
	}

	// Separator
	b.WriteString("\n")
	b.WriteString(m.separator())
	b.WriteString("\n")

	// Footer
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("Enter Confirm  Esc Cancel"))

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
