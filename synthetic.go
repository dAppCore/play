package play

import core "dappco.re/go"

// SyntheticEngine is a smoke-test engine that runs no external emulator.
type SyntheticEngine struct{}

// Name returns the synthetic engine identifier.
func (engine SyntheticEngine) Name() string {
	return "synthetic"
}

// Platforms returns platforms supported by the synthetic smoke engine.
func (engine SyntheticEngine) Platforms() []string {
	return []string{
		"synthetic",
	}
}

// Run prints a deterministic success marker for end-to-end smoke tests.
func (engine SyntheticEngine) Run(_ string, config EngineConfig) error {
	if config.Output != nil {
		core.Print(config.Output, "SYNTHETIC ENGINE OK")
	}

	return nil
}

// Verify performs lightweight adapter validation.
func (engine SyntheticEngine) Verify() error {
	return nil
}

// CodeIdentity returns the synthetic runtime integrity identity.
func (engine SyntheticEngine) CodeIdentity() EngineCodeIdentity {
	return EngineCodeIdentity{
		Name:   engine.Name(),
		Path:   engine.Name(),
		SHA256: virtualEngineCodeSHA256(engine.Name()),
	}
}

// PlanLaunch builds a launch plan for a synthetic bundle.
func (engine SyntheticEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if bundle.Manifest.Runtime.Engine != engine.Name() {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/runtime-mismatch",
			Name:    engine.Name(),
			Message: "bundle runtime does not match the synthetic engine",
		}
	}
	if !supportsPlatform(engine.Platforms(), bundle.Manifest.Platform) {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/platform-unsupported",
			Name:    engine.Name(),
			Message: "bundle platform is not supported by the synthetic engine",
		}
	}

	return LaunchPlan{
		Engine:           engine.Name(),
		Executable:       engine.Name(),
		Arguments:        []string{bundle.Manifest.Artefact.Path},
		WorkingDirectory: ".",
		Entrypoint:       bundle.Manifest.Artefact.Path,
		RuntimeConfig:    bundle.Manifest.Runtime.Config,
		ReadPaths:        manifestLaunchReadPaths(bundle.Manifest),
		WritePaths:       manifestLaunchWritePaths(bundle.Manifest),
		Resources:        bundle.Manifest.Resources,
		NetworkAllowed:   bundle.Manifest.Permissions.Network,
	}, nil
}
