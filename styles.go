package main

import lipgloss "charm.land/lipgloss/v2"

// Hex color values, shared by the lipgloss styles (TUI) and the tmux status
// line markup (cc-watch status). Single source of truth.
const (
	hexRed     = "#e06c75"
	hexYellow  = "#e5c07b"
	hexGreen   = "#98c379"
	hexCyan    = "#56b6c2"
	hexDim     = "#5c6370"
	hexBlue    = "#61afef"
	hexMagenta = "#c678dd"
)

var (
	colorRed     = lipgloss.Color(hexRed)
	colorYellow  = lipgloss.Color(hexYellow)
	colorGreen   = lipgloss.Color(hexGreen)
	colorCyan    = lipgloss.Color(hexCyan)
	colorDim     = lipgloss.Color(hexDim)
	colorBlue    = lipgloss.Color(hexBlue)
	colorMagenta = lipgloss.Color(hexMagenta)

	stylePlan     = lipgloss.NewStyle().Foreground(colorBlue)
	styleApproval = lipgloss.NewStyle().Foreground(colorRed)
	styleYesNo    = lipgloss.NewStyle().Foreground(colorYellow)
	styleEnter    = lipgloss.NewStyle().Foreground(colorCyan)
	styleCost     = lipgloss.NewStyle().Foreground(colorYellow)
	styleQuestion = lipgloss.NewStyle().Foreground(colorYellow)
	stylePrompt   = lipgloss.NewStyle().Foreground(colorGreen)
	styleThinking = lipgloss.NewStyle().Foreground(colorBlue)
	styleRunning  = lipgloss.NewStyle().Foreground(colorMagenta)

	styleDim       = lipgloss.NewStyle().Foreground(colorDim)
	styleStatusBar = lipgloss.NewStyle().Foreground(colorDim)
)

// tmuxColorFor returns the hex color for a wait type, for tmux status markup.
// Keep in sync with waitStyle.
func tmuxColorFor(wt WaitType) string {
	switch wt {
	case WaitApproval:
		return hexRed
	case WaitYesNo, WaitCost, WaitQuestion:
		return hexYellow
	case WaitEnter:
		return hexCyan
	case WaitPlan:
		return hexBlue
	case WaitPrompt:
		return hexGreen
	case WaitRunning:
		return hexMagenta
	}
	return hexDim
}

func waitStyle(wt WaitType) lipgloss.Style {
	switch wt {
	case WaitPlan:
		return stylePlan
	case WaitApproval:
		return styleApproval
	case WaitYesNo:
		return styleYesNo
	case WaitEnter:
		return styleEnter
	case WaitCost:
		return styleCost
	case WaitQuestion:
		return styleQuestion
	case WaitPrompt:
		return stylePrompt
	case WaitThinking:
		return styleThinking
	case WaitRunning:
		return styleRunning
	default:
		return styleDim
	}
}
