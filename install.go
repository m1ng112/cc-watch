package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// hookEvents are the Claude Code hook events cc-watch registers, with an
// optional tool matcher. PreToolUse is narrowed to the blocking tools so the
// hook does not fire on every tool call.
var hookEvents = []struct {
	event   string
	matcher string
}{
	{"UserPromptSubmit", ""},
	{"Stop", ""},
	{"Notification", ""},
	{"SessionStart", ""},
	{"SessionEnd", ""},
	{"PreToolUse", "ExitPlanMode|AskUserQuestion"},
}

// hookEntry is the command hook that invokes `cc-watch hook`. Using command +
// args avoids any shell quoting concerns with the executable path.
func hookEntry(exe string) map[string]any {
	return map[string]any{
		"type":    "command",
		"command": exe,
		"args":    []any{"hook"},
	}
}

// eventHasOurHook reports whether an event's hook groups already contain a
// cc-watch hook, so install is idempotent.
func eventHasOurHook(groups []any, exe string) bool {
	base := filepath.Base(exe)
	for _, g := range groups {
		gm, ok := g.(map[string]any)
		if !ok {
			continue
		}
		hs, _ := gm["hooks"].([]any)
		for _, h := range hs {
			b, _ := json.Marshal(h)
			s := string(b)
			if strings.Contains(s, base) && strings.Contains(s, "hook") {
				return true
			}
		}
	}
	return false
}

// mergeHooks merges cc-watch's hook registrations into an existing settings
// map, preserving all other keys and any existing user hooks. It is idempotent.
func mergeHooks(settings map[string]any, exe string) map[string]any {
	if settings == nil {
		settings = map[string]any{}
	}
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	for _, he := range hookEvents {
		groups, _ := hooks[he.event].([]any)
		if eventHasOurHook(groups, exe) {
			continue
		}
		group := map[string]any{"hooks": []any{hookEntry(exe)}}
		if he.matcher != "" {
			group["matcher"] = he.matcher
		}
		hooks[he.event] = append(groups, group)
	}
	settings["hooks"] = hooks
	return settings
}

// installHooks merges cc-watch hooks into ~/.claude/settings.json (creating a
// backup) so the running binary's status is reported via Claude Code hooks.
func installHooks() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate cc-watch binary: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".claude", "settings.json")

	var settings map[string]any
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("%s is not valid JSON; not modifying it: %w", path, err)
		}
		backup := path + ".cc-watch-bak"
		if err := os.WriteFile(backup, data, 0o644); err != nil {
			return fmt.Errorf("could not write backup %s: %w", backup, err)
		}
		fmt.Printf("バックアップを作成: %s\n", backup)
	case os.IsNotExist(err):
		// fresh settings file
	default:
		return err
	}

	merged := mergeHooks(settings, exe)
	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("cc-watch hooks を %s に登録しました。\n", path)
	fmt.Println("既存の Claude Code セッションには新しい hooks が反映されません。各セッションで /hooks を確認するか、再起動してください。")
	return nil
}
