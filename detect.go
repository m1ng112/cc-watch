package main

import "regexp"

// WaitType represents the type of waiting state detected in a pane.
type WaitType int

const (
	WaitNone WaitType = iota
	WaitPlan
	WaitApproval
	WaitYesNo
	WaitEnter
	WaitCost
	WaitQuestion
	WaitPrompt
	WaitThinking
	WaitRunning
	WaitActive
)

type waitPattern struct {
	wtype   WaitType
	pattern *regexp.Regexp
}

// Ordered by priority (first match wins).
var waitPatterns = []waitPattern{
	// Plan-mode approval (ExitPlanMode). "No, keep planning" is the last option
	// and stays at the bottom even when a long plan pushes the "Ready to code?"
	// header off screen. Checked first so it wins over the generic approval
	// keywords below, which a plan prompt may also contain.
	{WaitPlan, regexp.MustCompile(`keep planning|Ready to code\?|Here is Claude's plan`)},
	{WaitApproval, regexp.MustCompile(`Do you want to proceed|Would you like to|Shall I|should I proceed|Run anyway|Allow|Deny`)},
	{WaitYesNo, regexp.MustCompile(`\(y/n\)|\(Y/n\)|\(yes/no\)|Yes/No`)},
	{WaitEnter, regexp.MustCompile(`(?i)press enter|hit enter|\[Enter\]`)},
	{WaitCost, regexp.MustCompile(`API Cost|tokens used|cost estimate`)},
	// Interactive selection widget, e.g. AskUserQuestion: the ❯ selector points
	// at a numbered option ("❯ 1. ..."). Claude is blocked waiting for the user
	// to pick an answer. Checked after the specific keyword prompts above so they
	// keep their more precise labels.
	{WaitQuestion, regexp.MustCompile(`(?m)^[\s\x{00A0}]*❯[\s\x{00A0}]+[0-9]+\.`)},
	// Active spinner line, e.g. "✢ Kneading… (14m 14s · ↓ 59.8k tokens)".
	// The elapsed-time / token counter (or "esc to interrupt") is only rendered
	// while Claude is generating, making it the one reliable "busy" signal.
	// A leftover ⏺ bullet, by contrast, stays on screen after Claude finishes,
	// so it must NOT be treated as busy.
	{WaitRunning, regexp.MustCompile(`…\s*\([0-9]+m?\s?[0-9]*s\b|\(esc to interrupt\)`)},
	// Empty input box: the prompt glyph followed only by whitespace, including
	// the non-breaking space (U+00A0) Claude Code emits. Present when Claude has
	// finished its turn and is waiting for the user to respond.
	{WaitPrompt, regexp.MustCompile(`(?m)^[\s\x{00A0}]*❯[\s\x{00A0}]*$`)},
}

// WaitInfo holds display information for a wait type.
type WaitInfo struct {
	Icon  string
	Label string
}

var waitInfo = map[WaitType]WaitInfo{
	WaitPlan:     {Icon: "✓", Label: "承認待ち"},
	WaitApproval: {Icon: "!", Label: "確認待ち"},
	WaitYesNo:    {Icon: "?", Label: "Yes / No"},
	WaitEnter:    {Icon: "↵", Label: "Enter待ち"},
	WaitCost:     {Icon: "$", Label: "コスト確認"},
	WaitQuestion: {Icon: "▸", Label: "回答待ち"},
	WaitPrompt:   {Icon: "›", Label: "指示待ち"},
	WaitThinking: {Icon: "✻", Label: "思考中"},
	WaitRunning:  {Icon: "⏺", Label: "実行中"},
	WaitActive:   {Icon: "·", Label: "動作中"},
}

func detectType(content string, isClaude bool) WaitType {
	for _, wp := range waitPatterns {
		if wp.pattern.MatchString(content) {
			if (wp.wtype == WaitPrompt || wp.wtype == WaitQuestion) && !isClaude {
				continue
			}
			return wp.wtype
		}
	}
	return WaitNone
}
