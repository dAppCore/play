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

> `core play` runs preserved software from deterministic STIM bundles.
> Games are the demo. Legacy enterprise is the product.

**Module:** `dappco.re/go/play`  
**Repository:** `core/play`  
**Status:** Draft, spec-first  
**Primary dependency:** `core/app` supplies the manifest and boot contract

---

## 1. Purpose

Core Play is the runtime for preserved software packaged as **STIM bundles**.
A STIM bundle combines:

1. the original artefact
2. the runtime needed to execute it
3. the metadata needed to verify, launch, and preserve it

The immediate product shape is a games runtime. The broader product shape is a
general preservation and controlled execution system for legacy software.

### 1.1 Position in Core

Core Play is not a general-purpose application framework. It sits on top of the
Core stack and consumes existing runtime primitives.

- `core/app` defines manifest discovery and boot flow
- `core/gui` owns windows, rendering surfaces, and platform shell concerns
- `core/go/build` produces deterministic bundle artefacts
- `core/cli` exposes command registration and command routing
- Shield-style verification surfaces provide integrity, SBOM, and threat checks

### 1.2 Goals

- Run preserved software from a deterministic, verifiable bundle
- Separate the **artefact** from the **runtime engine** cleanly
- Keep execution sandboxed and auditable
- Make bundle verification possible without launch
- Support both curated catalogue distribution and local BYOROM workflows
- Reuse the same bundle contract for games and enterprise preservation

### 1.3 Non-goals

- Writing emulator engines from scratch
- Replacing `core/app` or `core/gui`
- Doing per-title game porting work inside this module
- Defining commercial rights for individual games inside the runtime itself
- Solving cloud streaming as a requirement for local playback

---

## 2. Terms

### 2.1 STIM

**STIM** means **Sandboxed Temporal Isolation Module**.

A STIM bundle is a self-describing directory or archive that preserves:

- the original software artefact
- the engine or runner required to execute it
- the verification chain needed to prove integrity

### 2.2 Artefact

The original software payload:

- ROM
- binary
- installer
- data files
- application package

### 2.3 Engine

The runtime used to execute the artefact:

- emulator core
- compatibility layer
- native runner

Examples include DOSBox, DOSBox-X, ScummVM, RetroArch-based adapters, or a
native wrapper.

### 2.4 Bundle

The on-disk shape used by `core play`. A bundle can be:

- a directory in development
- a deterministic archive for distribution

### 2.5 Catalogue

A known index of playable or installable bundles, typically rights-cleared,
curated, and pre-verified.

### 2.6 BYOROM

**Bring Your Own ROM** or, more generally, **Bring Your Own Artefact**.
The user supplies the original software. Core Play supplies verification,
packaging, and runtime selection where possible.

---

## 3. Product framing

Core Play has three jobs at once:

1. preserve software
2. run software safely
3. prove the Core distribution pipeline on a simpler product than CoreLEM

The games-facing product is important because it is visible, demoable, and
emotionally legible. The same machinery also fits long-tail enterprise cases:

- preserved COBOL workloads
- old Win32 operational tools
- legacy monitoring agents
- internal software that must survive platform churn

This RFC therefore treats **games as the initial catalogue**, not as the only
valid content type.

---

## 4. Bundle contract

Every STIM bundle has a deterministic and inspectable structure.

```text
mega-lo-mania/
├── manifest.yaml
├── rom/
│   └── MegaLoMania.zip
├── emulator.yaml
├── sbom.json
└── checksums.sha256
```

The directory name is the bundle code and defaults to the runnable name.

### 4.1 Required files

| File | Required | Purpose |
|------|----------|---------|
| `manifest.yaml` | Yes | Identity, artefact metadata, runtime binding, permissions |
| `rom/` | Yes | Original artefact payload |
| `emulator.yaml` | Yes | Engine-specific launch configuration |
| `sbom.json` | Yes | CycloneDX or equivalent bundle inventory |
| `checksums.sha256` | Yes | Hash chain for deterministic verification |

### 4.2 Optional files

| File | Purpose |
|------|---------|
| `cover.png` | Catalogue artwork |
| `licence.txt` | Rights or redistribution text |
| `notes.md` | Preservation notes and provenance |
| `patches/` | Optional compatibility patches declared in the manifest |
| `saves/seed/` | Optional starter save state or test fixtures |

### 4.3 `manifest.yaml`

The manifest is the bundle contract. It describes what the bundle is, what it
contains, what engine it expects, what permissions it needs, and how integrity
is proven.

```yaml
format_version: stim-v1
name: mega-lo-mania
title: "Mega lo Mania"
author: "Sensible Software"
year: 1991
platform: sega-genesis
genre: strategy
licence: freeware

artefact:
  path: rom/MegaLoMania.zip
  sha256: "9f0f..."
  size: 554192
  media_type: application/zip
  source: "Rights-cleared redistribution"

runtime:
  engine: retroarch
  profile: genesis
  config: emulator.yaml
  entrypoint: rom/MegaLoMania.zip
  acceleration: auto
  filter: nearest

verification:
  chain: checksums.sha256
  sbom: sbom.json
  deterministic: true

permissions:
  network: false
  microphone: false
  filesystem:
    read:
      - rom/
    write:
      - saves/
      - screenshots/

resources:
  cpu_percent: 75
  memory_bytes: 268435456

save:
  path: saves/
  screenshots: screenshots/

distribution:
  mode: catalogue
  byorom: false
```

### 4.4 Manifest fields

| Field | Required | Meaning |
|------|----------|---------|
| `format_version` | Yes | STIM manifest schema version |
| `name` | Yes | Stable machine identifier |
| `title` | Yes | Human-readable title |
| `author` | No | Original developer or studio |
| `year` | No | Original release year |
| `platform` | Yes | Target platform of the artefact |
| `genre` | No | Catalogue metadata |
| `licence` | Yes | Rights model for bundle distribution |
| `artefact` | Yes | Original payload metadata |
| `runtime` | Yes | Engine binding and launch entry |
| `runtime.acceleration` | No | Preferred acceleration policy: `off`, `auto`, or `required` |
| `runtime.filter` | No | Preferred display filter such as `none`, `nearest`, `bilinear`, `scanline`, or `crt` |
| `verification` | Yes | Integrity chain declaration |
| `permissions` | Yes | Sandbox and runtime capability declaration |
| `resources` | No | CPU and memory ceilings for sandboxed execution |
| `save` | No | Save-state and screenshot layout |
| `distribution` | No | Delivery-mode hints |

The current supported manifest format is `stim-v1`. Manifests without
`format_version` are treated as legacy RFC-era bundles and are normalised to
`stim-v1` at load time. Unknown future format versions must be rejected rather
than silently interpreted under the wrong schema.

### 4.5 `emulator.yaml`

`emulator.yaml` carries engine-specific launch details which do not belong in
the top-level manifest.

```yaml
engine: retroarch
profile: genesis

input:
  type: gamepad
  mapping: default-genesis

display:
  scale: 3x
  acceleration: auto
  filter: nearest
  aspect: original

audio:
  enabled: true
  sample_rate: 44100

performance:
  rewind: false
  fast_forward: true
```

### 4.6 Hash chain

`checksums.sha256` records hashes for every material file in the bundle.

At minimum:

- `manifest.yaml`
- `emulator.yaml`
- every file in `rom/`
- `sbom.json`
- any declared patch or auxiliary content

The hash list must be stable under deterministic rebuilds.

The verification chain is treated as a coverage contract:

- paths must be canonical relative bundle paths
- duplicate path entries are invalid
- required files must appear in the chain
- material files not recorded in the chain are reported as unverified content
- the chain file itself is exempt from self-hashing

### 4.7 SBOM

`sbom.json` represents the bundle inventory, not only the engine binary.
It should include:

- artefact payload
- engine package or engine identifier
- launch configuration
- applied patches
- build provenance where available

CycloneDX is the preferred initial format.

---

## 5. Resolution and command surfaces

Core Play is a CoreCommand and should map cleanly to CLI, HTTP, MCP, and
localisation surfaces.

| Surface | Path | Purpose |
|---------|------|---------|
| CLI | `core play {name}` | Run a bundle |
| CLI | `core play/list` | List available bundles |
| CLI | `core play/verify {name}` | Verify without launching |
| CLI | `core play/bundle` | Create a bundle |
| HTTP | `GET /play/{name}` | Launch or request launch |
| MCP | `play` | Agent-facing bundle interaction |
| i18n | `play.*` | User-facing strings |

### 5.1 Resolution rules

`core play` resolves bundles in the following order:

1. current working directory if `manifest.yaml` exists
2. named child directory in a configured catalogue root
3. installed local bundle index
4. explicit path argument if provided

Examples:

```bash
core play
core play mega-lo-mania
core play ./bundles/mega-lo-mania
core play/list
core play/verify mega-lo-mania
```

### 5.2 Command registration shape

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
        Description: "Verify a bundle without running it",
        Action:      cmdPlayVerify,
    })
    c.Command("play/bundle", core.Command{
        Description: "Create a deterministic STIM bundle",
        Action:      cmdPlayBundle,
    })
}
```

### 5.3 Exit expectations

| Command | Success | Failure |
|---------|---------|---------|
| `play` | Launches or hands off to launcher | Verification, resolution, or engine error |
| `play/list` | Returns catalogue entries | Index error |
| `play/verify` | Reports verified status | Hash, manifest, or SBOM mismatch |
| `play/bundle` | Produces deterministic bundle | Invalid inputs or non-deterministic output |

---

## 6. Runtime architecture

### 6.1 Boot sequence

The runtime should follow a strict boot order:

1. discover bundle
2. load `manifest.yaml`
3. verify hash chain and bundle structure
4. resolve engine adapter
5. prepare sandbox
6. mount save-state directories
7. launch engine with declared config
8. capture outcome, logs, and optional screenshots

### 6.2 Engine registry

Engine support is provided through adapters, not hard-coded branches.

```go
type Engine interface {
    Name() string
    Platforms() []string
    Verify() error
    Launch(bundle Bundle, cfg Config) error
}
```

Registration should use build tags where that keeps platform-specific or
licence-sensitive adapters separate.

```go
//go:build engine_dosbox

func init() {
    play.RegisterEngine(&DOSBoxEngine{})
}
```

Initial adapter coverage:

| Engine | Platforms | Notes |
|--------|-----------|-------|
| `dosbox` | DOS | Simple DOS artefact launch |
| `dosbox-x` | DOS, PC-98, Windows 3.x, Windows 9x | Machine/profile-aware boot planning |
| `retroarch` | Genesis, SNES, NES, Game Boy families | Libretro core selection through `runtime.profile` |
| `scummvm` | ScummVM and point-and-click bundles | Game ID supplied through `runtime.profile` |
| `mame` | Arcade, Neo Geo | Driver/profile launch with ROM directory isolation |
| `vice` | C64, C128, VIC-20 | Autostart launch for Commodore disk/tape artefacts |
| `fuse` | ZX Spectrum 48K/128K | Machine profile mapped to FUSE machine selection |
| `snes9x` | SNES, Super Nintendo | Standalone SNES runner where RetroArch is not desired |

### 6.3 Engine selection

Engine selection should be explicit first, heuristic second.

Order:

1. manifest `runtime.engine`
2. manifest `runtime.profile`
3. platform match
4. configured default adapter

If no valid adapter exists, verification may still succeed, but launch fails
with a precise engine-resolution error.

### 6.4 Sandbox policy

Default runtime policy:

- no outbound network
- no writes outside the declared save-state root
- no arbitrary process spawning from the bundle
- resource ceilings set from manifest defaults or platform policy
- engine binaries must pass integrity checks before launch

Manifest permissions can narrow access further. They should not silently widen
host access beyond platform policy.

Before launch, the resolved engine plan must be checked against the prepared
sandbox policy. A launch plan that requests network access, read paths, runtime
config access, resource ceilings, or write paths outside the manifest-derived
allowlist is rejected before process start. Required runtime files such as the
artefact, entrypoint, and `emulator.yaml` are folded into the effective read
allowlist so adapters cannot rely on undeclared bundle reads.

### 6.5 Save-state layout

```text
~/.core/play/
├── mega-lo-mania/
│   ├── saves/
│   ├── screenshots/
│   └── session.log
└── command-and-conquer/
    ├── saves/
    └── screenshots/
```

The runtime owns these directories. Bundle content should be treated as
read-only after verification.

### 6.6 Observability

The runtime should capture:

- engine selected
- verification result
- launch duration
- exit code or failure class
- save-state path
- optional screenshot artefacts for catalogue QA

---

## 7. Verification and Shield

Every STIM bundle is also a verification subject.

| Surface | Purpose |
|---------|---------|
| SBOM | Describe bundle inventory and provenance |
| Code | Verify engine package or adapter integrity |
| Content | Verify original artefact hash |
| Threat | Detect unexpected payloads or tampering |

### 7.1 Verification guarantees

`core play/verify` should confirm:

- required files exist
- manifest parses and is internally coherent
- every declared file hash matches
- the SBOM file exists and matches declared location
- the engine named by the manifest is known or marked unresolved
- ZIP artefacts reject unsafe paths, executable/script payloads, excessive
  expansion, oversized entries, oversized aggregate expansion, and excessive
  path nesting before launch

### 7.2 Deterministic bundle expectations

The same inputs should produce the same bundle outputs, including:

- file names
- file ordering
- timestamps normalised where supported
- checksum file contents
- archive layout

This matters for preservation and for supply-chain trust.

### 7.3 Failure classes

Verification failures should be categorised clearly:

- `bundle/not-found`
- `bundle/invalid-structure`
- `manifest/invalid`
- `hash/mismatch`
- `sbom/missing`
- `threat/entry-size`
- `threat/path-depth`
- `engine/unavailable`
- `sandbox/policy-denied`

These codes should be shared across CLI, HTTP, and MCP where practical.

---

## 8. Bundle creation

`core play/bundle` turns a raw artefact into a runnable STIM bundle.

### 8.1 Inputs

Typical inputs:

- bundle name
- title and metadata
- source artefact
- platform
- engine
- runtime profile
- rights and redistribution mode

Example:

```bash
core play/bundle \
  --name mega-lo-mania \
  --title "Mega lo Mania" \
  --rom MegaLoMania.zip \
  --platform sega-genesis \
  --engine retroarch \
  --profile genesis \
  --licence freeware
```

### 8.2 Output

The command should produce:

- `manifest.yaml`
- `emulator.yaml`
- `checksums.sha256`
- `sbom.json`
- deterministic archive output if requested

### 8.3 Workflow

1. inspect input artefact
2. assign runtime adapter
3. emit manifest and runtime config
4. generate checksum chain
5. generate SBOM
6. verify the just-built bundle
7. optionally archive deterministically

### 8.4 BYOROM mode

For BYOROM or enterprise ingestion, bundle creation may omit any curated
catalogue metadata not required for execution.

That still does not relax:

- checksum generation
- manifest integrity
- sandbox declaration
- engine resolution

---

## 9. Distribution profiles

Core Play should support multiple distribution shapes without requiring the
runtime contract to fork.

### 9.1 Local free tier

- self-compiled player
- user-supplied artefacts
- no remote entitlement requirement
- full local verification path

This is the baseline preservation mode.

### 9.2 Curated catalogue

- rights-cleared bundles
- prebuilt verification metadata
- artwork and catalogue metadata
- consistent engine packaging

This is the likely default consumer experience.

### 9.3 Proposed Apple distribution profile

Core Play is also positioned as a pipeline proof before CoreLEM:

- simpler review surface than a full AI or compute-heavy product
- validates CoreGUI rendering, input, and packaging path
- exercises build, test, archive, and signing flow

This section is a **product proposal**, not a platform guarantee. Any platform
distribution plan must be revalidated against current policy and current rights.

Proposed commercial tiers:

| Tier | Shape | Notes |
|------|-------|-------|
| 1 | Platform-subsidy tier | Free to eligible subscribers if commercial terms exist |
| 2 | Direct paid tier | Monthly, yearly, or lifetime entitlement |
| 3 | Free source tier | Self-compiled, BYOROM, no catalogue entitlement |

### 9.4 Protected asset mode

Some catalogue bundles may use a protected artefact format, such as a `.stim`
payload delivered separately from the open player.

If used, the design constraints are:

- decryption should occur in memory
- protected content should not be required for free-tier BYOROM support
- entitlement checks must be separable from bundle execution logic
- the open runtime must remain functional without protected catalogue assets

This keeps preservation mode intact while allowing a commercial catalogue layer.

---

## 10. Initial catalogue and enterprise extension

### 10.1 Launch catalogue candidates

| Title | Year | Platform | Engine | Rights model |
|------|------|----------|--------|--------------|
| Mega lo Mania | 1991 | Sega Genesis | RetroArch profile | Freeware or licensed |
| Command & Conquer | 1995 | DOS | DOSBox | Rights-cleared freeware |
| Prince of Persia | 1989 | DOS | DOSBox | Open-source or authorised distribution |
| Cave Story | 2004 | Native | Native runner | Freeware |
| Tyrian | 1995 | DOS | DOSBox | Freeware |
| Beneath a Steel Sky | 1994 | ScummVM | ScummVM | Freeware |

These titles are examples for planning. Final inclusion depends on current
rights and packaging work.

### 10.2 Enterprise extension

The same bundle contract can preserve:

| Software type | Example use |
|---------------|-------------|
| Legacy COBOL workload | Bank migration bridge |
| Old Win32 internal tool | Operational continuity during replacement |
| Deprecated monitoring agent | Sandboxed stopgap while migration completes |

Enterprise value comes from preservation, verification, and controlled
execution, not from the games catalogue itself.

---

## 11. Delivery phases

| Phase | Outcome |
|-------|---------|
| 1 | Manifest parser and bundle validator |
| 2 | `play`, `play/list`, and `play/verify` command registration |
| 3 | Engine registry and adapter contract |
| 4 | First engine adapter, likely DOSBox |
| 5 | Save-state layout and sandbox policy |
| 6 | `play/bundle` creation flow |
| 7 | Shield-aligned verification and SBOM generation |
| 8 | Catalogue index and launch candidates |
| 9 | Additional adapter coverage, including DOSBox-X for PC-98 and Windows-era images |
| 10 | Optional protected asset and entitlement profile |

The first milestone worth shipping is:

- directory bundle validation
- one runnable engine
- one known-good title
- local BYOROM support

---

## 12. Open questions

- Should `emulator.yaml` remain engine-specific, or be renamed to a more
  neutral `runtime.yaml` later?
- Does the runtime ever need per-bundle signed manifests beyond checksum
  verification?
- Which engine adapters ship by default versus behind build tags?
- Is protected catalogue delivery a separate module, or an optional extension of
  `core/play`?
- How much of the save-state format should be normalised across engines?
- Should enterprise bundles reuse the same `rom/` directory name, or move to a
  more general `artefact/` path?

These questions do not block the initial implementation, but they should remain
visible while the first runnable slice lands.

---

## Changelog

- 2026-04-27: Pass 7 added manifest-declared CPU and memory resource ceilings
  to sandbox policy and launch-plan validation.
- 2026-04-27: Pass 5 added explicit STIM manifest format versioning with a
  legacy migration path, and hardened Shield threat scanning against oversized
  ZIP entries, oversized aggregate archive expansion, and deeply nested
  artefact paths.
- 2026-04-27: Pass 4 added MAME, VICE, FUSE, and standalone Snes9x adapter
  scaffolds, tightened canonical bundle path handling, and expanded sandbox
  read/write allowlist enforcement around runtime config access.
- 2026-04-27: Pass 3 tightened checksum-chain coverage semantics, added
  DOSBox-X adapter planning, enforced launch-plan sandbox boundaries, and pinned
  parser/catalogue verification tests.
- 2026-04-08: Reworked into a cleaner draft with goals, bundle contract,
  runtime architecture, verification model, distribution profiles, delivery
  phases, and open questions.
- 2026-04-01: Initial notes covering STIM bundles, engine registry, verification
  chain, and catalogue direction.
