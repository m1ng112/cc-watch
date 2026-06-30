package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEventToAction(t *testing.T) {
	tests := []struct {
		name       string
		in         hookInput
		wantState  string
		wantDelete bool
	}{
		{"prompt submit -> running", hookInput{HookEventName: "UserPromptSubmit"}, "running", false},
		{"stop -> idle", hookInput{HookEventName: "Stop"}, "idle", false},
		{"session start -> idle", hookInput{HookEventName: "SessionStart", Source: "startup"}, "idle", false},
		{"session end -> delete", hookInput{HookEventName: "SessionEnd", Source: "logout"}, "", true},
		{"pretooluse ExitPlanMode -> plan", hookInput{HookEventName: "PreToolUse", ToolName: "ExitPlanMode"}, "plan", false},
		{"pretooluse AskUserQuestion -> question", hookInput{HookEventName: "PreToolUse", ToolName: "AskUserQuestion"}, "question", false},
		{"pretooluse Bash -> ignore", hookInput{HookEventName: "PreToolUse", ToolName: "Bash"}, "", false},
		{"permissionrequest Edit -> approval", hookInput{HookEventName: "PermissionRequest", ToolName: "Edit"}, "approval", false},
		{"permissionrequest Bash -> approval", hookInput{HookEventName: "PermissionRequest", ToolName: "Bash"}, "approval", false},
		{"permissionrequest ExitPlanMode -> plan", hookInput{HookEventName: "PermissionRequest", ToolName: "ExitPlanMode"}, "plan", false},
		{"permissionrequest AskUserQuestion -> question", hookInput{HookEventName: "PermissionRequest", ToolName: "AskUserQuestion"}, "question", false},
		{"notification permission -> approval", hookInput{HookEventName: "Notification", Type: "permission_prompt"}, "approval", false},
		{"notification idle -> idle", hookInput{HookEventName: "Notification", Type: "idle_prompt"}, "idle", false},
		{"notification elicitation -> question", hookInput{HookEventName: "Notification", Type: "elicitation_dialog"}, "question", false},
		{"notification auth -> ignore", hookInput{HookEventName: "Notification", Type: "auth_success"}, "", false},
		{"unknown event -> ignore", hookInput{HookEventName: "FileChanged"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventToAction(tt.in)
			if got.state != tt.wantState || got.delete != tt.wantDelete {
				t.Errorf("eventToAction(%+v) = {state:%q delete:%v}, want {state:%q delete:%v}",
					tt.in, got.state, got.delete, tt.wantState, tt.wantDelete)
			}
		})
	}
}

func TestStateToWaitType(t *testing.T) {
	cases := map[string]WaitType{
		"running":  WaitRunning,
		"idle":     WaitPrompt,
		"approval": WaitApproval,
		"plan":     WaitPlan,
		"question": WaitQuestion,
	}
	for state, want := range cases {
		if got, ok := stateToWaitType(state); !ok || got != want {
			t.Errorf("stateToWaitType(%q) = %v,%v want %v,true", state, got, ok, want)
		}
	}
	if _, ok := stateToWaitType("bogus"); ok {
		t.Errorf("stateToWaitType(bogus) should be !ok")
	}
}

func TestShouldWrite(t *testing.T) {
	now := time.Unix(1000, 0)
	recent := &paneStatus{State: "plan", TS: 995}     // 5s old
	stalePlan := &paneStatus{State: "plan", TS: 980}  // 20s old
	running := &paneStatus{State: "running", TS: 999} // 1s old

	tests := []struct {
		name     string
		incoming string
		cur      *paneStatus
		want     bool
	}{
		{"running always writes", "running", recent, true},
		{"idle always writes", "idle", recent, true},
		{"approval writes when no current", "approval", nil, true},
		{"approval does NOT downgrade recent plan", "approval", recent, false},
		{"approval writes over stale plan", "approval", stalePlan, true},
		{"approval writes over running", "approval", running, true},
		{"question writes over recent plan (equal rank)", "question", recent, true},
		{"plan writes over recent plan", "plan", recent, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldWrite(tt.incoming, tt.cur, now); got != tt.want {
				t.Errorf("shouldWrite(%q, %+v) = %v, want %v", tt.incoming, tt.cur, got, tt.want)
			}
		})
	}
}

func TestResolveWaitType(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CC_WATCH_STATE_DIR", dir)

	spinnerTail := "✢ Kneading… (3s · ↓ 1.0k tokens)\n❯ "
	emptyTail := "some output with no markers"

	// hook status takes precedence over whatever the terminal shows.
	if err := writeStatus("%1", "plan", time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	if got := resolveWaitType("%1", spinnerTail, true); got != WaitPlan {
		t.Errorf("hook status should win: got %v, want WaitPlan", got)
	}

	// no hook status -> scraping fallback (spinner -> running).
	if got := resolveWaitType("%2", spinnerTail, true); got != WaitRunning {
		t.Errorf("scraping fallback: got %v, want WaitRunning", got)
	}

	// no hook status, nothing detectable -> WaitActive.
	if got := resolveWaitType("%3", emptyTail, true); got != WaitActive {
		t.Errorf("default fallback: got %v, want WaitActive", got)
	}

	// corrupt/unknown hook state -> fall through to scraping.
	if err := writeStatus("%4", "running", time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	// overwrite with a bogus state directly.
	if err := os.WriteFile(statusPath("%4"), []byte(`{"state":"bogus","ts":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveWaitType("%4", spinnerTail, true); got != WaitRunning {
		t.Errorf("unknown hook state should fall through to scraping: got %v, want WaitRunning", got)
	}

	// Safety net: a blocking prompt is on screen but the hook went stale at
	// "idle" (e.g. a tool-permission dialog the Notification hook missed).
	// Scraping must win so the pane still surfaces as needing attention.
	questionTail := "どちらにしますか？\n❯ 1. A案\n  2. B案"
	if err := writeStatus("%5", "idle", time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	if got := resolveWaitType("%5", questionTail, true); got != WaitQuestion {
		t.Errorf("stale idle hook + on-screen prompt should surface via scraping: got %v, want WaitQuestion", got)
	}

	// A running hook suppresses scraping ONLY while the spinner is on screen
	// (Claude is actually generating); incidental prompt-like text then does
	// not cause a false positive.
	spinnerWithPrompt := "✶ Working… (3s · ↓ 1.0k tokens)\n❯ 1. A案\n  2. B案"
	if err := writeStatus("%6", "running", time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	if got := resolveWaitType("%6", spinnerWithPrompt, true); got != WaitRunning {
		t.Errorf("running hook + spinner should suppress scraping: got %v, want WaitRunning", got)
	}

	// But a running hook with NO spinner and a blocking prompt on screen means
	// the hook went stale (e.g. a tool-permission prompt fires no hook): the
	// screen wins. This is the docker/Bash permission case.
	if err := writeStatus("%8", "running", time.Unix(1, 0)); err != nil {
		t.Fatal(err)
	}
	approvalTail := "  ls /etc\n Do you want to proceed?\n ❯ 1. Yes\n   2. No\n Esc to cancel"
	if got := resolveWaitType("%8", approvalTail, true); got != WaitApproval {
		t.Errorf("running hook but no spinner + permission prompt should surface: got %v, want WaitApproval", got)
	}

	// No hook + on-screen prompt -> scraping attention.
	if got := resolveWaitType("%7", questionTail, true); got != WaitQuestion {
		t.Errorf("no hook + on-screen prompt: got %v, want WaitQuestion", got)
	}
}

func TestStatusStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CC_WATCH_STATE_DIR", dir)

	if _, ok := readStatus("%99"); ok {
		t.Fatal("readStatus on missing pane should be !ok")
	}

	now := time.Unix(2000, 0)
	if err := writeStatus("%99", "plan", now); err != nil {
		t.Fatalf("writeStatus: %v", err)
	}
	got, ok := readStatus("%99")
	if !ok {
		t.Fatal("readStatus after write should be ok")
	}
	if got.State != "plan" || got.TS != 2000 {
		t.Errorf("readStatus = %+v, want {plan 2000}", got)
	}

	// pane id is sanitized into the filename.
	if _, err := os.Stat(filepath.Join(dir, "_99.json")); err != nil {
		t.Errorf("expected sanitized status file _99.json: %v", err)
	}

	if err := deleteStatus("%99"); err != nil {
		t.Fatalf("deleteStatus: %v", err)
	}
	if _, ok := readStatus("%99"); ok {
		t.Error("readStatus after delete should be !ok")
	}
	// deleting a missing pane is not an error.
	if err := deleteStatus("%99"); err != nil {
		t.Errorf("deleteStatus on missing pane should be nil, got %v", err)
	}
}
