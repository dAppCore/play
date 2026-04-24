---
module: core/play
repo: core/play
lang: multi
tier: consumer
depends:
  - code/core/app
  - code/core/gui
  - code/core/go-process
tags:
  - games
  - apple
  - darwin
  - input
  - audio
  - presentation
  - runtime
---

# Core Play RFC — Apple Platform Runtime for Games and Emulation

> `core play` needs Apple-specific runtime work beyond MLX.
> That work should stay separate from Metal compute.

**Module:** `dappco.re/go/play`  
**Repository:** `core/play`  
**Status:** Draft, companion RFC  
**Audience:** `core/play` and `core/gui` maintainers

---

## 1. Purpose

This RFC defines the Apple-platform runtime concerns needed by `core/play` that
sit **outside** `go-mlx`.

The MLX RFC covers Metal compute for frame processing. This RFC covers the
Apple-specific runtime layer around it:

- presentation and frame pacing
- input and hot-plug handling
- low-latency audio output
- low-copy hand-off into the UI layer
- platform policy such as fullscreen, power, and lifecycle handling

This RFC is supplemental to `docs/RFC.md`. It does not replace the STIM bundle
contract, command surface, or verification model.

---

## 2. Why a separate RFC is needed

`docs/RFC.mlx-gaming.md` correctly keeps `go-mlx` focused on compute.

That leaves a second set of Apple-specific needs which are still important for a
good preserved-software runtime:

- stable frame pacing on Apple displays
- low-overhead presentation of emulator or native frames
- controller and keyboard input using Apple platform services
- audio output that stays in sync with the frame loop
- runtime policy for resize, fullscreen, suspend, and thermal pressure

These concerns are real, but they are not MLX concerns.

---

## 3. Boundary with MLX

### 3.1 `go-mlx` should own

- compute kernels
- pixel and byte buffer processing
- metrics for compute work
- optional acceleration for frame transforms

### 3.2 Apple runtime layer should own

- presentation surfaces and present timing
- controller and keyboard input
- audio device output
- lifecycle and window policy
- runtime integration with `core/gui`

### 3.3 `core/play` should still own

- engine selection
- bundle verification
- save-state policy
- acceleration policy and fallback rules

---

## 4. Goals

- Provide a clean Apple runtime path for games and preserved software
- Keep MLX optional and tightly scoped to compute
- Make frame pacing, presentation, input, and audio predictable on Apple
  platforms
- Preserve CPU-only fallback behaviour when GPU acceleration is unavailable
- Keep the runtime aligned with the STIM sandbox and verification model

### 4.1 Non-goals

- replacing `core/gui`
- turning `core/play` into a bespoke Apple-only game engine
- moving emulator CPU logic into Apple platform APIs
- duplicating MLX compute responsibilities in a second package

---

## 5. Capabilities needed

### 5.1 Presentation and frame pacing

Need a runtime surface for:

- display-linked present scheduling
- window resize and fullscreen handling
- aspect-ratio, overscan, and integer-scaling policy
- portrait or rotated-title presentation support
- graceful fallback when MLX is absent

### 5.2 Low-copy presentation interop

Need a practical hand-off from `core/play` frame output into presentation:

- CPU-produced RGBA frame path
- MLX-processed frame path
- clear documentation of copy versus low-copy behaviour
- stable surface contracts so adapters do not depend on Apple internals

The exact mechanism may vary, but the contract should allow efficient transfer
without pulling presentation into `go-mlx`.

### 5.3 Input

Need a platform input layer for:

- controller discovery and hot-plug
- keyboard mapping
- per-engine input profile selection
- system-safe pause, resume, and focus transitions

### 5.4 Audio

Need an Apple output path for:

- low-latency playback
- sample-rate adaptation where required
- synchronisation with emulator timing
- clean suspend, resume, and mute behaviour

### 5.5 Runtime policy

Need policy hooks for:

- thermal or power-pressure downgrade decisions
- foreground and background lifecycle handling
- fullscreen and window-mode transitions
- capture or screenshot policy where allowed by bundle runtime rules

---

## 6. First implementation slice

The first slice should stay narrow:

1. present CPU or MLX-produced RGBA frames through `core/gui`
2. add display-linked frame pacing
3. add keyboard and controller basics
4. add low-latency audio output
5. keep fallback ordinary when any Apple-specific feature is unavailable

This is enough to make the Apple runtime path usable without overcommitting to a
large platform layer too early.

---

## 7. Nice-to-have later

- per-device capability tiers for presentation and filter budgets
- save-state thumbnail generation hooks
- haptics for supported controllers
- external-display or AirPlay-aware presentation policy
- bundle-level defaults for aspect ratio, overscan, and rotation
- richer frame-pacing diagnostics for performance tuning

---

## 8. Relationship to STIM bundles

This RFC does not change the STIM bundle shape directly.

If later work needs new manifest fields, they should remain optional and should
describe runtime preference rather than bundle validity. Examples may include:

- preferred aspect policy
- default rotation
- controller profile hints
- audio-latency preference bands

Bundle verification, hash integrity, and save-state boundaries remain governed
by `docs/RFC.md`.

---

## 9. Requested outcome

The practical request is:

1. keep `go-mlx` compute-focused
2. define a separate Apple runtime surface for presentation, input, and audio
3. integrate that surface through `core/gui` and `core/play` policy
4. preserve deterministic and portable fallback behaviour

If this lands, `core/play` can adopt Apple-specific quality improvements without
overloading the MLX integration or blurring module boundaries.

---

## Changelog

- 2026-04-15: Initial companion RFC for Apple-specific runtime work outside
  MLX.
