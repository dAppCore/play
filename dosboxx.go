package play

// DOSBoxXEngine adapts DOSBox-X for DOS, PC-98, and Windows-era bundles.
type DOSBoxXEngine struct {
	Binary       string
	BinarySHA256 string
}

// Name returns the engine identifier.
func (engine DOSBoxXEngine) Name() string {
	return "dosbox-x"
}

// Platforms returns platforms supported by DOSBox-X.
func (engine DOSBoxXEngine) Platforms() []string {
	return []string{
		"dos",
		"pc-98",
		"windows-3x",
		"windows-9x",
	}
}

// Acceleration reports DOSBox-X's preferred frame-acceleration behaviour.
func (engine DOSBoxXEngine) Acceleration() AccelerationDescriptor {
	return AccelerationDescriptor{
		Mode: AccelerationAuto,
		PreferredFilters: []FrameFilter{
			FrameFilterNearest,
			FrameFilterBilinear,
			FrameFilterCRT,
		},
	}
}

// Verify performs lightweight adapter validation.
func (engine DOSBoxXEngine) Verify() error {
	if engine.Binary == "" {
		return EngineError{
			Kind:    "engine/binary-required",
			Name:    engine.Name(),
			Message: "binary path is required",
		}
	}

	return nil
}

// CodeIdentity returns the DOSBox-X runtime integrity identity.
func (engine DOSBoxXEngine) CodeIdentity() EngineCodeIdentity {
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

// Run executes an artefact through DOSBox-X using Core's process primitive.
func (engine DOSBoxXEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}

	arguments, err := dosBoxXArguments(artefact, config.ConfigPath, config.Profile)
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

// PlanLaunch builds a launch plan for a DOSBox-X-backed bundle.
func (engine DOSBoxXEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if bundle.Manifest.Runtime.Engine != engine.Name() {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/runtime-mismatch",
			Name:    engine.Name(),
			Message: "bundle runtime does not match the DOSBox-X engine",
		}
	}
	if !supportsPlatform(engine.Platforms(), bundle.Manifest.Platform) {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/platform-unsupported",
			Name:    engine.Name(),
			Message: "bundle platform is not supported by DOSBox-X",
		}
	}

	entrypoint := bundle.Manifest.Runtime.Entrypoint
	if entrypoint == "" {
		entrypoint = bundle.Manifest.Artefact.Path
	}

	arguments, err := dosBoxXArguments(entrypoint, bundle.Manifest.Runtime.Config, bundle.Manifest.Runtime.Profile)
	if err != nil {
		return LaunchPlan{}, err
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
		NetworkAllowed:   bundle.Manifest.Permissions.Network,
	}, nil
}

type dosBoxXProfile struct {
	Machine string
	Boot    bool
}

func dosBoxXArguments(entrypoint string, configPath string, profile string) ([]string, error) {
	selectedProfile, ok := dosBoxXProfileFor(defaultString(profile, "dos"))
	if !ok {
		return nil, EngineError{
			Kind:    "engine/profile-unsupported",
			Name:    "dosbox-x",
			Message: "runtime profile is not supported by DOSBox-X",
		}
	}

	arguments := []string{}
	if configPath != "" {
		arguments = append(arguments, "-conf", configPath)
	}
	if selectedProfile.Machine != "" {
		arguments = append(arguments, "-set", "dosbox machine="+selectedProfile.Machine)
	}
	if selectedProfile.Boot {
		arguments = append(arguments, "-c", "BOOT "+entrypoint)
	} else {
		arguments = append(arguments, entrypoint)
	}

	return arguments, nil
}

func dosBoxXProfileFor(profile string) (dosBoxXProfile, bool) {
	switch profile {
	case "dos":
		return dosBoxXProfile{Machine: "svga_s3"}, true
	case "pc-98":
		return dosBoxXProfile{Machine: "pc98", Boot: true}, true
	case "windows-3x":
		return dosBoxXProfile{Machine: "svga_s3"}, true
	case "windows-9x":
		return dosBoxXProfile{Machine: "svga_s3", Boot: true}, true
	default:
		return dosBoxXProfile{}, false
	}
}
