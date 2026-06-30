package main

import (
	"fmt"
	"strings"
)

// needsAttention reports whether a wait state requires the user to act
// (answer/approve), as opposed to running, idle, or unknown.
func needsAttention(wt WaitType) bool {
	switch wt {
	case WaitApproval, WaitPlan, WaitQuestion, WaitYesNo, WaitEnter, WaitCost:
		return true
	}
	return false
}

// maxStatusItems caps how many panes are listed inline before collapsing the
// rest into a "+N" overflow marker, to keep the tmux status line short.
const maxStatusItems = 6

// formatStatus renders a one-line tmux status summary listing the panes that
// need attention as "<icon> <session:window>", colored with tmux markup.
// Returns "" when nothing needs attention.
func formatStatus(panes []PaneInfo) string {
	var parts []string
	extra := 0
	for _, p := range panes {
		if !needsAttention(p.WaitType) {
			continue
		}
		if len(parts) >= maxStatusItems {
			extra++
			continue
		}
		info := waitInfo[p.WaitType]
		parts = append(parts, fmt.Sprintf("#[fg=%s]%s %s#[default]", tmuxColorFor(p.WaitType), info.Icon, p.Loc))
	}
	if extra > 0 {
		parts = append(parts, fmt.Sprintf("+%d", extra))
	}
	return strings.Join(parts, "  ")
}
