package play

import "path"

// RetroArchEngine is a first-pass RetroArch adapter scaffold.
type RetroArchEngine struct {
	Binary        string
	BinarySHA256  string
	CoreDirectory string
}

// Name returns the engine identifier.
func (engine RetroArchEngine) Name() string {
	return "retroarch"
}

// Platforms returns platforms supported by the initial RetroArch scaffold.
func (engine RetroArchEngine) Platforms() []string {
	return []string{
		"sega-genesis",
		"sega-mega-drive",
		"snes",
		"super-nintendo",
		"nes",
		"game-boy",
		"game-boy-color",
		"game-boy-advance",
		"gba",
	}
}

// Acceleration reports RetroArch's preferred frame-acceleration behaviour.
func (engine RetroArchEngine) Acceleration() AccelerationDescriptor {
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
func (engine RetroArchEngine) Verify() error {
	if engine.Binary == "" {
		return EngineError{
			Kind:    "engine/binary-required",
			Name:    engine.Name(),
			Message: "binary path is required",
		}
	}

	return nil
}

// CodeIdentity returns the RetroArch runtime integrity identity.
func (engine RetroArchEngine) CodeIdentity() EngineCodeIdentity {
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

// Run executes a RetroArch artefact through Core's process primitive.
func (engine RetroArchEngine) Run(artefact string, config EngineConfig) error {
	if err := engine.Verify(); err != nil {
		return err
	}
	if config.Profile == "" {
		return EngineError{
			Kind:    "engine/profile-required",
			Name:    engine.Name(),
			Message: "runtime profile is required",
		}
	}

	coreName, found := retroArchCoreName(config.Profile)
	if !found {
		return EngineError{
			Kind:    "engine/profile-unsupported",
			Name:    engine.Name(),
			Message: "runtime profile is not supported by RetroArch",
		}
	}

	corePath := path.Join(retroArchCoreDirectory(engine), coreName)
	return runLaunchPlan(LaunchPlan{
		Engine:           engine.Name(),
		Executable:       engine.Binary,
		Arguments:        []string{"-L", corePath, artefact},
		WorkingDirectory: ".",
		Entrypoint:       artefact,
		RuntimeConfig:    config.ConfigPath,
		NetworkAllowed:   config.NetworkAllowed,
	}, config)
}

// PlanLaunch builds a launch plan for a RetroArch-backed bundle.
func (engine RetroArchEngine) PlanLaunch(bundle Bundle) (LaunchPlan, error) {
	if err := engine.Verify(); err != nil {
		return LaunchPlan{}, err
	}
	if bundle.Manifest.Runtime.Engine != engine.Name() {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/runtime-mismatch",
			Name:    engine.Name(),
			Message: "bundle runtime does not match the RetroArch engine",
		}
	}
	if !supportsPlatform(engine.Platforms(), bundle.Manifest.Platform) {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/platform-unsupported",
			Name:    engine.Name(),
			Message: "bundle platform is not supported by RetroArch",
		}
	}
	if bundle.Manifest.Runtime.Profile == "" {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/profile-required",
			Name:    engine.Name(),
			Message: "runtime profile is required",
		}
	}

	coreName, found := retroArchCoreName(bundle.Manifest.Runtime.Profile)
	if !found {
		return LaunchPlan{}, EngineError{
			Kind:    "engine/profile-unsupported",
			Name:    engine.Name(),
			Message: "runtime profile is not supported by RetroArch",
		}
	}

	entrypoint := bundle.Manifest.Runtime.Entrypoint
	if entrypoint == "" {
		entrypoint = bundle.Manifest.Artefact.Path
	}

	corePath := path.Join(retroArchCoreDirectory(engine), coreName)
	arguments := []string{
		"-L",
		corePath,
		entrypoint,
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

func retroArchCoreDirectory(engine RetroArchEngine) string {
	if engine.CoreDirectory == "" {
		return "cores"
	}

	return engine.CoreDirectory
}

func retroArchCoreName(profile string) (string, bool) {
	switch profile {
	case "genesis":
		return "genesis_plus_gx", true
	case "snes":
		return "snes9x", true
	case "nes":
		return "nestopia", true
	case "game-boy", "game-boy-color":
		return "gambatte", true
	case "game-boy-advance", "gba":
		return "mgba", true
	default:
		return "", false
	}
}
