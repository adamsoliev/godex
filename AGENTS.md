# Repository Guidelines

## Project Structure & Module Organization
The Go module lives at the repo root (`go.mod`). The CLI entry point is `cmd/godex/main.go`, which drives history ingestion, LLM calls, and reporting. Supporting packages stay under `internal/` (currently `internal/history` for shell history access), and Makefile builds place binaries in `bin/`.

## Architecture
```mermaid
graph LR
    history["Today's shell history"] --> summary["LLM summarization & intent inference"]
    summary --> optimize["Workflow optimization ideas"]
    optimize --> web["Web-powered suggestions"]
```

### Upcoming Features
- Categorization call of whether each suggested intent needs further clarify
- If so, intent drill-down that resolves custom aliases and functions directly from shell configs after each summary.
- Else, historical usage inspection (e.g., `history | grep <command>`) to surface how often each highlighted intent appears in prior sessions.

## Build, Test, and Development Commands
`go run ./cmd/godex` executes the CLI without building. `make build` compiles to `bin/godex`, `make run` rebuilds then launches the binary, `make test` invokes `go test ./...`, `make fmt` runs `gofmt` on `cmd/` and `internal/`, and `make clean` removes `bin/`.

## Coding Style & Naming Conventions
Stay idiomatic Go: tabs for indentation, exported identifiers in PascalCase, unexported in camelCase, and package names matching directory names. Run `gofmt` (or `make fmt`) before submitting and keep comments minimal yet explanatory.

## Testing Guidelines
Write `_test.go` files beside the code they cover and prefer table-driven tests for variations. Ensure new logic passes `go test ./...`; stub external APIs when exercising OpenAI-dependent behavior to keep tests deterministic.

## Commit & Pull Request Guidelines
Commits in this repo use short, lowercase, imperative summaries (e.g., `cleaned up prompts`); follow suit and keep each commit focused. Pull requests should note the change scope, include test evidence (`go test ./...`), link issues when relevant, and flag any required configuration updates.

## Agent-Specific Notes
Avoid destructive operations on user dotfiles or unnecessary network calls. Prefer the Makefile targets for reproducible builds and formatting, and surface any permissions or sandbox limitations in discussion threads.
