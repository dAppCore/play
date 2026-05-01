package play

import (
	"context"
	"path"
	"strconv"

	core "dappco.re/go"
)

// ScummVMEngine is a first-pass ScummVM adapter scaffold.
type ScummVMEngine struct {
	Binary       string
	BinarySHA256 string
	Core         *core.Core
	GameID       string
	SavePath     string
	AudioRate    int
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
	if engine.Core == nil || !engine.Core.Process().Exists() {
		return nil
	}

	result := engine.Core.Process().Run(context.Background(), engine.Binary, "--version")
	if !result.OK {
		return EngineError{
			Kind:    "engine/version-unavailable",
			Name:    engine.Name(),
			Message: resultMessage(result.Value),
		}
	}
	output, ok := result.Value.(string)
	if !ok || !scummVMVersionAtLeast(output, 2, 7) {
		return EngineError{
			Kind:    "engine/version-unsupported",
			Name:    engine.Name(),
			Message: "ScummVM version must be 2.7 or newer",
		}
	}

	return nil
}

// CodeIdentity returns the ScummVM runtime integrity identity.
func (engine ScummVMEngine) CodeIdentity() EngineCodeIdentity {
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

// Run executes a ScummVM artefact through Core's process primitive.
func (engine ScummVMEngine) Run(artefact string, config EngineConfig) error {
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

	arguments := scummVMArguments(scummVMConfig{
		GameID:    defaultString(engine.GameID, config.Profile),
		DataPath:  path.Dir(artefact),
		SavePath:  defaultString(engine.SavePath, config.SaveRoot),
		AudioRate: engine.AudioRate,
	})

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

	arguments := scummVMArguments(scummVMConfig{
		GameID:    bundle.Manifest.Runtime.Profile,
		DataPath:  dataPath,
		SavePath:  bundle.Manifest.Save.Path,
		AudioRate: engine.AudioRate,
	})

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

type scummVMConfig struct {
	GameID    string
	DataPath  string
	SavePath  string
	AudioRate int
}

func scummVMArguments(config scummVMConfig) []string {
	arguments := []string{
		"--path=" + config.DataPath,
	}
	if config.SavePath != "" {
		arguments = append(arguments, "--savepath="+config.SavePath)
	}
	if config.AudioRate > 0 {
		arguments = append(arguments, "--output-rate="+strconv.Itoa(config.AudioRate))
	}
	arguments = append(arguments, config.GameID)

	return arguments
}

func scummVMVersionAtLeast(output string, wantedMajor int, wantedMinor int) bool {
	normalised := core.Replace(output, "\n", " ")
	for _, field := range core.Split(normalised, " ") {
		major, minor, ok := scummVMVersionPair(field)
		if !ok {
			continue
		}
		if major > wantedMajor {
			return true
		}
		if major == wantedMajor && minor >= wantedMinor {
			return true
		}

		return false
	}

	return false
}

func scummVMVersionPair(value string) (int, int, bool) {
	start := -1
	for index := 0; index < len(value); index++ {
		if value[index] >= '0' && value[index] <= '9' {
			start = index
			break
		}
	}
	if start < 0 {
		return 0, 0, false
	}

	parts := core.Split(value[start:], ".")
	if len(parts) < 2 {
		return 0, 0, false
	}

	major, majorErr := strconv.Atoi(numericPrefix(parts[0]))
	minor, minorErr := strconv.Atoi(numericPrefix(parts[1]))
	if majorErr != nil || minorErr != nil {
		return 0, 0, false
	}

	return major, minor, true
}

func numericPrefix(value string) string {
	end := 0
	for end < len(value) {
		if value[end] < '0' || value[end] > '9' {
			break
		}
		end++
	}

	return value[:end]
}
