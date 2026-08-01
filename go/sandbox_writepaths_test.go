package play

import "testing"

func TestSandboxWritePaths_DefaultManifestWritePaths_Good(testingT *testing.T) {
	testingT.Parallel()

	// A manifest with no declared write permissions falls back to the manifest
	// save and screenshot directories.
	manifest := Manifest{}
	manifest.Save.Path = "saves/"
	manifest.Save.Screenshots = "screenshots/"

	paths := manifestLaunchWritePaths(manifest)
	if !sandboxPathAllowed("saves/game.sav", paths) {
		testingT.Fatalf("default write paths missing saves: %v", paths)
	}
	if !sandboxPathAllowed("screenshots/capture.png", paths) {
		testingT.Fatalf("default write paths missing screenshots: %v", paths)
	}
}

func TestSandboxWritePaths_DefaultManifestWritePaths_Ugly(testingT *testing.T) {
	testingT.Parallel()

	// With no save or screenshot overrides the built-in defaults still apply.
	paths := manifestLaunchWritePaths(Manifest{})
	if !sandboxPathAllowed("saves/game.sav", paths) {
		testingT.Fatalf("default write paths missing fallback saves: %v", paths)
	}
	if !sandboxPathAllowed("screenshots/capture.png", paths) {
		testingT.Fatalf("default write paths missing fallback screenshots: %v", paths)
	}
}

func TestSandboxWritePaths_ManifestSavePath_Good(testingT *testing.T) {
	testingT.Parallel()

	manifest := Manifest{}
	manifest.Save.Path = "custom-saves/"
	if got := manifestSavePath(manifest); got != "custom-saves/" {
		testingT.Fatalf("manifestSavePath = %q, want custom-saves/", got)
	}
	if got := manifestSavePath(Manifest{}); got != "saves/" {
		testingT.Fatalf("manifestSavePath default = %q, want saves/", got)
	}
}
