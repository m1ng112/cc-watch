# cc-watch

A TUI tool that scans tmux panes running **Claude Code** sessions, detects which ones are waiting for user input, and lets you jump to them instantly.

Auto-refreshes every 3 seconds so you never miss a prompt.

**[日本語版はこちら (Japanese)](README.ja.md)**

## Screenshot

```
┌── cc-watch  3 pane(s) waiting ─────────────────────────────────┐
│                                │                               │
│ ▸ (1) !  Approval   0:3: claude│  [Preview: last 30 lines of   │
│   (2) ›  Prompt     0:1: claude│   selected pane with ANSI     │
│   (3) ?  Yes / No   0:5: claude│   colors preserved]           │
│                                │                               │
├────────────────────────────────┴───────────────────────────────┤
│ 14:32:05  ↑↓ move  enter jump  q quit  r refresh              │
└────────────────────────────────────────────────────────────────┘
```

## Detected States

| Icon | Label (JA) | Meaning |
|------|------------|---------|
| `✓` | 承認待ち | Waiting for plan approval (ExitPlanMode) |
| `!` | 確認待ち | Waiting for command/tool permission |
| `▸` | 回答待ち | Blocked on a multiple-choice question (AskUserQuestion) |
| `?` | Yes / No | A `(y/n)` prompt |
| `↵` | Enter待ち | `Press Enter` etc. |
| `$` | コスト確認 | `API Cost` etc. |
| `›` | 指示待ち | Turn finished, idle waiting for the next instruction (empty prompt) |
| `⏺` | 実行中 | Generating (spinner) |
| `·` | 動作中 | Fallback when nothing could be determined |

## Detection: hooks (recommended) and scraping

cc-watch determines state two ways:

1. **Claude Code hooks (recommended, reliable)** — running `cc-watch install-hooks` registers hooks in `~/.claude/settings.json`. Each Claude session then reports its state transitions (start / finish / approval / permission / question) directly to cc-watch, tagged with `$TMUX_PANE`. This is layout-independent and accurate.
2. **Terminal scraping (fallback)** — for panes without hooks (or sessions outside tmux), the pane tail is parsed with regexes to infer state.

```sh
# Register hooks (your existing settings are backed up; other hooks are preserved)
cc-watch install-hooks
```

> Newly registered hooks do not apply to already-running Claude Code sessions — restart them or check `/hooks`. Only Claude started inside tmux can be correlated.

## Requirements

- **tmux** — session management
- **Go 1.23+** — build

## Install

```sh
go install github.com/m1ng112/cc-watch@latest
```

Or build manually:

```sh
git clone https://github.com/m1ng112/cc-watch.git
cd cc-watch
go build -o cc-watch .
ln -s "$(pwd)/cc-watch" ~/.local/bin/cc-watch
```

## Usage

Run inside a tmux session:

```sh
cc-watch
```

### Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Enter` | Jump to selected pane |
| `r` | Rescan now |
| `q` / `Esc` / `Ctrl+C` | Quit |

### Auto-Refresh

Panes are rescanned every 3 seconds. When no sessions are waiting, "✓ No waiting panes" is displayed.

### tmux Key Binding Example

```sh
# prefix + w to launch cc-watch
bind-key w run-shell -b 'cc-watch'
```

## License

MIT
