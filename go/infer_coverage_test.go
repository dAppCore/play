package play

import "testing"

func TestInferCoverage_EngineForPlatform_Good(testingT *testing.T) {
	testingT.Parallel()

	cases := map[string]string{
		"dos":              "dosbox",
		"pc-98":            "dosbox-x",
		"windows-3x":       "dosbox-x",
		"windows-9x":       "dosbox-x",
		"sega-genesis":     "retroarch",
		"sega-mega-drive":  "retroarch",
		"snes":             "retroarch",
		"super-nintendo":   "retroarch",
		"nes":              "retroarch",
		"game-boy":         "retroarch",
		"game-boy-color":   "retroarch",
		"game-boy-advance": "retroarch",
		"gba":              "retroarch",
		"scummvm":          "scummvm",
		"point-and-click":  "scummvm",
		"arcade":           "mame",
		"neo-geo":          "mame",
		"commodore-64":     "vice",
		"c64":              "vice",
		"commodore-128":    "vice",
		"c128":             "vice",
		"vic-20":           "vice",
		"zx-spectrum":      "fuse",
		"spectrum":         "fuse",
		"zx-spectrum-48k":  "fuse",
		"zx-spectrum-128k": "fuse",
	}
	for platform, want := range cases {
		if got := inferEngineForPlatform(platform); got != want {
			testingT.Fatalf("inferEngineForPlatform(%q) = %q, want %q", platform, got, want)
		}
	}
}

func TestInferCoverage_EngineForPlatform_Ugly(testingT *testing.T) {
	testingT.Parallel()

	if got := inferEngineForPlatform("unknown-platform"); got != "" {
		testingT.Fatalf("inferEngineForPlatform(unknown) = %q, want empty", got)
	}
}

func TestInferCoverage_ProfileForPlatform_Good(testingT *testing.T) {
	testingT.Parallel()

	type entry struct {
		platform string
		engine   string
		want     string
	}
	cases := []entry{
		{"", "dosbox", "dos"},
		{"pc-98", "dosbox-x", "pc-98"},
		{"windows-3x", "dosbox-x", "windows-3x"},
		{"windows-9x", "dosbox-x", "windows-9x"},
		{"dos", "dosbox-x", "dos"},
		{"sega-genesis", "retroarch", "genesis"},
		{"sega-mega-drive", "retroarch", "genesis"},
		{"snes", "retroarch", "snes"},
		{"super-nintendo", "retroarch", "snes"},
		{"nes", "retroarch", "nes"},
		{"game-boy", "retroarch", "game-boy"},
		{"game-boy-color", "retroarch", "game-boy-color"},
		{"game-boy-advance", "retroarch", "gba"},
		{"gba", "retroarch", "gba"},
		{"commodore-128", "vice", "c128"},
		{"c128", "vice", "c128"},
		{"vic-20", "vice", "vic-20"},
		{"commodore-64", "vice", "c64"},
		{"zx-spectrum-128k", "fuse", "128k"},
		{"zx-spectrum", "fuse", "48k"},
	}
	for _, candidate := range cases {
		if got := inferProfileForPlatform(candidate.platform, candidate.engine); got != candidate.want {
			testingT.Fatalf("inferProfileForPlatform(%q, %q) = %q, want %q",
				candidate.platform, candidate.engine, got, candidate.want)
		}
	}
}

func TestInferCoverage_ProfileForPlatform_Ugly(testingT *testing.T) {
	testingT.Parallel()

	// Unknown retroarch platform and unknown engine both fall through to empty.
	if got := inferProfileForPlatform("unknown", "retroarch"); got != "" {
		testingT.Fatalf("inferProfileForPlatform(unknown, retroarch) = %q, want empty", got)
	}
	if got := inferProfileForPlatform("anything", "mame"); got != "" {
		testingT.Fatalf("inferProfileForPlatform(anything, mame) = %q, want empty", got)
	}
}

func TestInferCoverage_AccelerationMode_Good(testingT *testing.T) {
	testingT.Parallel()

	if got := inferAccelerationMode(AccelerationOff, "retroarch"); got != AccelerationOff {
		testingT.Fatalf("inferAccelerationMode(off) = %q, want off", got)
	}
	if got := inferAccelerationMode("", ""); got != AccelerationAuto {
		testingT.Fatalf("inferAccelerationMode(empty, empty) = %q, want auto", got)
	}
	if got := inferAccelerationMode("", "retroarch"); got != AccelerationAuto {
		testingT.Fatalf("inferAccelerationMode(empty, retroarch) = %q, want auto", got)
	}
}

func TestInferCoverage_FrameFilter_Good(testingT *testing.T) {
	testingT.Parallel()

	if got := inferFrameFilter(FrameFilterBilinear, "retroarch"); got != FrameFilterBilinear {
		testingT.Fatalf("inferFrameFilter(bilinear) = %q, want bilinear", got)
	}
	for _, engine := range []string{"dosbox", "dosbox-x", "retroarch", "scummvm", "mame", "vice", "fuse", "snes9x"} {
		if got := inferFrameFilter("", engine); got != FrameFilterNearest {
			testingT.Fatalf("inferFrameFilter(empty, %q) = %q, want nearest", engine, got)
		}
	}
	if got := inferFrameFilter("", "unknown"); got != FrameFilterNone {
		testingT.Fatalf("inferFrameFilter(empty, unknown) = %q, want none", got)
	}
}
