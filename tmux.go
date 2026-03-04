package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// PaneInfo represents a tmux pane that is waiting for user input.
type PaneInfo struct {
	PaneID   string
	Name     string
	WaitType WaitType
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[mKHJA-Za-z]`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

var claudeUIRegex = regexp.MustCompile(`✻ (Thinking|Cooked|Crunched|Cogitated|Pondered)|⏺ |─{4,}`)

func hasClaudeUI(content string) bool {
	return claudeUIRegex.MatchString(content)
}

// getPaneTail captures a pane's content, strips ANSI codes, removes blank lines,
// and returns the last n non-empty lines as a single string.
func getPaneTail(paneID string, n int) string {
	out, err := exec.Command("tmux", "capture-pane", "-p", "-t", paneID).Output()
	if err != nil {
		return ""
	}
	content := stripANSI(string(out))
	var nonEmpty []string
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	if len(nonEmpty) > n {
		nonEmpty = nonEmpty[len(nonEmpty)-n:]
	}
	return strings.Join(nonEmpty, "\n")
}

// capturePane returns the last 30 lines of a pane with ANSI color codes preserved.
func capturePane(paneID string) string {
	out, err := exec.Command("tmux", "capture-pane", "-p", "-e", "-t", paneID).Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 30 {
		lines = lines[len(lines)-30:]
	}
	return strings.Join(lines, "\n")
}

// isClaudeProcess checks if a pane's shell has a claude/claude-code child process.
func isClaudeProcess(panePID string) bool {
	out, err := exec.Command("pgrep", "-P", panePID).Output()
	if err != nil {
		return false
	}
	for _, child := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if child == "" {
			continue
		}
		comm, err := exec.Command("ps", "-p", child, "-o", "comm=").Output()
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		if name == "claude" || name == "claude-code" {
			return true
		}
		if name == "node" {
			args, err := exec.Command("ps", "-p", child, "-o", "args=").Output()
			if err != nil {
				continue
			}
			if strings.Contains(string(args), "claude") {
				return true
			}
		}
	}
	return false
}

// scanAllPanes scans all tmux panes and returns those waiting for user input.
func scanAllPanes() []PaneInfo {
	out, err := exec.Command("tmux", "list-panes", "-a",
		"-F", "#{pane_id}|#{session_name}|#{window_index}|#{pane_index}|#{window_name}|#{pane_pid}").Output()
	if err != nil {
		return nil
	}

	var panes []PaneInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 6 {
			continue
		}
		paneID := parts[0]
		session := parts[1]
		win := parts[2]
		// paneIdx := parts[3]
		wname := parts[4]
		panePID := parts[5]

		tail := getPaneTail(paneID, 12)
		if tail == "" {
			continue
		}

		procOK := isClaudeProcess(panePID)
		uiOK := hasClaudeUI(tail)
		if !procOK && !uiOK {
			continue
		}

		isClaude := procOK || uiOK
		wtype := detectType(tail, isClaude)
		if wtype == WaitNone {
			wtype = WaitActive
		}

		var name string
		if wname != "" && wname != win {
			name = fmt.Sprintf("%s:%s: %s", session, win, wname)
		} else {
			name = fmt.Sprintf("%s:%s", session, win)
		}

		panes = append(panes, PaneInfo{
			PaneID:   paneID,
			Name:     name,
			WaitType: wtype,
		})
	}
	return panes
}

// switchToPane switches the current tmux client to the specified pane.
func switchToPane(paneID string) error {
	return exec.Command("tmux", "switch-client", "-t", paneID).Run()
}

// inTmux checks if we are running inside a tmux session.
func inTmux() bool {
	_, err := exec.Command("tmux", "display-message", "-p", "#{session_name}").Output()
	return err == nil
}
