package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hook":
			// Invoked by Claude Code hooks. Reads the event from stdin and
			// records pane status. Must stay silent and always succeed.
			runHook(os.Stdin)
			return
		case "install-hooks":
			if err := installHooks(); err != nil {
				fmt.Fprintln(os.Stderr, "cc-watch:", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "cc-watch: 不明なサブコマンド %q (使用可能: hook, install-hooks)\n", os.Args[1])
			os.Exit(1)
		}
	}

	if os.Getenv("TMUX") == "" {
		fmt.Fprintln(os.Stderr, "cc-watch: tmux セッション内で実行してください")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if m, ok := result.(model); ok && m.jumpTarget != "" {
		switchToPane(m.jumpTarget)
	}
}
