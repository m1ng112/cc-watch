package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
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
