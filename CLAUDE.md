# CLAUDE.md — core/play

Reference: `docs/RFC.md` + `docs/RFC.app.md` (core/app keystone dep).

## Identity

`dappco.re/go/play` — STIM Game & Software Preservation runtime. Wraps preserved software (ROMs, binaries, installers) in deterministic hash-verified bundles, runs them via appropriate emulator core / compat layer / native runner.

## Dependency chain

```
core/play → core/app (manifest/boot runtime) → go-scm (manifest/compile/sign)
         → core/gui (windowing)
         → core/go (primitives)
         → core/cli (command registration)
         → core/go/build (deterministic archive production)
```

core/app must land first — core/play is a consumer of the keystone boot flow.

## Scope boundary

- IN: STIM bundle format, manifest schema, runtime selection (emulator core / compat / native), verification chain, SBOM tracking, deterministic archive extraction, sandboxed execution
- OUT: game-specific porting work, emulator implementations (use upstream), UI chrome (core/gui owns)

## Core conventions

- Banned imports: fmt, errors, os, os/exec, strings, path/filepath, encoding/json, log → core primitives
- Tests: `TestFilename_Function_{Good,Bad,Ugly}` — one test file per source file, all three mandatory
- UK English, usage-example comments, predictable names
- Never manually edit `go.mod` — use `go get`, `go mod tidy`
