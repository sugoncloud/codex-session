# codex-sess

A small terminal UI for browsing and managing Codex sessions on the machine where they are stored.

[简体中文](README.zh-CN.md)

- Browse both archived and unarchived sessions.
- Search by title, UUID, or working directory.
- Archive, restore, and delete sessions with confirmation.
- Keep the selection near its previous position after deletion.
- Print session lists as a table or JSON for scripting.

The UI uses English labels and preserves session titles in their original language. The list renders cells on demand and limits title previews to 100 Unicode characters, keeping navigation independent of the full title length.

## Requirements

- Go **1.26 or later** to build.
- A terminal on the machine containing your Codex data.
- A Codex CLI supporting `archive`, `unarchive`, and `delete` for mutation operations (verified with `codex-cli 0.153.4`). Listing does not require the CLI.

SQLite support is built in using a pure Go driver. No `sqlite3` executable or CGO is required. The current CI runs on Linux.

## Build

```sh
make build
./bin/codex-sess
```

Without Make:

```sh
CGO_ENABLED=0 go build -buildvcs=false -trimpath -o bin/codex-sess .
./bin/codex-sess
```

To embed a version or install into your Go binary directory:

```sh
make build VERSION=0.1.8
make install VERSION=0.1.8
```

`--version` reports `dev` unless a version is supplied at build time. `-buildvcs=false` also allows builds from source archives and directories without valid Git metadata.

## Usage

```sh
# Interactive session manager
codex-sess

# Use a specific Codex data directory or CLI binary
codex-sess --home "$HOME/.codex"
codex-sess --codex /path/to/codex

# Print all sessions, including archived sessions
codex-sess --list

# Filter and search
codex-sess --list --status archived --search project
codex-sess --status active

# Export the full session metadata as JSON
codex-sess --json
```

The default data directory is `$CODEX_HOME`, or `~/.codex` when unset. The same directory is passed to Codex CLI operations. Run the tool as the user who owns the sessions. For SSH projects, run it on the remote machine.

`--status` accepts `all`, `active`, or `archived`. Here, `active` means **unarchived**, not currently running. The tool reads the selected user's indexed sessions; it does not aggregate other hosts, accounts, or cloud tasks. Table and JSON exports retain complete titles; the 100-character limit applies to the TUI list and delete confirmation.

## Keyboard controls

| Key | Action |
| --- | --- |
| Up / Down or j / k | Select a session |
| PageUp / PageDown | Scroll by page |
| / | Focus search |
| Enter / Esc in search | Return to the list |
| Tab in the list | Cycle All / Unarchived / Archived |
| a | Archive selected session |
| u | Restore selected session |
| d | Permanently delete selected session |
| r | Refresh |
| q / Ctrl+C | Quit |

Mouse selection is supported. Deletion displays the UUID and the first 100 title characters, with **Cancel** selected by default. Choose **Delete permanently** to proceed; no typed confirmation is required. After removal, selection moves to the following row, or the preceding row if the last row was removed.

The UI remains navigable while an operation runs. Only additional mutations are blocked. Operations have a 30-second timeout and are followed by a refresh. If an operation times out, check the refreshed state before retrying.

## Storage and compatibility

The highest-numbered `state_*.sqlite` file in the selected data directory is opened **read-only**. Changes are delegated to Codex CLI with the session UUID as a separate process argument. The tool never writes SQL updates or moves session files itself.

Codex's database schema is internal and may change. This implementation reads `id`, `title`, `cwd`, `archived`, and `updated_at` from `threads`; incompatible schemas produce a read error. Transcript files not indexed in that database are not listed.

Deletion is permanent. Archiving or deleting a running session can affect that session's execution.

## Development

```sh
make fmt       # Format Go source
make check     # Check formatting, run tests, and run go vet
make bench     # Run rendering/navigation benchmarks
make build     # Build a standalone binary
```

Tests use temporary databases, a fake Codex executable, and a simulated terminal. They do not mutate real Codex sessions. GitHub Actions runs the checks and a CGO-free build using Go 1.26.x on pushes and pull requests.

| File | Responsibility |
| --- | --- |
| `main.go` | Process entry and version |
| `cli.go` | Flag parsing, startup, and list/JSON output |
| `ui.go` | TUI navigation, confirmations, and background operations |
| `table.go` | Lazy table cells and bounded previews |
| `display.go` | Display escaping and status labels |
| `session.go` | Session model, filtering, and timeouts |
| `store.go` | Database discovery and read-only access |
| `runner.go` | Validated Codex CLI invocation |

The module path is `github.com/sugoncloud/codex-session`. To install the executable with its `codex-sess` name, clone the repository and run `make install` (or download a release archive). The repository module name and the executable name are intentionally different.

## Releases

Push a version tag to trigger `.github/workflows/release.yml`:

```sh
git tag -a v0.1.8 -m "Release v0.1.8"
git push origin v0.1.8
```

GoReleaser publishes Linux, macOS, and Windows archives for amd64 and arm64. The workflow also publishes checksums and uses the repository's `GITHUB_TOKEN`.
