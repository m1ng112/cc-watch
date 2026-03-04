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

## Detected Wait States

| Icon | Label | Detection Pattern |
|------|-------|-------------------|
| `!` | Approval | `Do you want to proceed`, `Allow`, `Deny` |
| `?` | Yes / No | `(y/n)`, `(yes/no)` |
| `↵` | Enter | `Press Enter`, `[Enter]` |
| `$` | Cost | `API Cost`, `tokens used` |
| `›` | Prompt | Prompt character (`❯`) |
| `✻` | Thinking | `✻ Thinking` / `Pondered` etc. |
| `⏺` | Running | Tool execution indicator |

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
