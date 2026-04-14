---
module: core/play
repo: core/play
lang: multi
tier: consumer
depends:
  - code/core/app
  - code/core/gui
tags:
  - games
  - stim
  - preservation
  - emulation
  - bundles
---

# Core Play RFC — STIM Game & Software Preservation

> `core play` — run preserved software in deterministic STIM bundles.
> Games are the demo. Legacy enterprise is the product.

**Module:** `dappco.re/go/play`
**Repository:** `core/play`
**Dependencies:** `core/go` (primitives), `core/cli` (command registration), `core/go/build` (deterministic archives)

---

## 1. Overview

Core Play runs preserved software inside STIM (Sandboxed Temporal Isolation Module) containers. A STIM bundle is a deterministic, hash-verified, SBOM-tracked archive containing:

1. The original artefact (ROM, binary, installer)
2. A runtime engine (emulator core, compatibility layer, or native runner)
3. A manifest describing inputs, outputs, and verification chain

```bash
core play                          # read manifest.yaml from cwd
core play mega-lo-mania            # from games/ parent dir
core play command-and-conquer      # same
```

Directory name IS the game name. Same pattern as `core build` reads `build.yaml`.

---

## 2. STIM Bundle Structure

```
mega-lo-mania/
├── manifest.yaml          # bundle metadata + verification
├── rom/
│   └── MegaLoMania.zip    # original artefact (542 kB)
├── emulator.yaml          # runtime configuration
├── sbom.json              # CycloneDX SBOM
└── checksums.sha256       # deterministic hash chain
```

### 2.1 manifest.yaml

```yaml
name: mega-lo-mania
title: "Mega lo Mania"
author: "Sensible Software"
year: 1991
platform: sega-genesis
genre: strategy

artefact:
  path: rom/MegaLoMania.zip
  sha256: "..."
  size: 542kB
  source: "freeware — distributed by retrogames.cz, GOG, Steam"

runtime:
  engine: kega-fusion        # or: dosbox, scummvm, retroarch
  config: emulator.yaml

licence: freeware
preservation:
  verified: true
  chain: checksums.sha256
```

### 2.2 emulator.yaml

```yaml
engine: kega-fusion
platform: sega-genesis
input:
  type: gamepad
  mapping: default-genesis
display:
  scale: 3x
  filter: nearest            # or: crt, scanline
audio:
  enabled: true
  sample_rate: 44100
```

---

## 3. CoreCommand Integration

As a CoreCommand, `play` maps to all surfaces:

| Surface | Path | Usage |
|---------|------|-------|
| CLI | `core play {name}` | Run from terminal |
| HTTP | `GET /play/{name}` | Launch via API |
| MCP | `play` tool | AI agent can test bundles |
| i18n | `play.*` | Localised strings |

### 3.1 Command Registration

```go
func Register(c *core.Core) {
    c.Command("play", core.Command{
        Description: "Run preserved software in a STIM bundle",
        Action:      cmdPlay,
    })
    c.Command("play/list", core.Command{
        Description: "List available STIM bundles",
        Action:      cmdPlayList,
    })
    c.Command("play/verify", core.Command{
        Description: "Verify hash chain without running",
        Action:      cmdPlayVerify,
    })
    c.Command("play/bundle", core.Command{
        Description: "Create a STIM bundle from artefact",
        Action:      cmdPlayBundle,
    })
}
```

### 3.2 CLI Usage

```bash
core play                              # cwd has manifest.yaml
core play mega-lo-mania                # resolve from games dir
core play --list                       # list available bundles
core play --verify mega-lo-mania       # verify hash chain without running
core play --info mega-lo-mania         # show manifest + SBOM
```

---

## 4. STIM Runtime

### 4.1 Isolation

STIM bundles run in sandboxed containers:

- No network access (unless manifest explicitly permits)
- No filesystem writes outside save-state directory
- Process isolation via Core's process primitives (`c.Process()`)
- Resource limits (CPU, memory) from manifest

### 4.2 Engine Registry

```go
type Engine interface {
    // Name returns the engine identifier (e.g. "dosbox", "retroarch")
    Name() string
    // Platforms returns supported platforms (e.g. ["sega-genesis", "dos"])
    Platforms() []string
    // Run executes the artefact with the given config
    Run(artefact string, config EngineConfig) error
    // Verify checks the engine binary integrity
    Verify() error
}
```

Engines register via `init()` with build tags — same pattern as CLI variants:

```go
//go:build engine_dosbox

func init() {
    play.RegisterEngine(&DOSBoxEngine{})
}
```

### 4.3 Save States

```
~/.core/play/
├── mega-lo-mania/
│   ├── saves/             # save states
│   └── screenshots/       # auto-captured
└── command-and-conquer/
    └── saves/
```

`.core/` directory convention — same as config, store, agent workspace.

---

## 5. Shield Integration

Every STIM bundle is a Shield artefact:

| Shield Surface | What It Checks |
|---------------|----------------|
| SBOM | Full dependency chain (ROM + engine + config) |
| Code | Engine binary integrity |
| Content | ROM hash matches original release |
| Threat | No injected payloads in artefact |

The marketing: "This binary hasn't changed since 1991. Here's the SBOM. Here's the hash chain. Shield verified: 0 modifications."

---

## 6. Bundle Creation

```bash
core play bundle --name mega-lo-mania \
    --rom MegaLoMania.zip \
    --engine kega-fusion \
    --platform sega-genesis \
    --year 1991 \
    --author "Sensible Software"
```

Produces a deterministic STIM bundle with:
- SHA-256 checksums for every file
- CycloneDX SBOM
- manifest.yaml
- Deterministic zip (from go-build)

---

## 7. Catalogue

### 7.1 Launch Titles (Freeware / Open Source)

| Game | Year | Platform | Engine | Size | Licence |
|------|------|----------|--------|------|---------|
| Mega lo Mania | 1991 | Sega Genesis | kega-fusion | 542 kB | Freeware |
| Command & Conquer | 1995 | DOS | dosbox | ~50 MB | EA Freeware |
| Prince of Persia | 1989 | DOS | dosbox | ~1 MB | Open Source |
| Cave Story | 2004 | Native | native | ~3 MB | Freeware |
| Tyrian | 1995 | DOS | dosbox | ~8 MB | Freeware |
| Beneath a Steel Sky | 1994 | ScummVM | scummvm | ~70 MB | Freeware |

### 7.2 Enterprise Use Case

Same STIM bundle format, same verification, same isolation:

| Software | Use Case |
|----------|----------|
| Legacy COBOL system | Bank migration bridge |
| Old Win32 internal tool | Preserved with TransformerIn/Out CLI compat |
| Deprecated monitoring agent | Sandboxed execution until replacement ships |

---

## 8. Implementation Priority

1. `manifest.yaml` schema and parser
2. `core play` command registration
3. Engine interface and registry
4. DOSBox engine adapter
5. RetroArch engine adapter (covers Genesis, SNES, etc.)
6. STIM sandbox (process isolation, no-network, save states)
7. `core play bundle` creation command
8. Shield integration (SBOM, hash verification)
9. ScummVM engine adapter
10. Catalogue index and `--list` command

---

---

## 9. App Store Distribution

### 9.1 The Apple Play

CorePlay is the **pipeline proof** before CoreLEM. Simpler review (no AI, no Metal compute), proves CoreGUI rendering, input, audio, and the full Xcode Cloud → TestFlight → App Store pipeline.

### 9.2 Business Model — Three Tiers

**Tier 1: Apple Arcade+ / Apple One Subscribers (FREE)**
- CorePlay detects active Arcade+ or Apple One via StoreKit 2 (`Transaction.currentEntitlements`)
- Full access to entire game library, gratis
- Family Sharing (up to 5 people) for the duration of their Apple subscription
- If they cancel Apple One/Arcade+, they can purchase CorePlay directly (Tier 2)
- Apple gets: happier subscribers who see MORE value in Arcade+
- We get: distribution to Apple's entire subscriber base, zero acquisition cost

**Tier 2: Non-Arcade Subscribers (PAID)**
- £2.99/mo or £24.99/year or £49.99 lifetime
- Apple gets 30% (year 1) → 15% (year 2+, Small Business Program)
- Same full library access
- If they later get Arcade+, auto-switches to free tier

**Tier 3: Free (source compile, EUPL-1.2)**
- Same player binary, self-compiled
- ROMs not included (user provides their own)
- No DRM, no streaming, no sync
- Works perfectly, just manual

### 9.3 STIM DRM via Borg

Same system as dapp.fm music streaming:
- ROM encrypted as `.stim` (Sandboxed Temporal Isolation Module)
- CDN-hosted (Garage S3 or Apple CloudKit)
- Decrypted in-memory in WebView2 (never touches disk)
- Stream key derived from active subscription entitlement
- Zero-trust: no backend server validates — the key IS the entitlement
- Browser-based: works in WebView2, works in Safari, works anywhere CoreTS runs

```
User opens game
  → StoreKit 2 checks entitlement (Arcade+ OR CorePlay sub)
  → Derives stream key from entitlement token
  → Fetches .stim from CDN
  → Decrypts in WebView2 memory
  → Emulator loads ROM from ArrayBuffer
  → Game plays
```

### 9.4 The Pitch to Apple

> "We're giving ALL your Apple One and Arcade+ subscribers access to our
> entire retro game library for free. Family Sharing included. We're a
> UK Community Interest Company — we chose to. You get 30-50% of
> non-Arcade revenue. Your customers get more value from their existing
> subscription. 100% in your ecosystem — Xcode Cloud built, signed,
> notarised. Everyone wins."

### 9.5 Why Apple Cares

1. More value for Arcade+ subscribers → less churn → more revenue
2. Retro gaming is nostalgia-driven → high engagement, low support cost
3. 100% in Apple's ecosystem — no external servers, no sideloading
4. Signed, notarised, Xcode Cloud built — clean provenance
5. CIC structure = not extractive, cooperative by charter
6. Same tech (CoreGUI + MLX) leads to CoreLEM → Apple Intelligence ecosystem
7. Display tech (STIM isolation) could interest Apple engineers

### 9.6 Xcode Cloud Pipeline

```
Local: iterate with 4-agent IDE (Claude + Codex + Lemrd + core-agent)
Push to main
  → Xcode Cloud: Build macOS + iOS → Test (parallel) → Archive → Notarise
  → TestFlight: Snider plays Mega-Lo-Mania on iPhone (first test)
  → App Store: Submit when stable
```

25 free compute hours/month covers the build. Revenue covers upgrades.

### 9.7 Platform Targets

| Platform | Engine | Priority |
|----------|--------|----------|
| macOS (arm64) | Native WebView2 | MVP |
| iOS / iPadOS | WKWebView + touch input | Post-MVP |
| watchOS | Game launcher / remote | Future |
| tvOS | AirPlay + controller input | Future |

### 9.8 Legal

- Emulators are legal (Apple allows them as of 2024 policy change)
- ROMs: only distribute games we have rights to (freeware, open source)
- Partner with rights holders where needed
- User-provided ROMs: supported in free tier (BYOROM)
- EUPL-1.2 for the player, separate content licence for game library

---

## 10. Cross-References

- project/lthn/desktop/RFC.md — CoreGUI shell
- project/lthn/desktop/RFC.xcode-pipeline.md — Xcode Cloud pipeline
- code/core/gui/RFC.md — WebView2 + CoreTS preload
- code/core/ts/RFC.md — Global scope control, storage polyfills
- code/core/go/io/RFC.md — Borg .stim, io.Medium
- code/core/go/store/RFC.md — Save states, SQLite KV
- code/core/lem/RFC.md — Same business model template, same pipeline
- rfc/snider/RFC-BORG-*.md — Borg DRM, .stim format

---

## Changelog

- 2026-04-08: Added §9 App Store distribution — Arcade+ free tier, Borg .stim DRM, StoreKit 2, Xcode Cloud pipeline, platform targets, legal. CorePlay proves pipeline before CoreLEM ships.
- 2026-04-01: Initial RFC — STIM bundles, engine registry, CoreCommand integration, Shield verification, launch catalogue. "Games are the demo. Legacy enterprise is the product."
