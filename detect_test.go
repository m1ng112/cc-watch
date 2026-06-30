package main

import "testing"

// nbsp is the non-breaking space (U+00A0) that Claude Code renders after the
// prompt glyph on its empty input line.
const nbsp = " "

func TestDetectType(t *testing.T) {
	// idle: Claude finished its turn and is waiting for the user. The input box
	// shows the prompt glyph followed by a non-breaking space, and there is no
	// spinner line.
	idle := "  実装計画を作成します。次は承認をもらってから進めます。\n" +
		"────────────────────────────────────────\n" +
		"❯" + nbsp + "\n" +
		"────────────────────────────────────────\n" +
		"  Opus 4.8 (1M context) │ upcycle:main      new task? /clear to save 337.6k tokens"

	// working: a live spinner line with the elapsed-time / token counter is
	// present. The empty prompt box is ALSO present, so the spinner must win.
	working := "  プランを全面的に書き直します。\n" +
		"✢ Kneading… (14m 14s · ↓ 59.8k tokens · thinking with xhigh effort)\n" +
		"────────────────────────────────────────\n" +
		"❯" + nbsp + "\n" +
		"────────────────────────────────────────\n" +
		"  Opus 4.8 (1M context) │ upcycle:main"

	// working (short form): spinner without the leading minutes segment.
	workingShort := "⏺ Bash(echo hi)\n" +
		"  ⎿  Running…\n" +
		"· Gesticulating… (34s · ↓ 3.8k tokens)\n" +
		"❯" + nbsp

	// idle-with-bullet: Claude printed a final ⏺ message and is now waiting.
	// The ⏺ bullet is still on screen but there is no spinner, so this is
	// "回答待ち", not "実行中".
	idleWithBullet := "⏺ 完了しました。確認してください。\n" +
		"────────────────────────────────────────\n" +
		"❯" + nbsp + "\n" +
		"────────────────────────────────────────\n" +
		"  Opus 4.8 (1M context) │ cc-watch:main"

	// plan approval (ExitPlanMode). For a long plan only the bottom options are
	// visible; "No, keep planning" is the stable last-option marker.
	planLong := "  5. cd api && make check でグリーン確認。\n" +
		"────────────────────────────────────────\n" +
		"❯ 1. Yes\n" +
		"  2. Yes, and auto-accept edits\n" +
		"  3. No, keep planning\n" +
		"────────────────────────────────────────\n" +
		"  Opus 4.8 (1M context) │ upcycle:main"

	// plan approval with the header still visible (short plan).
	planShort := "  Ready to code?\n" +
		"  Here is Claude's plan:\n" +
		"  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌\n" +
		"  概要を実装します。\n" +
		"❯ 1. Yes\n" +
		"  2. No, keep planning"

	// command-execution permission. No plan markers -> stays 確認待ち.
	approval := " Path contains '..' traversal\n" +
		" Do you want to proceed?\n" +
		" ❯ 1. Yes\n" +
		"   2. No"

	// file-edit permission prompt ("Do you want to make this edit ...").
	editPermission := " Do you want to make this edit to assessment_chart.py?\n" +
		" ❯ 1. Yes\n" +
		"   2. Yes, allow all edits during this session (shift+tab)\n" +
		"   3. No"

	// question: Claude posed a selection (e.g. AskUserQuestion) with no
	// approval keyword. The ❯ selector points at a numbered option, so Claude
	// is blocked waiting for the user's answer.
	question := "  どちらの方針で進めますか？\n" +
		"❯ 1. A案: シンプルに実装\n" +
		"  2. B案: 拡張性を重視\n" +
		"  3. C案: 両者の折衷"

	// question with NBSP between selector and number.
	questionNBSP := "  Which option?\n" +
		"❯" + nbsp + "1. Foo\n" +
		"  2. Bar"

	yesno := "Overwrite the file? (y/n)"
	enter := "Press enter to continue"
	cost := "Total API Cost: $1.23"

	tests := []struct {
		name     string
		content  string
		isClaude bool
		want     WaitType
	}{
		{"idle empty prompt -> 指示待ち", idle, true, WaitPrompt},
		{"working spinner -> 実行中", working, true, WaitRunning},
		{"working short spinner -> 実行中", workingShort, true, WaitRunning},
		{"idle with leftover bullet -> 指示待ち", idleWithBullet, true, WaitPrompt},
		{"selection UI -> 回答待ち", question, true, WaitQuestion},
		{"selection UI with NBSP -> 回答待ち", questionNBSP, true, WaitQuestion},
		{"plan approval (long, options only) -> 承認待ち", planLong, true, WaitPlan},
		{"plan approval (short, header visible) -> 承認待ち", planShort, true, WaitPlan},
		{"command permission stays 確認待ち", approval, true, WaitApproval},
		{"edit permission -> 確認待ち", editPermission, true, WaitApproval},
		{"yes/no prompt", yesno, true, WaitYesNo},
		{"press enter prompt", enter, true, WaitEnter},
		{"api cost prompt", cost, true, WaitCost},
		{"empty prompt on non-claude shell is ignored", "❯" + nbsp, false, WaitNone},
		{"selection UI on non-claude shell is ignored", "❯ 1. foo", false, WaitNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectType(tt.content, tt.isClaude)
			if got != tt.want {
				t.Errorf("detectType() = %v (%s), want %v (%s)",
					got, waitInfo[got].Label, tt.want, waitInfo[tt.want].Label)
			}
		})
	}
}

func TestWaitLabels(t *testing.T) {
	cases := []struct {
		wt   WaitType
		want string
	}{
		{WaitPlan, "承認待ち"},
		{WaitQuestion, "回答待ち"},
		{WaitPrompt, "指示待ち"},
	}
	for _, c := range cases {
		if got := waitInfo[c.wt].Label; got != c.want {
			t.Errorf("waitInfo[%v].Label = %q, want %q", c.wt, got, c.want)
		}
	}
}
