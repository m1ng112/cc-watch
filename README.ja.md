# cc-watch

tmux 上で動いている **Claude Code** セッションのうち、**ユーザー入力を待っているもの**をスキャンして一覧表示し、選択するとそのペインにジャンプする TUI ツールです。

3秒ごとに自動リフレッシュし、待機セッションの変化をリアルタイムに把握できます。

**[English version](README.md)**

## スクリーンショット

```
┌── cc-watch  3 pane(s) waiting ──────────────────────────────────┐
│                                │                                │
│ ▸ (1) !  確認待ち    0:3: claude│  [プレビュー: 選択中の           │
│   (2) ›  入力待ち    0:1: claude│   ペインの末尾30行を             │
│   (3) ?  Yes / No   0:5: claude│   ANSI カラー付きで表示]         │
│                                │                                │
├────────────────────────────────┴────────────────────────────────┤
│ 14:32:05  ↑↓ 移動  enter ジャンプ  q 終了  r 更新               │
└─────────────────────────────────────────────────────────────────┘
```

## 検出する状態

| アイコン | ラベル | 意味 |
|---------|--------|------|
| `✓` | 承認待ち | プランモードの承認待ち（ExitPlanMode） |
| `!` | 確認待ち | コマンド／ツール実行の許可待ち |
| `▸` | 回答待ち | 選択式の質問でブロック中（AskUserQuestion など） |
| `?` | Yes / No | `(y/n)` 形式のプロンプト |
| `↵` | Enter待ち | `Press Enter` など |
| `$` | コスト確認 | `API Cost` など |
| `›` | 指示待ち | 応答が完了し、次の指示を待っている（空プロンプト） |
| `⏺` | 実行中 | 生成中（スピナー） |
| `·` | 動作中 | 判定できなかった場合のフォールバック |

## 検出方式: hooks（推奨）とスクレイピング

cc-watch は 2 通りの方法で状態を判定します。

1. **Claude Code hooks（推奨・確実）** — `cc-watch install-hooks` を実行すると、`~/.claude/settings.json` に hooks が登録され、各 Claude セッションが状態遷移（実行開始 / 完了 / 承認・確認・質問待ち）を `$TMUX_PANE` 付きで直接 cc-watch に通知します。端末レイアウトに依存せず正確です。
2. **端末スクレイピング（フォールバック）** — hooks が未登録のペインや tmux 外のセッションでは、ペインの末尾を正規表現で解析して状態を推定します。

```sh
# hooks を登録（既存設定はバックアップされ、他の hooks は保持されます）
cc-watch install-hooks
```

> 登録後、既存の Claude Code セッションには反映されません。各セッションを再起動するか `/hooks` で確認してください。tmux 内で起動した Claude のみ対応付け可能です。

## ステータスバーに常時表示（cc-watch status）

常時サイドバー（cmux 風）は画面を占有して邪魔、毎回 popup を出すのも手間 — その中間として、tmux のステータスバーに**注意が必要なペインだけ**を常時1行で表示できます。

```tmux
# ~/.config/tmux/tmux.conf
set -g status-interval 5            # 更新間隔（秒）
set -g status-right-length 100
# 既存の status-right の先頭に #(cc-watch status) を追加
set -g status-right "#(cc-watch status) ...既存の内容..."
```

- `cc-watch status` は承認待ち／確認待ち／回答待ち／Yes・No／Enter待ち／コスト確認のペインを「アイコン `session:window`」で出力します（**実行中・指示待ちは出さない**、0件なら何も出さない）
- 多い場合は先頭数件＋`+N` に集約します
- フルの一覧表示とジャンプは従来どおり popup（例: `prefix + o`）で
- `#()` は tmux の shell で実行されるため PATH に無ければ絶対パスを指定: `#(/path/to/cc-watch status)`
- 正確な状態判定には `install-hooks` の併用を推奨（スクレイピングはペイン内容に左右される場合があります）

## 必要なもの

- **tmux** — セッション管理
- **Go 1.23+** — ビルド

## インストール

```sh
go install github.com/m1ng112/cc-watch@latest
```

または手動ビルド：

```sh
git clone https://github.com/m1ng112/cc-watch.git
cd cc-watch
go build -o cc-watch .
ln -s "$(pwd)/cc-watch" ~/.local/bin/cc-watch
```

## 使い方

tmux セッション内で実行します。

```sh
cc-watch
```

### キーバインド

| キー | 動作 |
|------|------|
| `↑` / `k` | カーソル上 |
| `↓` / `j` | カーソル下 |
| `Enter` | 選択ペインにジャンプ |
| `r` | 即時リスキャン |
| `q` / `Esc` / `Ctrl+C` | 終了 |

### 自動リフレッシュ

3秒ごとに自動でペインをスキャンし、一覧を更新します。待機中のセッションがない場合は「✓ 待機なし」と表示されます。

### tmux キーバインド例

```sh
# prefix + w で cc-watch を起動
bind-key w run-shell -b 'cc-watch'
```

## ライセンス

MIT
