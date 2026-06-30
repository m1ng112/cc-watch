package main

import (
	"strings"
	"testing"
)

func TestNeedsAttention(t *testing.T) {
	attention := []WaitType{WaitApproval, WaitPlan, WaitQuestion, WaitYesNo, WaitEnter, WaitCost}
	for _, wt := range attention {
		if !needsAttention(wt) {
			t.Errorf("needsAttention(%v) = false, want true", wt)
		}
	}
	notAttention := []WaitType{WaitRunning, WaitPrompt, WaitActive, WaitNone}
	for _, wt := range notAttention {
		if needsAttention(wt) {
			t.Errorf("needsAttention(%v) = true, want false", wt)
		}
	}
}

func TestFormatStatusEmpty(t *testing.T) {
	// no panes, or only non-attention panes -> empty string.
	if got := formatStatus(nil); got != "" {
		t.Errorf("formatStatus(nil) = %q, want empty", got)
	}
	only := []PaneInfo{
		{PaneID: "%1", Loc: "a:1", WaitType: WaitRunning},
		{PaneID: "%2", Loc: "b:2", WaitType: WaitPrompt},
	}
	if got := formatStatus(only); got != "" {
		t.Errorf("formatStatus(running/idle only) = %q, want empty", got)
	}
}

func TestFormatStatusListsAttentionPanes(t *testing.T) {
	panes := []PaneInfo{
		{PaneID: "%1", Loc: "upcycle:1", WaitType: WaitPlan},
		{PaneID: "%2", Loc: "work:3", WaitType: WaitApproval},
		{PaneID: "%3", Loc: "busy:4", WaitType: WaitRunning}, // excluded
		{PaneID: "%4", Loc: "lab:5", WaitType: WaitQuestion},
	}
	got := formatStatus(panes)

	// each attention pane's location appears.
	for _, want := range []string{"upcycle:1", "work:3", "lab:5"} {
		if !strings.Contains(got, want) {
			t.Errorf("formatStatus missing %q in %q", want, got)
		}
	}
	// running pane excluded.
	if strings.Contains(got, "busy:4") {
		t.Errorf("formatStatus should exclude running pane, got %q", got)
	}
	// icons present.
	for _, icon := range []string{waitInfo[WaitPlan].Icon, waitInfo[WaitApproval].Icon, waitInfo[WaitQuestion].Icon} {
		if !strings.Contains(got, icon) {
			t.Errorf("formatStatus missing icon %q in %q", icon, got)
		}
	}
	// tmux color markup present for at least the approval (red) pane.
	if !strings.Contains(got, "#[fg="+hexRed+"]") {
		t.Errorf("formatStatus missing red tmux markup in %q", got)
	}
	if !strings.Contains(got, "#[default]") {
		t.Errorf("formatStatus missing #[default] reset in %q", got)
	}
}

func TestFormatStatusOverflow(t *testing.T) {
	var panes []PaneInfo
	for i := 0; i < maxStatusItems+3; i++ {
		panes = append(panes, PaneInfo{PaneID: "%x", Loc: "s:1", WaitType: WaitApproval})
	}
	got := formatStatus(panes)
	if !strings.Contains(got, "+3") {
		t.Errorf("formatStatus overflow should contain +3, got %q", got)
	}
}
