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
	case "pc-98", "windows-3x", "windows-9x":
		return "dosbox-x"
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
	case "arcade", "neo-geo":
		return "mame"
	case "commodore-64", "c64", "commodore-128", "c128", "vic-20":
		return "vice"
	case "zx-spectrum", "spectrum", "zx-spectrum-48k", "zx-spectrum-128k":
		return "fuse"
	default:
		return ""
	}
}

func inferProfileForPlatform(platform string, engine string) string {
	switch engine {
	case "dosbox":
		return "dos"
	case "dosbox-x":
		switch platform {
		case "pc-98":
			return "pc-98"
		case "windows-3x":
			return "windows-3x"
		case "windows-9x":
			return "windows-9x"
		default:
			return "dos"
		}
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
	case "vice":
		switch platform {
		case "commodore-128", "c128":
			return "c128"
		case "vic-20":
			return "vic-20"
		default:
			return "c64"
		}
	case "fuse":
		switch platform {
		case "zx-spectrum-128k":
			return "128k"
		default:
			return "48k"
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
	case "dosbox", "dosbox-x":
		return FrameFilterNearest
	case "retroarch":
		return FrameFilterNearest
	case "scummvm":
		return FrameFilterNearest
	case "mame", "vice", "fuse", "snes9x":
		return FrameFilterNearest
	default:
		return FrameFilterNone
	}
}
