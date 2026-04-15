# AGENTS.md — core/play

Primary references: `README.md`, `CLAUDE.md`, `docs/RFC.md`, and `docs/RFC.app.md`.

## Identity

- Module: `dappco.re/go/play`
- Product: STIM Game & Software Preservation runtime
- Purpose: run preserved software from deterministic, hash-verified STIM bundles
- Status: bootstrapped repo; spec-first, implementation pending

## Scope

### In scope

- STIM bundle format and manifest/schema work
- Runtime selection for emulator, compatibility, or native runners
- Verification chain, SBOM tracking, and deterministic extraction/execution
- CLI, HTTP, and MCP surfaces described in the RFC

### Out of scope

- Writing emulator implementations from scratch
- Game-specific porting work
- UI chrome owned by `core/gui`

## Repo map

- `docs/RFC.md` — primary product and command spec for `core play`
- `docs/RFC.app.md` — upstream keystone runtime dependency context
- `docs/CLAUDE.md` — broader product and architecture notes
- `README.md` — concise repo summary and current status
- `go.mod` — Go module definition

## Working rules

- Treat `docs/RFC.md` as the source of truth for command names, bundle shape, and runtime behaviour.
- Keep changes small and spec-aligned; if implementation diverges from the RFC, update the docs in the same change.
- Respect the dependency chain: `core/play` is a consumer of `core/app`, not a replacement for it.
- Prefer upstream emulator/runner integrations over bespoke implementations.
- Use UK English in user-facing and documentation text.

## Go conventions

- Prefer core primitives over standard library shortcuts.
- Banned imports: `fmt`, `errors`, `os`, `os/exec`, `strings`, `path/filepath`, `encoding/json`, `log`
- Use predictable, descriptive names.
- Add usage-example comments where they match surrounding patterns.
- Never edit `go.mod` manually; use `go get` or `go mod tidy`.

## Test conventions

- Use test names shaped like `TestFilename_Function_Good`, `TestFilename_Function_Bad`, and `TestFilename_Function_Ugly`.
- Keep one test file per source file.
- When adding a new source file, add the matching tests unless the repo is still purely spec/documentation.

## Implementation notes

- Preserve the STIM concepts from the RFC: deterministic bundles, hash verification, SBOM tracking, sandboxed execution, and engine registry selection.
- Keep bundle terminology consistent: artefact, engine, manifest, verification chain, save-state directory.
- Command surface should stay aligned with the RFC examples: `play`, `play/list`, `play/verify`, and `play/bundle`.
