package play

import (
	"testing"
	"testing/fstest"
)

func TestRender_BundlePlan_Good(testingT *testing.T) {
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
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	rendered, err := plan.Render()
	if err != nil {
		testingT.Fatalf("Render returned error: %v", err)
	}

	if rendered.Path != "mega-lo-mania" {
		testingT.Fatalf("unexpected rendered path: %q", rendered.Path)
	}
	if len(rendered.Files) != 4 {
		testingT.Fatalf("unexpected rendered file count: %d", len(rendered.Files))
	}

	manifestData := renderedFileData(rendered, "manifest.yaml")
	manifest, err := LoadManifest(manifestData)
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}
	if manifest.Name != "mega-lo-mania" {
		testingT.Fatalf("unexpected manifest name: %q", manifest.Name)
	}

	checksumData := string(renderedFileData(rendered, "checksums.sha256"))
	if !containsLine(checksumData, validArtefactSHA256+"  rom/MegaLoMania.zip") {
		testingT.Fatalf("checksum output missing artefact line: %q", checksumData)
	}
	runtimeConfigData := string(renderedFileData(rendered, "emulator.yaml"))
	if !containsLine(runtimeConfigData, "acceleration: auto") {
		testingT.Fatalf("runtime config missing acceleration: %q", runtimeConfigData)
	}
	if !containsLine(runtimeConfigData, "filter: nearest") {
		testingT.Fatalf("runtime config missing filter: %q", runtimeConfigData)
	}

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	bundle, err := LoadBundle(renderedBundleFS(rendered, "rom/MegaLoMania.zip", []byte("rom")), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	verifyIssues := bundle.VerifyWithRegistry(registry)
	if verifyIssues.HasIssues() {
		testingT.Fatalf("VerifyWithRegistry returned issues: %v", verifyIssues)
	}
}

func TestRender_BundlePlan_Bad(testingT *testing.T) {
	testingT.Parallel()

	plan := BundlePlan{}
	_, err := plan.Render()
	if err == nil {
		testingT.Fatal("Render expected an error for an invalid plan")
	}
}

func TestRender_BundlePlan_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:              "command-and-conquer",
		Title:             "Command & Conquer",
		Platform:          "dos",
		Licence:           "freeware",
		Engine:            "dosbox",
		Profile:           "dos",
		ArtefactPath:      "rom/CNC.zip",
		ArtefactSHA256:    validArtefactSHA256,
		RuntimeConfigPath: "config/emulator.yaml",
		SBOMPath:          "meta/sbom.json",
		VerificationChain: "meta/checksums.sha256",
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	rendered, err := plan.Render()
	if err != nil {
		testingT.Fatalf("Render returned error: %v", err)
	}

	if renderedFileData(rendered, "config/emulator.yaml") == nil {
		testingT.Fatal("Render missing custom runtime config path")
	}
	if renderedFileData(rendered, "meta/sbom.json") == nil {
		testingT.Fatal("Render missing custom SBOM path")
	}
	if renderedFileData(rendered, "meta/checksums.sha256") == nil {
		testingT.Fatal("Render missing custom checksum path")
	}
}

func TestRender_Service_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(BundleRequest{
		Name:           "mega-lo-mania",
		Title:          "Mega lo Mania",
		Platform:       "sega-genesis",
		Licence:        "freeware",
		Engine:         "retroarch",
		Profile:        "genesis",
		ArtefactPath:   "rom/MegaLoMania.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}

	if len(rendered.Files) != 4 {
		testingT.Fatalf("unexpected rendered file count: %d", len(rendered.Files))
	}
}

func TestRender_Service_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	_, err := service.RenderBundle(BundleRequest{})
	if err == nil {
		testingT.Fatal("RenderBundle expected an error for an invalid request")
	}
}

func TestRender_Service_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(BundleRequest{
		Name:           "quoted-title",
		Title:          `Mega "Quoted" Game`,
		Platform:       "dos",
		Licence:        "freeware",
		Engine:         "dosbox",
		ArtefactPath:   "rom/game.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}

	sbomData := string(renderedFileData(rendered, "sbom.json"))
	if !containsLine(sbomData, `"name":"quoted-title"`) {
		testingT.Fatalf("SBOM output missing quoted bundle name: %q", sbomData)
	}
}

func renderedFileData(rendered RenderedBundle, wantedPath string) []byte {
	for _, file := range rendered.Files {
		if file.Path == wantedPath {
			return file.Data
		}
	}

	return nil
}

func renderedBundleFS(rendered RenderedBundle, artefactPath string, artefactData []byte) fstest.MapFS {
	filesystem := fstest.MapFS{
		artefactPath: &fstest.MapFile{Data: artefactData},
	}

	for _, file := range rendered.Files {
		filesystem[file.Path] = &fstest.MapFile{
			Data: file.Data,
		}
	}

	return filesystem
}

func containsLine(content string, wanted string) bool {
	return len(content) >= len(wanted) && indexOf(content, wanted) >= 0
}

func indexOf(content string, wanted string) int {
	contentLength := len(content)
	wantedLength := len(wanted)
	if wantedLength == 0 {
		return 0
	}
	if wantedLength > contentLength {
		return -1
	}

	for index := 0; index+wantedLength <= contentLength; index++ {
		if content[index:index+wantedLength] == wanted {
			return index
		}
	}

	return -1
}
