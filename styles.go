package main

import lipgloss "charm.land/lipgloss/v2"

var (
	colorRed    = lipgloss.Color("#e06c75")
	colorYellow = lipgloss.Color("#e5c07b")
	colorGreen  = lipgloss.Color("#98c379")
	colorCyan   = lipgloss.Color("#56b6c2")
	colorDim     = lipgloss.Color("#5c6370")
	colorBlue    = lipgloss.Color("#61afef")
	colorMagenta = lipgloss.Color("#c678dd")

	styleApproval = lipgloss.NewStyle().Foreground(colorRed)
	styleYesNo    = lipgloss.NewStyle().Foreground(colorYellow)
	styleEnter    = lipgloss.NewStyle().Foreground(colorCyan)
	styleCost     = lipgloss.NewStyle().Foreground(colorYellow)
	stylePrompt   = lipgloss.NewStyle().Foreground(colorGreen)
	styleThinking = lipgloss.NewStyle().Foreground(colorBlue)
	styleRunning  = lipgloss.NewStyle().Foreground(colorMagenta)

	styleDim       = lipgloss.NewStyle().Foreground(colorDim)
	styleStatusBar = lipgloss.NewStyle().Foreground(colorDim)
)

func waitStyle(wt WaitType) lipgloss.Style {
	switch wt {
	case WaitApproval:
		return styleApproval
	case WaitYesNo:
		return styleYesNo
	case WaitEnter:
		return styleEnter
	case WaitCost:
		return styleCost
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
