package play

// RuntimeDefaults describes inferred runtime values for bundle planning.
type RuntimeDefaults struct {
	Engine       string
	Profile      string
	Acceleration AccelerationMode
	Filter       FrameFilter
}

func inferRuntimeDefaults(request BundleRequest) RuntimeDefaults {
	engine := request.Engine
	if engine == "" {
		engine = inferEngineForPlatform(request.Platform)
	}

	profile := request.Profile
	if profile == "" {
		profile = inferProfileForPlatform(request.Platform, engine)
	}

	return RuntimeDefaults{
		Engine:       engine,
		Profile:      profile,
		Acceleration: inferAccelerationMode(request.Acceleration, engine),
		Filter:       inferFrameFilter(request.Filter, engine),
	}
}

func inferEngineForPlatform(platform string) string {
	switch platform {
	case "dos":
		return "dosbox"
	case "sega-genesis", "sega-mega-drive":
		return "retroarch"
	case "snes", "super-nintendo":
		return "retroarch"
	case "nes":
		return "retroarch"
	case "game-boy", "game-boy-color":
		return "retroarch"
	case "game-boy-advance", "gba":
		return "retroarch"
	case "scummvm", "point-and-click":
		return "scummvm"
	default:
		return ""
	}
}

func inferProfileForPlatform(platform string, engine string) string {
	switch engine {
	case "dosbox":
		return "dos"
	case "retroarch":
		switch platform {
		case "sega-genesis", "sega-mega-drive":
			return "genesis"
		case "snes", "super-nintendo":
			return "snes"
		case "nes":
			return "nes"
		case "game-boy":
			return "game-boy"
		case "game-boy-color":
			return "game-boy-color"
		case "game-boy-advance", "gba":
			return "gba"
		default:
			return ""
		}
	default:
		return ""
	}
}

func inferAccelerationMode(requested AccelerationMode, engine string) AccelerationMode {
	if requested != "" && requested.valid() {
		return requested
	}
	if engine == "" {
		return AccelerationAuto
	}

	return AccelerationAuto
}

func inferFrameFilter(requested FrameFilter, engine string) FrameFilter {
	if requested != "" && requested.valid() {
		return requested
	}

	switch engine {
	case "dosbox":
		return FrameFilterNearest
	case "retroarch":
		return FrameFilterNearest
	case "scummvm":
		return FrameFilterNearest
	default:
		return FrameFilterNone
	}
}
