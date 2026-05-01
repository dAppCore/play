package play

// Snes9xEngine adapts standalone Snes9x for Super Nintendo bundles.
type Snes9xEngine struct {
	Binary       string
	BinarySHA256 string
}

// Name returns the engine identifier.
func (engine Snes9xEngine) Name() string {
	return "snes9x"
}

// Platforms returns platforms supported by the initial Snes9x scaffold.
func (engine Snes9xEngine) Platforms() []string {
	return []string{
		"snes",
		"super-nintendo",
	}
}

// Acceleration reports Snes9x's preferred frame-acceleration behaviour.
func (engine Snes9xEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterScanline,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine Snes9xEngine) Verify() error {
	return requireEngineBinary(engine.Name(), engine.Binary)
}

// CodeIdentity returns the Snes9x runtime integrity identity.
func (engine Snes9xEngine) CodeIdentity() EngineCodeIdentity {
	return adapterCodeIdentity(engine.Name(), engine.Binary, engine.BinarySHA256)
}

// Run executes a SNES artefact through Snes9x using Core's process primitive.
func (engine Snes9xEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}

	arguments := snes9xArguments(artefact, config.ConfigPath)
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

// PlanLaunch builds a launch plan for a Snes9x-backed bundle.
func (engine Snes9xEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if err := validateAdapterBundle(engine.Name(), engine.Platforms(), bundle); err != nil {
		return LaunchPlan{}, err
	}

	entrypoint := manifestEntrypoint(bundle.Manifest)
	arguments := snes9xArguments(entrypoint, bundle.Manifest.Runtime.Config)

	return adapterLaunchPlan(engine.Name(), engine.Binary, bundle, arguments), nil
}

func snes9xArguments(entrypoint string, configPath string) []string {
	arguments := []string{}
	if configPath != "" {
		arguments = append(arguments, "-conf", configPath)
	}
	arguments = append(arguments, entrypoint)

	return arguments
}
