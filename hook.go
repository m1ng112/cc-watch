package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// hookInput is the subset of the Claude Code hook stdin JSON that we care about.
// Every hook event includes hook_event_name; tool events add tool_name; the
// Notification event adds type; SessionStart/End add source.
type hookInput struct {
	HookEventName string `json:"hook_event_name"`
	ToolName      string `json:"tool_name"`
	Type          string `json:"type"`
	Source        string `json:"source"`
}

// hookAction is the decision derived from a hook event: either write a state,
// delete the pane's status, or do nothing (zero value).
type hookAction struct {
	state  string
	delete bool
}

// paneStatus is what we persist per tmux pane.
type paneStatus struct {
	State string `json:"state"`
	TS    int64  `json:"ts"`
}

// eventToAction maps a hook event to the status action to take.
func eventToAction(in hookInput) hookAction {
	switch in.HookEventName {
	case "UserPromptSubmit":
		return hookAction{state: "running"}
	case "Stop":
		return hookAction{state: "idle"}
	case "SessionStart":
		return hookAction{state: "idle"}
	case "SessionEnd":
		return hookAction{delete: true}
	case "PreToolUse", "PermissionRequest":
		switch in.ToolName {
		case "ExitPlanMode":
			return hookAction{state: "plan"}
		case "AskUserQuestion":
			return hookAction{state: "question"}
		}
		// Any other tool awaiting permission is a confirmation prompt.
		if in.HookEventName == "PermissionRequest" {
			return hookAction{state: "approval"}
		}
	case "Notification":
		switch in.Type {
		case "permission_prompt":
			return hookAction{state: "approval"}
		case "idle_prompt":
			return hookAction{state: "idle"}
		case "elicitation_dialog":
			return hookAction{state: "question"}
		}
	}
	return hookAction{}
}

// blockRank ranks the "blocked" states so a less specific event cannot clobber
// a more specific one. Non-blocked states (running/idle) rank 0.
func blockRank(state string) int {
	switch state {
	case "plan", "question":
		return 2
	case "approval":
		return 1
	}
	return 0
}

// staleSeconds is how long a blocked state is protected from being downgraded
// by a less specific blocked event.
const staleSeconds = 10

// shouldWrite decides whether an incoming state should overwrite the current
// stored status. running/idle are clear turn transitions and always apply;
// blocked states do not downgrade a more specific, recent blocked state.
func shouldWrite(incoming string, cur *paneStatus, now time.Time) bool {
	if incoming == "running" || incoming == "idle" {
		return true
	}
	if cur == nil {
		return true
	}
	if now.Unix()-cur.TS > staleSeconds {
		return true
	}
	return blockRank(incoming) >= blockRank(cur.State)
}

// resolveWaitType determines a pane's wait state, preferring the hook-reported
// status (event-driven, reliable) and falling back to terminal scraping.
func resolveWaitType(paneID, tail string, isClaude bool) WaitType {
	scraped := detectType(tail, isClaude)

	st, hasHook := readStatus(paneID)
	hookWT, hookOK := WaitNone, false
	if hasHook {
		hookWT, hookOK = stateToWaitType(st.State)
	}

	// Safety net: when the screen clearly shows a blocking prompt and the hook
	// is not actively tracking a running turn (no hook, or it went stale at
	// "idle"), trust the screen. This surfaces prompts the hooks miss — e.g. a
	// tool-permission dialog whose Notification hook may not fire while the pane
	// is focused — without letting incidental prompt-like text in a *running*
	// pane cause false positives.
	if needsAttention(scraped) && (!hookOK || st.State == "idle") {
		return scraped
	}
	if hookOK {
		return hookWT
	}
	if scraped != WaitNone {
		return scraped
	}
	return WaitActive
}

// stateToWaitType maps a persisted state string to a WaitType for display.
func stateToWaitType(state string) (WaitType, bool) {
	switch state {
	case "running":
		return WaitRunning, true
	case "idle":
		return WaitPrompt, true
	case "approval":
		return WaitApproval, true
	case "plan":
		return WaitPlan, true
	case "question":
		return WaitQuestion, true
	}
	return WaitNone, false
}

var paneIDSanitizer = regexp.MustCompile(`[^A-Za-z0-9_-]`)

// statusDir returns the directory holding per-pane status files. Honors
// CC_WATCH_STATE_DIR for testing / overrides.
func statusDir() string {
	if d := os.Getenv("CC_WATCH_STATE_DIR"); d != "" {
		return d
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = filepath.Join(os.TempDir(), "cc-watch-cache")
	}
	return filepath.Join(cache, "cc-watch", "panes")
}

func statusPath(paneID string) string {
	return filepath.Join(statusDir(), paneIDSanitizer.ReplaceAllString(paneID, "_")+".json")
}

func readStatus(paneID string) (paneStatus, bool) {
	data, err := os.ReadFile(statusPath(paneID))
	if err != nil {
		return paneStatus{}, false
	}
	var st paneStatus
	if err := json.Unmarshal(data, &st); err != nil {
		return paneStatus{}, false
	}
	return st, true
}

// writeStatus persists state for a pane, honoring the shouldWrite guard so a
// less specific blocked event cannot clobber a more specific recent one.
func writeStatus(paneID, state string, now time.Time) error {
	cur, ok := readStatus(paneID)
	var curPtr *paneStatus
	if ok {
		curPtr = &cur
	}
	if !shouldWrite(state, curPtr, now) {
		return nil
	}
	if err := os.MkdirAll(statusDir(), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(paneStatus{State: state, TS: now.Unix()})
	if err != nil {
		return err
	}
	return os.WriteFile(statusPath(paneID), data, 0o644)
}

func deleteStatus(paneID string) error {
	err := os.Remove(statusPath(paneID))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// runHook is the `cc-watch hook` entry point. It reads the hook JSON from stdin,
// correlates to the tmux pane via $TMUX_PANE, and updates the pane status.
// It never writes to stdout and always succeeds (exit 0) so it cannot disrupt
// the Claude Code session.
func runHook(stdin io.Reader) {
	pane := os.Getenv("TMUX_PANE")
	if pane == "" {
		return
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return
	}
	var in hookInput
	if err := json.Unmarshal(data, &in); err != nil {
		return
	}
	action := eventToAction(in)
	switch {
	case action.delete:
		_ = deleteStatus(pane)
	case action.state != "":
		_ = writeStatus(pane, action.state, time.Now())
	}
}
