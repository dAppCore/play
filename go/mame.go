package play

import "path"

// MAMEEngine adapts MAME for arcade and Neo Geo ROM-set bundles.
type MAMEEngine struct {
	Binary       string
	BinarySHA256 string
}

// Name returns the engine identifier.
func (engine MAMEEngine) Name() string {
	return "mame"
}

// Platforms returns platforms supported by the initial MAME scaffold.
func (engine MAMEEngine) Platforms() []string {
	return []string{
		"arcade",
		"neo-geo",
	}
}

// Acceleration reports MAME's preferred frame-acceleration behaviour.
func (engine MAMEEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterScanline,
			FrameFilterCRT,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine MAMEEngine) Verify() error {
	return requireEngineBinary(engine.Name(), engine.Binary)
}

// CodeIdentity returns the MAME runtime integrity identity.
func (engine MAMEEngine) CodeIdentity() EngineCodeIdentity {
	return adapterCodeIdentity(engine.Name(), engine.Binary, engine.BinarySHA256)
}

// Run executes an arcade artefact through MAME using Core's process primitive.
func (engine MAMEEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}

	arguments, err := mameArguments(artefact, config.ConfigPath, config.Profile)
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
		Resources:        config.Resources,
		NetworkAllowed:   config.NetworkAllowed,
	}, config)
}

// PlanLaunch builds a launch plan for a MAME-backed bundle.
func (engine MAMEEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if err := validateAdapterBundle(engine.Name(), engine.Platforms(), bundle); err != nil {
		return LaunchPlan{}, err
	}

	entrypoint := manifestEntrypoint(bundle.Manifest)
	arguments, err := mameArguments(entrypoint, bundle.Manifest.Runtime.Config, bundle.Manifest.Runtime.Profile)
	if err != nil {
		return LaunchPlan{}, err
	}

	return adapterLaunchPlan(engine.Name(), engine.Binary, bundle, arguments), nil
}

func mameArguments(entrypoint string, configPath string, profile string) ([]string, error) {
	if profile == "" {
		return nil, EngineError{
			Kind:    "engine/profile-required",
			Name:    "mame",
			Message: "runtime profile is required",
		}
	}

	arguments := []string{}
	if configPath != "" {
		configDirectory := path.Dir(configPath)
		if configDirectory == "." {
			configDirectory = "."
		}
		arguments = append(arguments, "-inipath", configDirectory)
	}

	romDirectory := path.Dir(entrypoint)
	if romDirectory != "." {
		arguments = append(arguments, "-rompath", romDirectory)
	}
	arguments = append(arguments, profile)

	return arguments, nil
}
