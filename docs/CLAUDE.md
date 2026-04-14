# CorePlay — STIM Game & Software Preservation

> Agent context summary for `plans/code/core/play/`. Games are the demo. Legacy enterprise is the product.

## What CorePlay Is

Runs preserved software inside STIM (Sandboxed Temporal Isolation Module) containers. A STIM bundle is a deterministic, hash-verified, SBOM-tracked archive containing the original artefact, a runtime engine, and a verification manifest. Also the **pipeline proof before CoreLEM** — simpler Apple review, proves CoreGUI + Xcode Cloud → App Store pipeline.

## Key Facts

- **Status**: Specced (389 lines), implementing
- **Repo**: `core/play`
- **Module**: `dappco.re/go/play`
- **Depends on**: core/app (manifest), core/gui (WebView2), core/go/build (deterministic archives), core/cli
- **Revenue priority**: Apple pivot — parallel path with CoreLEM

## STIM Bundle Structure

```
mega-lo-mania/
├── manifest.yaml       # metadata + verification
├── rom/                # original artefact
├── emulator.yaml       # runtime config (engine, platform, input, display, audio)
├── sbom.json           # CycloneDX SBOM
└── checksums.sha256    # deterministic hash chain
```

Directory name IS the game name. `core play mega-lo-mania` resolves from games directory.

## Architecture

- **Engine registry**: Interface pattern with build tags. Engines: DOSBox, RetroArch, ScummVM, native
- **Sandboxed**: No network (unless manifest permits), no writes outside save-state dir, process isolation via `c.Process()`, resource limits
- **CoreCommand integration**: CLI (`core play`), HTTP (`GET /play/{name}`), MCP (agent can test bundles), i18n
- **Save states**: `~/.core/play/{name}/saves/` and `screenshots/`
- **Shield integration**: Every STIM bundle is a Shield artefact (SBOM, code integrity, content hash, threat check)

## App Store Distribution (Apple Pivot)

### Three Tiers

1. **Arcade+ / Apple One subscribers: FREE** — StoreKit 2 entitlement check, full library, Family Sharing. Apple gets happier subscribers, we get zero-cost distribution
2. **Non-Arcade: PAID** — £2.99/mo or £24.99/yr or £49.99 lifetime. Apple gets 30%→15% (Small Business Program)
3. **Free (EUPL-1.2)** — self-compiled, BYOROM, no DRM

### STIM DRM via Borg

ROM encrypted as `.stim`, CDN-hosted, decrypted in-memory in WebView2 (never touches disk). Stream key derived from StoreKit entitlement token. Zero-trust — no backend validates.

### Xcode Cloud Pipeline

Local iterate → push to main → Xcode Cloud (build + test + archive + notarise) → TestFlight → App Store. 25 free compute hours/month.

### Platform Targets

macOS arm64 (MVP), iOS/iPadOS (post-MVP), watchOS (launcher), tvOS (AirPlay).

## Launch Catalogue (Freeware/Open Source)

Mega lo Mania (1991, Genesis), Command & Conquer (1995, DOS), Prince of Persia (1989, DOS), Cave Story (2004, native), Tyrian (1995, DOS), Beneath a Steel Sky (1994, ScummVM).

## Enterprise Use Case

Same STIM format for legacy COBOL systems (bank migration), old Win32 tools, deprecated monitoring agents. Preserved, sandboxed, verified.

## Critical Rules

- **Emulators are legal** — Apple policy change 2024. Only distribute games with rights (freeware, open source, licensed)
- **STIM bundles are deterministic** — SHA-256 every file, CycloneDX SBOM, reproducible zip from go-build
- **CorePlay proves the pipeline** — simpler than CoreLEM, same Xcode Cloud path. Ship this first
- **No streaming, no sync in free tier** — paid features only
- **EUPL-1.2 for the player, separate content licence for game library**

## Spec Index

| File | Scope |
|------|-------|
| [RFC.md](RFC.md) | **Full spec** — STIM bundles, engines, sandbox, Shield, App Store distribution, Borg DRM, Xcode Cloud (389 lines) |
