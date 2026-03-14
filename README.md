# github-watch

A terminal UI dashboard for monitoring GitHub pull requests, CI runs, and API rate limits — all in one view.

```
┌─────────────────────────────────────────────────────────────────┐
│ github-watch    Refresh: 15s    [tab/h/l]pane [j/k]nav ...     │
├───────────────────────────┬─────────────────────────────────────┤
│ Pull Requests             │ CI Runs                             │
│ ─────────────────────     │ ─────────────────────               │
│ #123 Fix auth flow  ✓ ap │ ✓ build        main   success  2m   │
│ #122 Add caching    ◷ re │ ⟳ test         feat   in_prog  30s  │
│ #121 Update deps    - no │ ✗ lint         fix    failure  5m   │
├───────────────────────────┴─────────────────────────────────────┤
│ GitHub API Rate Limits                                          │
│   Core     ████████████████████░░░░  4500/5000   90%            │
│   GraphQL  ██████████████░░░░░░░░░░  3500/5000   70%            │
│   Search   ████████████████████████    30/30    100%            │
│   Resets in: 42m                                                │
└─────────────────────────────────────────────────────────────────┘
```

## Features

- **3-pane layout** — PRs (top-left), CI runs (top-right), rate limits (bottom)
- **Auto-refresh** — PR/CI every 15s, rate limits every 30s
- **Keyboard-driven** — vim-style navigation, open in browser with Enter
- **Zero config** — uses `gh` CLI for auth, works in any git repo
- **Lightweight** — single binary, no background processes

## Requirements

- [GitHub CLI (`gh`)](https://cli.github.com/) — authenticated via `gh auth login`
- Go 1.25+ (for building from source)

## Installation

```sh
go install github.com/0daryo/github-watch@latest
```

Or build from source:

```sh
git clone https://github.com/0daryo/github-watch.git
cd github-watch
go build -o github-watch .
```

## Usage

Run inside any Git repository that has a GitHub remote:

```sh
github-watch
```

### Keybindings

| Key | Action |
|-----|--------|
| `tab` / `h` / `l` / `←` / `→` | Switch focus between PR and CI panes |
| `j` / `k` / `↓` / `↑` | Navigate items in the focused pane |
| `enter` | Open selected PR/run in browser |
| `r` | Manual refresh |
| `q` / `Ctrl+C` | Quit |

## How It Works

github-watch uses the [GitHub CLI](https://cli.github.com/) under the hood:

- `gh pr list` — fetches open pull requests with review and check status
- `gh run list` — fetches recent CI workflow runs
- `gh api rate_limit` — fetches API rate limit information

No tokens are stored or managed by github-watch — authentication is fully delegated to `gh`.

## License

[MIT](LICENSE)
