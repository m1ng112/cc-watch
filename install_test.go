package main

import (
	"encoding/json"
	"testing"
)

const exe = "/Users/x/.local/bin/cc-watch"

// hookEventsForResult walks a merged settings map and returns, per event, the
// number of hook groups and whether one of ours is present.
func eventGroups(t *testing.T, settings map[string]any, event string) []any {
	t.Helper()
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("settings has no hooks map: %+v", settings)
	}
	groups, _ := hooks[event].([]any)
	return groups
}

func TestMergeHooksFromEmpty(t *testing.T) {
	got := mergeHooks(map[string]any{}, exe)

	for _, he := range hookEvents {
		groups := eventGroups(t, got, he.event)
		if !eventHasOurHook(groups, exe) {
			t.Errorf("event %q missing our hook", he.event)
		}
		if he.matcher != "" {
			// the group carrying our hook should have the matcher.
			found := false
			for _, g := range groups {
				gm := g.(map[string]any)
				if gm["matcher"] == he.matcher {
					found = true
				}
			}
			if !found {
				t.Errorf("event %q missing matcher %q", he.event, he.matcher)
			}
		}
	}
}

func TestMergeHooksIdempotent(t *testing.T) {
	once := mergeHooks(map[string]any{}, exe)
	twice := mergeHooks(once, exe)

	for _, he := range hookEvents {
		groups := eventGroups(t, twice, he.event)
		count := 0
		for _, g := range groups {
			if eventHasOurHook([]any{g}, exe) {
				count++
			}
		}
		if count != 1 {
			t.Errorf("event %q has %d cc-watch hook groups after double merge, want 1", he.event, count)
		}
	}
}

func TestMergeHooksPreservesExisting(t *testing.T) {
	// existing settings: an unrelated top-level key and a user's own Stop hook.
	existing := map[string]any{
		"model": "opus",
		"hooks": map[string]any{
			"Stop": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"type": "command", "command": "/usr/bin/say done"},
					},
				},
			},
		},
	}
	got := mergeHooks(existing, exe)

	if got["model"] != "opus" {
		t.Errorf("unrelated key 'model' was dropped")
	}

	stop := eventGroups(t, got, "Stop")
	// must keep the user's hook AND add ours.
	b, _ := json.Marshal(stop)
	s := string(b)
	if !contains(s, "/usr/bin/say done") {
		t.Errorf("user's existing Stop hook was dropped: %s", s)
	}
	if !eventHasOurHook(stop, exe) {
		t.Errorf("our Stop hook was not added alongside the user's")
	}
	if len(stop) != 2 {
		t.Errorf("Stop should have 2 groups (user + ours), got %d", len(stop))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
