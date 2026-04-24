---
module: core/play
repo: core/play
lang: multi
tier: consumer
depends:
  - code/core/gui
  - code/core/go-mlx
  - code/core/go-process
tags:
  - games
  - metal
  - mlx
  - emulation
  - rendering
  - acceleration
---

# Core Play RFC — MLX Gaming and Emulation Acceleration

> `core play` should use Apple GPU acceleration where it helps.
> `go-mlx` is the likely compute layer, but only for the right jobs.

**Module:** `dappco.re/go/play`  
**Repository:** `core/play`  
**Related module:** `dappco.re/go/mlx`  
**Status:** Draft, feature-request RFC  
**Audience:** `core/play`, `go-mlx`, and `core/gui` maintainers

---

## 1. Purpose

This RFC defines how `core/play` should use Apple Metal through `go-mlx` for
gaming and emulation workloads on Apple Silicon.

It does **not** propose using MLX as a replacement for emulator CPU cores.
Instead, it defines the narrow, high-value ways in which Metal compute can help
the runtime:

- frame scaling
- colour conversion
- post-processing
- palette expansion
- low-copy framebuffer handling
- frame-oriented profiling and scheduling

It also defines the exact gaps between what `go-mlx` exposes today and what
`core/play` needs in order to adopt it cleanly.

---

## 2. Current state

`go-mlx` is already a strong Apple Silicon package, but it is currently aimed at
MLX inference and training workloads:

- LLM inference
- LoRA training
- tensor and array operations
- MLX-backed Metal compute
- model loading, tokenisation, sampling, and cache management

That is valuable, but it is not the same thing as an emulator-facing GPU API.

### 2.1 What `go-mlx` already gives us

From its current public and documented shape, `go-mlx` already offers:

- Apple Silicon and Metal availability detection
- memory and cache controls for Metal-backed workloads
- lazy-evaluated array operations
- explicit materialisation and cleanup
- fused compute kernels for ML workloads
- a stable Go wrapper over MLX / mlx-c

### 2.2 What it does not yet give us

`core/play` still lacks a suitable public surface for:

- frame-oriented compute dispatch
- framebuffer upload and download
- pixel-format-aware image processing
- reusable graphics-adjacent compute kernels
- explicit per-frame lifecycle control
- low-copy buffer interop for emulators

### 2.3 Important boundary

`go-mlx` is currently a **Metal compute** package, not a:

- windowing package
- swapchain or presentation package
- gamepad or keyboard input package
- audio output package
- emulator runtime package

That boundary is correct and should stay correct.

---

## 3. Decision

`core/play` should use `go-mlx` for **compute acceleration around emulation**,
not for the emulator itself.

### 3.1 Good uses of `go-mlx`

- scaling 240p, 480p, or handheld framebuffers to modern display sizes
- RGB565 or indexed-palette expansion into display-friendly formats
- scanline, CRT, LCD, or sharpening passes
- optional software-renderer acceleration for old 2D or fixed-function style output
- frame analytics and profiling hooks

### 3.2 Bad uses of `go-mlx`

- CPU instruction emulation
- audio playback
- controller input
- general process management
- rendering surface ownership
- replacing `core/gui`

### 3.3 Why this is the right split

Most emulator workloads still divide naturally like this:

1. CPU emulation happens on CPU
2. the emulator emits a framebuffer or video surface
3. optional GPU work transforms the frame
4. a UI layer presents the final image

That is the cleanest integration point for `go-mlx`.

---

## 4. Goals

- Reuse `go-mlx` as the Apple Metal compute layer for `core/play`
- Avoid creating a second Apple GPU compute package if `go-mlx` can host the work
- Keep presentation, input, and audio outside `go-mlx`
- Make the first integration useful for real games quickly
- Preserve a CPU-only fallback path
- Keep the `core/play` engine adapters portable even when GPU acceleration is unavailable

## 4.1 Non-goals

- JIT or shader-recompiler replacement for emulator cores
- full render-graph abstraction
- Metal UI integration inside `go-mlx`
- OpenGL / Vulkan emulation inside `go-mlx`
- universal cross-platform GPU abstraction in this first phase

---

## 5. Architecture position

The desired runtime stack for Apple Silicon should look like this:

```text
STIM bundle
   │
   ▼
core/play engine adapter
   │
   ├── CPU emulation / runner
   │
   ├── frame output
   ▼
optional go-mlx compute stage
   │
   ▼
core/gui or another presentation layer
   │
   ▼
display
```

### 5.1 Responsibilities by module

| Module | Responsibility |
|--------|----------------|
| `core/play` | engine selection, bundle verification, runtime policy, frame pipeline orchestration |
| `go-mlx` | Metal compute and buffer processing |
| `core/gui` | windows, view surfaces, presentation, scaling policy integration |
| engine adapters | emulator-specific frame extraction and launch semantics |

---

## 6. Exact features needed from `go-mlx`

This section is the core of the RFC.

### 6.1 Public raw-compute dispatch API

`core/play` needs a small public API that does **not** go through language-model
or training concepts.

Minimum capabilities:

- create or access a Metal-backed execution context
- create input buffers from Go byte slices
- create output buffers with explicit size and format
- dispatch a named compute kernel
- wait for completion
- read results back into Go memory
- return structured, non-LLM-specific errors

The key point is this:

`core/play` needs a **frame compute API**, not a `TextModel` API.

### 6.2 Framebuffer and image buffer support

The package needs first-class support for frame-style data, not only tensor-style
data.

Minimum capabilities:

- upload frame data with width, height, and stride
- represent packed pixel buffers directly
- read output into display-ready memory
- support predictable format conversion

Minimum pixel formats:

- `RGBA8`
- `BGRA8`
- `RGB565`
- `XRGB8888`
- indexed 8-bit palette source, if feasible

### 6.3 Reusable frame-processing kernels

`core/play` needs a reusable kernel set aimed at emulator presentation.

MVP kernels:

- nearest-neighbour scale
- bilinear scale
- integer scale with clean pixel edges
- RGB565 → RGBA8 conversion
- BGRA / RGBA swizzle
- palette expansion

Good second-wave kernels:

- scanline filter
- CRT mask / phosphor approximation
- LCD ghosting / grid simulation for handheld systems
- sharpening or softening passes
- simple bloom or glow for suitable systems

### 6.4 Per-frame execution lifecycle

Emulation is a frame loop. The API should acknowledge that.

Minimum capabilities:

- begin frame job
- submit one or more compute passes
- finish and synchronise the frame
- optionally query timings

The goal is not a full render graph. The goal is a predictable, composable
execution model for 60 Hz or similar workloads.

### 6.5 Low-copy or low-overhead buffer interop

The package needs a fast path between emulator memory and Metal buffers.

Minimum capabilities:

- upload from Go-managed byte slices without hidden format gymnastics
- avoid forcing everything through tensor-shaped abstractions
- document the copy behaviour clearly
- provide the least-copy path supported safely by MLX and Go

### 6.6 Explicit profiling and capacity reporting

`core/play` needs enough observability to decide whether an acceleration path is
worth using.

Minimum capabilities:

- device info
- availability flags
- kernel timing or per-frame timing
- memory used by active jobs
- peak memory for a frame pipeline

This allows runtime policy such as:

- use GPU filter on M2 and above
- disable heavy filter if memory is tight
- fall back to CPU scaling if frame budget is exceeded

---

## 7. Features that should stay outside `go-mlx`

The following should **not** be added to `go-mlx` for this work:

- window creation
- layer-backed presentation
- swapchain management
- view ownership
- controller input
- keyboard input
- audio device output
- process management
- sandbox policy

Those belong elsewhere:

| Concern | Better home |
|---------|-------------|
| presentation | `core/gui` or a focused Metal presentation wrapper |
| input | `core/gui` or platform input layer |
| audio | dedicated audio package |
| process isolation | `core/go-process` and runtime policy |

---

## 8. Proposed public API shape

The exact names may vary, but the surface should look roughly like this.

### 8.1 Device and context

```go
type Compute interface {
    Available() bool
    DeviceInfo() DeviceInfo
    NewSession(opts ...SessionOption) (Session, error)
}

type Session interface {
    Close() error
    NewPixelBuffer(desc PixelBufferDesc) (PixelBuffer, error)
    NewByteBuffer(size int) (ByteBuffer, error)
    Run(kernel string, args KernelArgs) error
    Sync() error
    Metrics() SessionMetrics
}
```

### 8.2 Buffer types

```go
type PixelFormat string

const (
    PixelRGBA8   PixelFormat = "rgba8"
    PixelBGRA8   PixelFormat = "bgra8"
    PixelRGB565  PixelFormat = "rgb565"
    PixelXRGB8888 PixelFormat = "xrgb8888"
)

type PixelBufferDesc struct {
    Width  int
    Height int
    Stride int
    Format PixelFormat
}
```

### 8.3 Kernel execution

```go
type KernelArgs struct {
    Inputs  map[string]Buffer
    Outputs map[string]Buffer
    Scalars map[string]float64
}
```

### 8.4 Intent of this API

This does **not** need to look like an MLX training API.
It only needs to be:

- stable
- explicit
- frame-friendly
- usable from `core/play`

---

## 9. `core/play` integration model

### 9.1 Engine adapter contract

Engine adapters in `core/play` should be able to say:

- CPU-only
- GPU-optional
- GPU-preferred for post-processing

The adapter itself should not need to know MLX internals.

### 9.2 Frame pipeline model

For a typical emulator frame:

1. emulator core produces a framebuffer
2. adapter chooses the acceleration policy
3. frame is uploaded to `go-mlx`
4. one or more kernels run
5. output buffer is handed to presentation

### 9.3 Example policy

```text
Genesis game
  → CPU emulation in RetroArch core
  → framebuffer 320x224 RGB565
  → go-mlx converts RGB565 → RGBA8
  → nearest-neighbour integer scale to window backbuffer size
  → optional CRT pass
  → core/gui presents
```

### 9.4 Fallback policy

If any of these are false:

- not `darwin/arm64`
- Metal unavailable
- `go-mlx` unavailable
- kernel creation fails
- frame cost exceeds policy budget

Then `core/play` falls back to CPU-side conversion and normal presentation.

Fallback must be ordinary, not exceptional.

---

## 10. Why MLX is still a good fit

Although MLX is an ML-oriented framework, it is still attractive here because:

- it already owns the Apple Metal integration work
- it already manages GPU memory, materialisation, and error handling
- it already ships Apple Silicon-specific acceleration logic
- it keeps Core’s Apple GPU knowledge concentrated in one repo

The missing piece is simply the public surface for non-LLM workloads.

---

## 11. Why another API is still needed

Even after `go-mlx` grows the right compute surface, another API will still be
required for presentation and platform runtime concerns.

### 11.1 Presentation

`go-mlx` should not own:

- render surfaces
- windows
- layer presentation
- resize handling
- fullscreen policy

That is a `core/gui` concern.

### 11.2 Audio

`go-mlx` should not own:

- audio device enumeration
- mixing
- latency buffering
- resampling

### 11.3 Input

`go-mlx` should not own:

- gamepad polling
- keyboard mapping
- touch translation
- hot-plug lifecycle

So the answer to the original question is:

**No, another ML API is not needed. But another non-ML API still is, for
presentation, audio, and input.**

---

## 12. First implementation slice

The easiest useful first slice is deliberately narrow.

### 12.1 MVP

Add to `go-mlx`:

1. raw compute session
2. pixel buffer abstraction
3. nearest-neighbour scale kernel
4. RGB565 → RGBA8 conversion kernel
5. basic timing and memory metrics

Then wire into `core/play`:

1. RetroArch adapter exposes framebuffer metadata
2. `core/play` frame pipeline calls `go-mlx` when available
3. `core/gui` presents the processed frame

### 12.2 Why this slice first

Because it delivers visible value immediately:

- crisp scaling
- correct format conversion
- less CPU work in the hot frame path
- no need to solve full presentation architecture first

---

## 13. Second implementation slice

Once the MVP works:

- bilinear and integer-scaling variants
- scanline and CRT filters
- palette expansion kernels
- frame-time policy controls
- engine-level acceleration hints in `core/play`

---

## 14. Nice-to-have later

These are explicitly lower priority:

- shader-like post-processing chains
- LCD handheld simulation filters
- dynamic filter presets per title
- offline frame enhancement for screenshots
- compute-assisted texture unpacking for software renderers

Additional nice-to-have features that still fit the `go-mlx` boundary:

- rotation and orientation kernels for portrait arcade and handheld titles
- dirty-region or partial-frame upload paths for UI-heavy preserved software
- colour-management helpers for gamma, contrast, and display-target tuning
- descriptor and buffer-pool reuse to reduce per-frame allocation churn
- pipeline warm-up or kernel preloading to avoid first-frame stutter
- deterministic thumbnail and screenshot derivation for save-state previews
- bundle-declared filter profiles with runtime-safe user overrides
- capability tiers so filter policy can target M1, M2, or newer devices cleanly

### 14.1 Separate Apple runtime work

Some Apple-specific work is desirable, but it does **not** belong in `go-mlx`.

The companion RFC for that work should be:

- `docs/RFC.apple-platform-runtime.md`

That RFC should cover the Apple runtime concerns which sit beside MLX compute:

- display-link timing and frame pacing
- low-copy presentation hand-off into `core/gui`
- controller, keyboard, and hot-plug handling
- low-latency audio output and synchronisation
- thermal, power, and foreground/background runtime policy

This keeps the split clean:

- `go-mlx` accelerates compute around frames
- `core/play` owns policy and engine integration
- `core/gui` or an Apple runtime layer owns presentation, input, and audio

---

## 15. Security and runtime policy

The GPU path must not weaken the `core/play` sandbox model.

Requirements:

- acceleration is local-only
- no implicit network use
- no external helper processes required in the hot path
- no writes outside declared runtime paths
- failures remain contained to the bundle session

GPU acceleration is an optimisation layer, not a trust boundary escape hatch.

---

## 16. Testing expectations

### 16.1 `go-mlx` tests

Need tests for:

- buffer creation
- pixel-format validation
- kernel dispatch
- output correctness for known fixtures
- timing and memory reporting

### 16.2 `core/play` tests

Need tests for:

- policy selection with and without Metal
- fallback behaviour
- deterministic frame output for known fixtures
- adapter-level launch planning unaffected by GPU absence

### 16.3 Golden fixtures

Use fixed frame fixtures:

- Genesis RGB565 test frame
- SNES RGBA test frame
- ScummVM point-and-click background frame
- palette-indexed handheld frame

---

## 17. Risks

### 17.1 Wrong abstraction risk

If `go-mlx` exposes only tensor-shaped ML concepts, `core/play` will fight the
API and integration will stay awkward.

### 17.2 Scope creep risk

If `go-mlx` grows into a presentation or multimedia package, it will become too
broad and harder to maintain.

### 17.3 Over-optimisation risk

Some emulators will not benefit materially from GPU work. The runtime should
avoid turning acceleration into a requirement.

### 17.4 Compatibility risk

This work is Apple-Silicon-specific by nature. `core/play` must preserve
portable CPU behaviour on all other targets.

---

## 18. Open questions

- Should the first public compute surface live at the root of `go-mlx`, or in a
  narrow subpackage intended for non-LLM use?
- Should pixel buffers be distinct from generic byte buffers, or the same type
  with stronger descriptors?
- Should kernel names be hard-coded, or should `go-mlx` expose a registration
  model for reusable compute kernels?
- Should frame profiling be per-kernel, per-session, or both?
- Should `core/play` select filters per bundle manifest, per user preference, or
  both?

---

## 19. Requested outcome from `go-mlx`

The exact ask is:

1. keep `go-mlx` focused on Metal compute
2. expose a small non-LLM compute API
3. support pixel-buffer workloads directly
4. ship a handful of emulator-relevant kernels
5. leave presentation, input, and audio outside the package

If that lands, `core/play` can adopt Apple GPU acceleration without inventing a
second compute stack.

---

## Changelog

- 2026-04-15: Initial RFC for using `go-mlx` as the Metal compute layer for
  `core/play` gaming and emulation acceleration.
- 2026-04-15: Expanded the nice-to-have roadmap and named a companion Apple
  runtime RFC for non-MLX platform work.
