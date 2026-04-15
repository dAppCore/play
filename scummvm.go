package play

import "path"

// ScummVMEngine is a first-pass ScummVM adapter scaffold.
type ScummVMEngine struct {
	Binary string
}

// Name returns the engine identifier.
func (engine ScummVMEngine) Name() string {
	return "scummvm"
}

// Platforms returns platforms supported by the initial ScummVM scaffold.
func (engine ScummVMEngine) Platforms() []string {
	return []string{
		"scummvm",
		"point-and-click",
	}
}

// Acceleration reports ScummVM's preferred frame-acceleration behaviour.
func (engine ScummVMEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterBilinear,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine ScummVMEngine) Verify() error {
	if engine.Binary == "" {
		return EngineError{
			Kind:    "engine/binary-required",
			Name:    engine.Name(),
			Message: "binary path is required",
		}
	}

	return nil
}

// PlanLaunch builds a launch plan for a ScummVM-backed bundle.
func (engine ScummVMEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if bundle.Manifest.Runtime.Engine != engine.Name() {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/runtime-mismatch",
			Name:    engine.Name(),
			Message: "bundle runtime does not match the ScummVM engine",
		}
	}
	if !supportsPlatform(engine.Platforms(), bundle.Manifest.Platform) {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/platform-unsupported",
			Name:    engine.Name(),
			Message: "bundle platform is not supported by ScummVM",
		}
	}
	if bundle.Manifest.Runtime.Profile == "" {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/profile-required",
			Name:    engine.Name(),
			Message: "runtime profile is required",
		}
	}

	entrypoint := bundle.Manifest.Runtime.Entrypoint
	if entrypoint == "" {
		entrypoint = bundle.Manifest.Artefact.Path
	}

	dataPath := path.Dir(entrypoint)
	if dataPath == "." {
		dataPath = path.Dir(bundle.Manifest.Artefact.Path)
	}

	arguments := []string{
		"--path=" + dataPath,
		bundle.Manifest.Runtime.Profile,
	}

	return LaunchPlan{
		Engine:           engine.Name(),
		Executable:       engine.Binary,
		Arguments:        arguments,
		WorkingDirectory: ".",
		Entrypoint:       entrypoint,
		RuntimeConfig:    bundle.Manifest.Runtime.Config,
		ReadPaths:        clonePaths(bundle.Manifest.Permissions.FileSystem.Read),
		WritePaths:       clonePaths(bundle.Manifest.Permissions.FileSystem.Write),
		NetworkAllowed:   bundle.Manifest.Permissions.Network,
	}, nil
}
