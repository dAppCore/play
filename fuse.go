package play

// FUSEEngine adapts the Free Unix Spectrum Emulator for ZX Spectrum bundles.
type FUSEEngine struct {
	Binary       string
	BinarySHA256 string
}

// Name returns the engine identifier.
func (engine FUSEEngine) Name() string {
	return "fuse"
}

// Platforms returns platforms supported by the initial FUSE scaffold.
func (engine FUSEEngine) Platforms() []string {
	return []string{
		"zx-spectrum",
		"spectrum",
		"zx-spectrum-48k",
		"zx-spectrum-128k",
	}
}

// Acceleration reports FUSE's preferred frame-acceleration behaviour.
func (engine FUSEEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterCRT,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine FUSEEngine) Verify() error {
	return requireEngineBinary(engine.Name(), engine.Binary)
}

// CodeIdentity returns the FUSE runtime integrity identity.
func (engine FUSEEngine) CodeIdentity() EngineCodeIdentity {
	return adapterCodeIdentity(engine.Name(), engine.Binary, engine.BinarySHA256)
}

// Run loads a Spectrum artefact through FUSE using Core's process primitive.
func (engine FUSEEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}

	arguments, err := fuseArguments(artefact, config.ConfigPath, config.Profile)
	if err != nil {
		return err
	}

	return runLaunchPlan(LaunchPlan{
		Engine:           engine.Name(),
		Executable:       engine.Binary,
		Arguments:        arguments,
		WorkingDirectory: ".",
		Entrypoint:       artefact,
		RuntimeConfig:    config.ConfigPath,
		NetworkAllowed:   config.NetworkAllowed,
	}, config)
}

// PlanLaunch builds a launch plan for a FUSE-backed bundle.
func (engine FUSEEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if err := validateAdapterBundle(engine.Name(), engine.Platforms(), bundle); err != nil {
		return LaunchPlan{}, err
	}

	entrypoint := manifestEntrypoint(bundle.Manifest)
	arguments, err := fuseArguments(entrypoint, bundle.Manifest.Runtime.Config, bundle.Manifest.Runtime.Profile)
	if err != nil {
		return LaunchPlan{}, err
	}

	return adapterLaunchPlan(engine.Name(), engine.Binary, bundle, arguments), nil
}

func fuseArguments(entrypoint string, configPath string, profile string) ([]string, error) {
	arguments := []string{}
	if configPath != "" {
		arguments = append(arguments, "--settings", configPath)
	}
	if profile != "" {
		machine, ok := fuseMachine(profile)
		if !ok {
			return nil, EngineError{
				Kind:    "engine/profile-unsupported",
				Name:    "fuse",
				Message: "runtime profile is not supported by FUSE",
			}
		}
		arguments = append(arguments, "--machine", machine)
	}
	arguments = append(arguments, entrypoint)

	return arguments, nil
}

func fuseMachine(profile string) (string, bool) {
	switch profile {
	case "48k", "spectrum-48k", "zx-spectrum-48k":
		return "48", true
	case "128k", "spectrum-128k", "zx-spectrum-128k":
		return "128", true
	default:
		return "", false
	}
}
