package play

// DOSBoxEngine is a first-pass DOSBox adapter scaffold.
type DOSBoxEngine struct {
	Binary       string
	BinarySHA256 string
}

// Name returns the engine identifier.
func (engine DOSBoxEngine) Name() string {
	return "dosbox"
}

// Platforms returns platforms supported by DOSBox.
func (engine DOSBoxEngine) Platforms() []string {
	return []string{
		"dos",
	}
}

// Acceleration reports DOSBox's preferred frame-acceleration behaviour.
func (engine DOSBoxEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterBilinear,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine DOSBoxEngine) Verify() error {
	if engine.Binary == "" {
		return EngineError{
			Kind:    "engine/binary-required",
			Name:    engine.Name(),
			Message: "binary path is required",
		}
	}

	return nil
}

// CodeIdentity returns the DOSBox runtime integrity identity.
func (engine DOSBoxEngine) CodeIdentity() EngineCodeIdentity {
	enginePath := defaultString(engine.Binary, engine.Name())
	engineHash := engine.BinarySHA256
	if engineHash == "" {
		engineHash = virtualEngineCodeSHA256(engine.Name())
	}

	return EngineCodeIdentity{
		Name:   engine.Name(),
		Path:   enginePath,
		SHA256: engineHash,
	}
}

// Run executes a DOS artefact through Core's process primitive.
func (engine DOSBoxEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}

	arguments := []string{artefact}
	if config.ConfigPath != "" {
		arguments = append(arguments, "-conf", config.ConfigPath)
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

// PlanLaunch builds a launch plan for a DOS bundle.
func (engine DOSBoxEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if bundle.Manifest.Runtime.Engine != engine.Name() {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/runtime-mismatch",
			Name:    engine.Name(),
			Message: "bundle runtime does not match the DOSBox engine",
		}
	}
	if !supportsPlatform(engine.Platforms(), bundle.Manifest.Platform) {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/platform-unsupported",
			Name:    engine.Name(),
			Message: "bundle platform is not supported by DOSBox",
		}
	}

	entrypoint := bundle.Manifest.Runtime.Entrypoint
	if entrypoint == "" {
		entrypoint = bundle.Manifest.Artefact.Path
	}

	arguments := []string{
		entrypoint,
	}
	if bundle.Manifest.Runtime.Config != "" {
		arguments = append(arguments, "-conf", bundle.Manifest.Runtime.Config)
	}

	return LaunchPlan{
		Engine:           engine.Name(),
		Executable:       engine.Binary,
		Arguments:        arguments,
		WorkingDirectory: ".",
		Entrypoint:       entrypoint,
		RuntimeConfig:    bundle.Manifest.Runtime.Config,
		ReadPaths:        manifestLaunchReadPaths(bundle.Manifest),
		WritePaths:       manifestLaunchWritePaths(bundle.Manifest),
		Resources:        bundle.Manifest.Resources,
		NetworkAllowed:   bundle.Manifest.Permissions.Network,
	}, nil
}

func supportsPlatform(platforms []string, wanted string) bool {
	for _, platform := range platforms {
		if platform == wanted {
			return true
		}
	}

	return false
}
