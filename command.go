package play

import "dappco.re/go/core"

// Core Play command names.
const (
	CommandPlay        = "play"
	CommandPlayList    = "play/list"
	CommandPlayVerify  = "play/verify"
	CommandPlayShield  = "play/shield-verify"
	CommandPlayBundle  = "play/bundle"
	CommandPlayEngines = "play/engines"
)

// Commands returns known Core Play command names in a stable order.
func Commands() []string {
	return []string{
		CommandPlay,
		CommandPlayList,
		CommandPlayVerify,
		CommandPlayShield,
		CommandPlayBundle,
		CommandPlayEngines,
	}
}

// Register installs Core Play commands on a Core runtime.
func Register(c *core.Core) {
	if c == nil {
		return
	}

	c.Command(CommandPlay, core.Command{
		Description: "Run preserved software in a STIM bundle",
		Action:      cmdPlay,
	})
	c.Command(CommandPlayList, core.Command{
		Description: "List available STIM bundles",
		Action:      cmdPlayList,
	})
	c.Command(CommandPlayVerify, core.Command{
		Description: "Verify Shield surfaces without running",
		Action:      cmdPlayVerify,
	})
	c.Command(CommandPlayShield, core.Command{
		Description: "Verify Shield surfaces without running",
		Action:      cmdPlayVerify,
	})
	c.Command(CommandPlayBundle, core.Command{
		Description: "Create a STIM bundle from artefact",
		Action:      cmdPlayBundle,
	})
	c.Command(CommandPlayEngines, core.Command{
		Description: "List available runtime engines",
		Action:      cmdPlayEngines,
	})
}

func cmdPlay(opts core.Options) core.Result {
	bundlePath := firstOptionString(opts, "_arg", "bundle", "name", "path")
	if bundlePath == "" {
		bundlePath = "."
	}

	return core.Result{
		Value: PlayRequest{BundlePath: bundlePath},
		OK:    true,
	}
}

func cmdPlayList(opts core.Options) core.Result {
	root := opts.String("root")
	if root == "" {
		root = "."
	}

	return core.Result{
		Value: ListRequest{
			Root: root,
			JSON: opts.Bool("json"),
		},
		OK: true,
	}
}

func cmdPlayVerify(opts core.Options) core.Result {
	bundlePath := firstOptionString(opts, "_arg", "bundle", "name", "path")
	if bundlePath == "" {
		bundlePath = "."
	}

	return core.Result{
		Value: VerifyRequest{BundlePath: bundlePath},
		OK:    true,
	}
}

func cmdPlayBundle(opts core.Options) core.Result {
	return core.Result{
		Value: BundleRequest{
			Name:              opts.String("name"),
			Title:             opts.String("title"),
			Author:            opts.String("author"),
			Year:              opts.Int("year"),
			Platform:          opts.String("platform"),
			Genre:             opts.String("genre"),
			Licence:           opts.String("licence"),
			Engine:            opts.String("engine"),
			Profile:           opts.String("profile"),
			Acceleration:      AccelerationMode(firstOptionString(opts, "acceleration")),
			Filter:            FrameFilter(firstOptionString(opts, "filter")),
			ArtefactPath:      firstOptionString(opts, "rom", "artefact", "artefact_path"),
			ArtefactData:      optionBytes(opts, "artefact_data", "rom_data"),
			ArtefactSHA256:    firstOptionString(opts, "artefact_sha256", "sha256"),
			ArtefactSize:      firstOptionInt64(opts, "artefact_size", "size"),
			ArtefactMediaType: firstOptionString(opts, "artefact_media_type", "media_type"),
			ArtefactSource:    opts.String("source"),
			EngineBinaryPath:  firstOptionString(opts, "engine_binary", "engine-binary"),
			EngineBinaryData:  optionBytes(opts, "engine_binary_data"),
			EngineBinarySHA256: firstOptionString(
				opts,
				"engine_binary_sha256",
				"engine-sha256",
			),
			ResourceLimits: ResourceLimits{
				CPUPercent:  firstOptionInt(opts, "cpu_percent", "cpu-percent"),
				MemoryBytes: firstOptionInt64(opts, "memory_bytes", "memory-bytes"),
			},
			DistributionMode:  firstOptionString(opts, "distribution_mode", "distribution"),
			BYOROM:            opts.Bool("byorom"),
			Entrypoint:        opts.String("entrypoint"),
			RuntimeConfigPath: opts.String("config"),
			VerificationChain: opts.String("chain"),
			SBOMPath:          opts.String("sbom"),
			SavePath:          firstOptionString(opts, "save_path", "save"),
			ScreenshotPath:    firstOptionString(opts, "screenshot_path", "screenshots"),
		},
		OK: true,
	}
}

func cmdPlayEngines(core.Options) core.Result {
	return core.Result{
		Value: RegisteredEngines(),
		OK:    true,
	}
}

// PlayRequest describes a request to prepare a bundle for launch.
type PlayRequest struct {
	BundlePath string
}

// ListRequest describes a request to enumerate bundles.
type ListRequest struct {
	Root string
	JSON bool
}

// VerifyRequest describes a request to verify a bundle.
type VerifyRequest struct {
	BundlePath string
}

// BundleRequest describes the input required to plan a new STIM bundle.
type BundleRequest struct {
	Name               string
	Title              string
	Author             string
	Year               int
	Platform           string
	Genre              string
	Licence            string
	Engine             string
	Profile            string
	Acceleration       AccelerationMode
	Filter             FrameFilter
	ArtefactPath       string
	ArtefactData       []byte
	ArtefactSHA256     string
	ArtefactSize       int64
	ArtefactMediaType  string
	ArtefactSource     string
	EngineBinaryPath   string
	EngineBinaryData   []byte
	EngineBinarySHA256 string
	ResourceLimits     ResourceLimits
	DistributionMode   string
	BYOROM             bool
	Entrypoint         string
	RuntimeConfigPath  string
	VerificationChain  string
	SBOMPath           string
	SavePath           string
	ScreenshotPath     string
}

func firstOptionString(opts core.Options, keys ...string) string {
	for _, key := range keys {
		if value := opts.String(key); value != "" {
			return value
		}
	}

	return ""
}

func firstOptionInt(opts core.Options, keys ...string) int {
	for _, key := range keys {
		if value := opts.Int(key); value != 0 {
			return value
		}
	}

	return 0
}

func firstOptionInt64(opts core.Options, keys ...string) int64 {
	for _, key := range keys {
		result := opts.Get(key)
		if !result.OK {
			continue
		}

		switch value := result.Value.(type) {
		case int64:
			if value != 0 {
				return value
			}
		case int:
			if value != 0 {
				return int64(value)
			}
		}
	}

	return 0
}

func optionBytes(opts core.Options, keys ...string) []byte {
	for _, key := range keys {
		result := opts.Get(key)
		if !result.OK {
			continue
		}

		data, ok := result.Value.([]byte)
		if ok && len(data) > 0 {
			return cloneBytes(data)
		}
	}

	return nil
}
