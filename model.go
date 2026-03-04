package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	lipgloss "charm.land/lipgloss/v2"
)

// Messages

type scanResultMsg struct {
	panes []PaneInfo
}

type tickMsg time.Time

// Model

type model struct {
	panes      []PaneInfo
	cursor     int
	viewport   viewport.Model
	width      int
	height     int
	jumpTarget string // pane ID to switch to on exit
}

func initialModel() model {
	return model{
		viewport: viewport.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(doScan(), doTick())
}

func doScan() tea.Cmd {
	return func() tea.Msg {
		return scanResultMsg{panes: scanAllPanes()}
	}
}

func doTick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.updatePreview()
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.updatePreview()
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.panes)-1 {
				m.cursor++
				m.updatePreview()
			}
			return m, nil

		case key.Matches(msg, keys.Enter):
			if len(m.panes) > 0 && m.cursor < len(m.panes) {
				m.jumpTarget = m.panes[m.cursor].PaneID
				return m, tea.Quit
			}
			return m, nil

		case key.Matches(msg, keys.Refresh):
			return m, doScan()
		}

	case scanResultMsg:
		// Preserve cursor by tracking PaneID
		var selectedID string
		if len(m.panes) > 0 && m.cursor < len(m.panes) {
			selectedID = m.panes[m.cursor].PaneID
		}
		m.panes = msg.panes

		// Restore cursor position
		found := false
		if selectedID != "" {
			for i, p := range m.panes {
				if p.PaneID == selectedID {
					m.cursor = i
					found = true
					break
				}
			}
		}
		if !found && m.cursor >= len(m.panes) {
			m.cursor = max(0, len(m.panes)-1)
		}
		m.updatePreview()
		return m, nil

	case tickMsg:
		return m, tea.Batch(doScan(), doTick())
	}

	return m, nil
}

func (m *model) updateLayout() {
	leftWidth := m.leftPanelWidth()
	rightWidth := m.width - leftWidth - 1 // 1 for separator
	if rightWidth < 10 {
		rightWidth = 10
	}
	contentHeight := m.contentHeight()
	m.viewport.SetWidth(rightWidth)
	m.viewport.SetHeight(contentHeight)
}

func (m *model) updatePreview() {
	if len(m.panes) > 0 && m.cursor < len(m.panes) {
		content := capturePane(m.panes[m.cursor].PaneID)
		m.viewport.SetContent(content)
		m.viewport.GotoBottom()
	} else {
		m.viewport.SetContent("")
	}
}

func (m model) leftPanelWidth() int {
	w := m.width * 40 / 100
	if w < 30 {
		w = 30
	}
	if w > m.width-15 {
		w = m.width - 15
	}
	return w
}

func (m model) contentHeight() int {
	h := m.height - 2 // header + status bar
	if h < 1 {
		h = 1
	}
	return h
}

func (m model) View() tea.View {
	if m.width == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("── cc-watch  %d pane(s) waiting ──", len(m.panes))
	b.WriteString(styleDim.Render(header))
	b.WriteString("\n")

	contentHeight := m.contentHeight()

	if len(m.panes) == 0 {
		// No waiting panes
		noWait := lipgloss.NewStyle().
			Foreground(colorGreen).
			Width(m.width).
			Height(contentHeight).
			Render("✓ 待機なし")
		b.WriteString(noWait)
		b.WriteString("\n")
	} else {
		leftWidth := m.leftPanelWidth()

		// Build left panel lines
		var leftLines []string
		for i, p := range m.panes {
			info := waitInfo[p.WaitType]
			style := waitStyle(p.WaitType)

			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}

			num := styleDim.Render(fmt.Sprintf("(%d)", i+1))
			icon := style.Render(info.Icon)
			label := style.Width(10).Render(info.Label)

			line := fmt.Sprintf("%s%s %s %s  %s", cursor, num, icon, label, p.Name)
			leftLines = append(leftLines, line)
		}

		leftPanel := lipgloss.NewStyle().
			Width(leftWidth).
			Height(contentHeight).
			Render(strings.Join(leftLines, "\n"))

		// Separator
		var sepLines []string
		for i := 0; i < contentHeight; i++ {
			sepLines = append(sepLines, "│")
		}
		sep := styleDim.Render(strings.Join(sepLines, "\n"))

		// Right panel (viewport preview)
		rightPanel := m.viewport.View()

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, sep, rightPanel))
		b.WriteString("\n")
	}

	// Status bar
	timeStr := time.Now().Format("15:04:05")
	status := fmt.Sprintf("%s  ↑↓ 移動  enter ジャンプ  q 終了  r 更新", timeStr)
	b.WriteString(styleStatusBar.Render(status))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
