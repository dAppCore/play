# core/play

STIM Game & Software Preservation runtime — run preserved software in deterministic hash-verified STIM bundles.

> "Games are the demo. Legacy enterprise is the product."

**Module:** `dappco.re/go/play`
**Spec:** [`docs/RFC.md`](docs/RFC.md)
**Depends on:** `core/app` (manifest/boot contract — see [`docs/RFC.app.md`](docs/RFC.app.md)), `core/gui` (windowing)

## Scope

`core play` runs preserved software inside STIM (Sandboxed Temporal Isolation Module) containers. A STIM bundle is a deterministic, hash-verified, SBOM-tracked archive containing:

1. The original artefact (ROM, binary, installer)
2. A runtime engine (emulator core, compatibility layer, or native runner)
3. A manifest describing inputs, outputs, and verification chain

```bash
core play                          # read manifest.yaml from cwd
core play mega-lo-mania            # from games/ parent dir
core play command-and-conquer      # same
```

## Status

Bootstrapped repo — spec in `docs/`, implementation pending. Depends on core/app (keystone boot runtime) landing first.

## Licence

EUPL-1.2
