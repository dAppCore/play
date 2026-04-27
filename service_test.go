package play

import (
	"testing"
	"testing/fstest"
)

func TestService_ListBundles_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(listBundleFS(), NewRegistry())
	summaries, err := service.ListBundles(ListRequest{Root: "."})
	if err != nil {
		testingT.Fatalf("ListBundles returned error: %v", err)
	}

	if len(summaries) != 2 {
		testingT.Fatalf("unexpected bundle count: %d", len(summaries))
	}
	if summaries[0].Path != "command-and-conquer" {
		testingT.Fatalf("unexpected first bundle path: %q", summaries[0].Path)
	}
}

func TestService_ListBundles_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, NewRegistry())
	_, err := service.ListBundles(ListRequest{Root: "."})
	if err == nil {
		testingT.Fatal("ListBundles expected an error for a missing filesystem")
	}
}

func TestService_ListBundles_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(listBundleFS(), NewRegistry())
	_, err := service.ListBundles(ListRequest{Root: "../escape"})
	if err == nil {
		testingT.Fatal("ListBundles expected an error for an invalid root path")
	}

	pathError, ok := err.(PathError)
	if !ok {
		testingT.Fatalf("ListBundles returned %T, want PathError", err)
	}
	if pathError.Kind != "bundle/path-invalid" {
		testingT.Fatalf("unexpected path error kind: %q", pathError.Kind)
	}
}

func TestService_VerifyBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	service := NewService(verifiedBundleFS(), registry)
	result, err := service.VerifyBundle(VerifyRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("VerifyBundle returned error: %v", err)
	}

	if !result.Verified {
		testingT.Fatalf("VerifyBundle expected a verified result: %v", result.Issues)
	}
}

func TestService_VerifyBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, NewRegistry())
	_, err := service.VerifyBundle(VerifyRequest{BundlePath: "."})
	if err == nil {
		testingT.Fatal("VerifyBundle expected an error for a missing filesystem")
	}
}

func TestService_VerifyBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	service := NewService(verifiedBundleFS(), registry)
	result, err := service.VerifyBundle(VerifyRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("VerifyBundle returned error: %v", err)
	}

	if !hasIssueCode(result.Issues, "engine/unavailable") {
		testingT.Fatalf("VerifyBundle missing engine/unavailable issue: %v", result.Issues)
	}
}

func TestService_PreparePlay_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	service := NewService(verifiedBundleFS(), registry)
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("PreparePlay returned error: %v", err)
	}

	if !plan.Ready {
		testingT.Fatalf("PreparePlay expected a ready plan: %v", plan.Issues)
	}
	if plan.Engine == nil {
		testingT.Fatal("PreparePlay expected a resolved engine")
	}
}

func TestService_PreparePlayDOSBox_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(DOSBoxEngine{Binary: "dosbox"}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	service := NewService(dosBundleFS(), registry)
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("PreparePlay returned error: %v", err)
	}

	if plan.Launch == nil {
		testingT.Fatal("PreparePlay expected a launch plan for DOSBox")
	}
	if plan.Launch.Executable != "dosbox" {
		testingT.Fatalf("unexpected launch executable: %q", plan.Launch.Executable)
	}
}

func TestService_PreparePlayRetroArch_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(RetroArchEngine{Binary: "retroarch"}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	service := NewService(verifiedBundleFS(), registry)
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("PreparePlay returned error: %v", err)
	}

	if plan.Launch == nil {
		testingT.Fatal("PreparePlay expected a launch plan for RetroArch")
	}
	if plan.Launch.Executable != "retroarch" {
		testingT.Fatalf("unexpected launch executable: %q", plan.Launch.Executable)
	}
	if len(plan.Launch.Arguments) < 3 {
		testingT.Fatalf("unexpected launch arguments: %v", plan.Launch.Arguments)
	}
}

func TestService_PreparePlayScummVM_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(ScummVMEngine{Binary: "scummvm"}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	service := NewService(scummVMBundleFS(), registry)
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("PreparePlay returned error: %v", err)
	}

	if plan.Launch == nil {
		testingT.Fatal("PreparePlay expected a launch plan for ScummVM")
	}
	if plan.Launch.Executable != "scummvm" {
		testingT.Fatalf("unexpected launch executable: %q", plan.Launch.Executable)
	}
	if len(plan.Launch.Arguments) != 3 {
		testingT.Fatalf("unexpected launch arguments: %v", plan.Launch.Arguments)
	}
}

func TestService_PreparePlayDOSBoxX_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(DOSBoxXEngine{Binary: "dosbox-x"}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	service := NewService(dosBoxXBundleFS(), registry)
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("PreparePlay returned error: %v", err)
	}

	if plan.Launch == nil {
		testingT.Fatal("PreparePlay expected a launch plan for DOSBox-X")
	}
	if plan.Launch.Executable != "dosbox-x" {
		testingT.Fatalf("unexpected launch executable: %q", plan.Launch.Executable)
	}
	if len(plan.Launch.Arguments) != 6 {
		testingT.Fatalf("unexpected launch arguments: %v", plan.Launch.Arguments)
	}
}

func TestService_PreparePlay_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, NewRegistry())
	_, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err == nil {
		testingT.Fatal("PreparePlay expected an error for a missing filesystem")
	}
}

func TestService_PreparePlay_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(verifiedBundleFS(), NewRegistry())
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	if err != nil {
		testingT.Fatalf("PreparePlay returned error: %v", err)
	}

	if plan.Ready {
		testingT.Fatalf("PreparePlay expected a non-ready plan: %v", plan.Issues)
	}
	if !hasIssueCode(plan.Issues, "engine/unavailable") {
		testingT.Fatalf("PreparePlay missing engine/unavailable issue: %v", plan.Issues)
	}
}

func TestService_PlanBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "mega-lo-mania",
		Title:          "Mega lo Mania",
		Platform:       "sega-genesis",
		Licence:        "freeware",
		Engine:         "retroarch",
		Profile:        "genesis",
		ArtefactPath:   "rom/MegaLoMania.zip",
		ArtefactSHA256: validArtefactSHA256,
		ResourceLimits: ResourceLimits{
			CPUPercent:  75,
			MemoryBytes: 268435456,
		},
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Config != "emulator.yaml" {
		testingT.Fatalf("unexpected runtime config: %q", plan.Manifest.Runtime.Config)
	}
	if plan.Manifest.FormatVersion != CurrentManifestFormatVersion {
		testingT.Fatalf("unexpected format version: %q", plan.Manifest.FormatVersion)
	}
	if plan.Manifest.Runtime.Acceleration != AccelerationAuto {
		testingT.Fatalf("unexpected runtime acceleration: %q", plan.Manifest.Runtime.Acceleration)
	}
	if plan.Manifest.Runtime.Filter != FrameFilterNearest {
		testingT.Fatalf("unexpected runtime filter: %q", plan.Manifest.Runtime.Filter)
	}
	if plan.Manifest.Resources.MemoryBytes != 268435456 {
		testingT.Fatalf("unexpected memory resource limit: %d", plan.Manifest.Resources.MemoryBytes)
	}
}

func TestService_PlanBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	_, issues := service.PlanBundle(BundleRequest{})
	if !issues.HasIssues() {
		testingT.Fatal("PlanBundle expected issues for an empty request")
	}
}

func TestService_PlanBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	_, issues := service.PlanBundle(BundleRequest{
		Name:           "mega-lo-mania",
		Title:          "Mega lo Mania",
		Platform:       "sega-genesis",
		Licence:        "freeware",
		Engine:         "retroarch",
		ArtefactPath:   "../rom/MegaLoMania.zip",
		ArtefactSHA256: validArtefactSHA256,
		BYOROM:         true,
	})
	if !hasIssueCode(issues, "manifest/artefact-path-invalid") {
		testingT.Fatalf("PlanBundle missing manifest/artefact-path-invalid issue: %v", issues)
	}
}

func listBundleFS() fstest.MapFS {
	commandManifest := []byte(`name: command-and-conquer
title: "Command & Conquer"
platform: dos
licence: freeware
artefact:
  path: rom/CNC.zip
  sha256: "` + validArtefactSHA256 + `"
runtime:
  engine: dosbox
  config: emulator.yaml
verification:
  chain: checksums.sha256
  sbom: sbom.json
  deterministic: true
permissions:
  network: false
  microphone: false
  filesystem:
    read:
      - rom/
    write:
      - saves/
save:
  path: saves/
distribution:
  mode: catalogue
`)

	return fstest.MapFS{
		"mega-lo-mania/manifest.yaml":          {Data: []byte(validManifestYAML())},
		"mega-lo-mania/emulator.yaml":          {Data: []byte("engine: retroarch\nprofile: genesis\n")},
		"mega-lo-mania/checksums.sha256":       {Data: []byte("placeholder")},
		"mega-lo-mania/sbom.json":              {Data: []byte("{}")},
		"mega-lo-mania/rom/MegaLoMania.zip":    {Data: []byte("rom")},
		"command-and-conquer/manifest.yaml":    {Data: commandManifest},
		"command-and-conquer/emulator.yaml":    {Data: []byte("engine: dosbox\n")},
		"command-and-conquer/checksums.sha256": {Data: []byte("placeholder")},
		"command-and-conquer/sbom.json":        {Data: []byte("{}")},
		"command-and-conquer/rom/CNC.zip":      {Data: []byte("rom")},
		"notes.txt":                            {Data: []byte("ignore")},
	}
}

func scummVMBundleFS() fstest.MapFS {
	artefactData := []byte("sky-disk")
	emulatorData := []byte("engine: scummvm\nprofile: sky\n")
	manifestData := []byte(`name: beneath-a-steel-sky
title: "Beneath a Steel Sky"
platform: scummvm
licence: freeware
artefact:
  path: game/BASS/sky.dsk
  sha256: "` + hashHex(artefactData) + `"
runtime:
  engine: scummvm
  profile: sky
  config: emulator.yaml
  entrypoint: game/BASS/sky.dsk
verification:
  chain: checksums.sha256
  sbom: sbom.json
  deterministic: true
permissions:
  network: false
  microphone: false
  filesystem:
    read:
      - game/
    write:
      - saves/
      - screenshots/
save:
  path: saves/
  screenshots: screenshots/
distribution:
  mode: catalogue
`)
	manifest, err := LoadManifest(manifestData)
	if err != nil {
		panic(err)
	}
	sbomData, err := BuildSBOM(manifest)
	if err != nil {
		panic(err)
	}
	checksumData := []byte(
		hashHex(manifestData) + "  manifest.yaml\n" +
			hashHex(emulatorData) + "  emulator.yaml\n" +
			hashHex(sbomData) + "  sbom.json\n" +
			hashHex(artefactData) + "  game/BASS/sky.dsk\n",
	)

	return fstest.MapFS{
		"manifest.yaml":     {Data: manifestData},
		"emulator.yaml":     {Data: emulatorData},
		"checksums.sha256":  {Data: checksumData},
		"sbom.json":         {Data: sbomData},
		"game/BASS/sky.dsk": {Data: artefactData},
	}
}
