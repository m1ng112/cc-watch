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

	approval := " Path contains '..' traversal\n" +
		" Do you want to proceed?\n" +
		" ❯ 1. Yes\n" +
		"   2. No"

	yesno := "Overwrite the file? (y/n)"
	enter := "Press enter to continue"
	cost := "Total API Cost: $1.23"

	tests := []struct {
		name     string
		content  string
		isClaude bool
		want     WaitType
	}{
		{"idle empty prompt -> 回答待ち", idle, true, WaitPrompt},
		{"working spinner -> 実行中", working, true, WaitRunning},
		{"working short spinner -> 実行中", workingShort, true, WaitRunning},
		{"idle with leftover bullet -> 回答待ち", idleWithBullet, true, WaitPrompt},
		{"approval prompt -> 確認待ち", approval, true, WaitApproval},
		{"yes/no prompt", yesno, true, WaitYesNo},
		{"press enter prompt", enter, true, WaitEnter},
		{"api cost prompt", cost, true, WaitCost},
		{"empty prompt on non-claude shell is ignored", "❯" + nbsp, false, WaitNone},
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

func TestWaitPromptLabelIsAnswerWaiting(t *testing.T) {
	if got := waitInfo[WaitPrompt].Label; got != "回答待ち" {
		t.Errorf("waitInfo[WaitPrompt].Label = %q, want %q", got, "回答待ち")
	}
}
