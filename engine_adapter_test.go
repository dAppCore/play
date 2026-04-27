package play

import (
	"path"
	"testing"
	"testing/fstest"
)

func TestEngineAdapter_ValidateBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	err = validateAdapterBundle("retroarch", []string{"sega-genesis"}, bundle)
	if err != nil {
		testingT.Fatalf("validateAdapterBundle returned error: %v", err)
	}
}

func TestEngineAdapter_ValidateBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	err = validateAdapterBundle("mame", []string{"arcade"}, bundle)
	if err == nil {
		testingT.Fatal("validateAdapterBundle expected a runtime mismatch")
	}
}

func TestEngineAdapter_ValidateBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	err = validateAdapterBundle("retroarch", []string{"arcade"}, bundle)
	if err == nil {
		testingT.Fatal("validateAdapterBundle expected an unsupported platform")
	}
}

type adapterBundleFixture struct {
	Name         string
	Title        string
	Platform     string
	Engine       string
	Profile      string
	ArtefactPath string
	ArtefactData []byte
	Entrypoint   string
}

func adapterBundleFS(testingT *testing.T, fixture adapterBundleFixture) fstest.MapFS {
	testingT.Helper()

	name := defaultString(fixture.Name, fixture.Engine+"-sample")
	title := defaultString(fixture.Title, name)
	artefactPath := defaultString(fixture.ArtefactPath, "rom/artefact.bin")
	artefactData := fixture.ArtefactData
	if len(artefactData) == 0 {
		artefactData = []byte("artefact")
	}
	readPath := path.Dir(artefactPath)
	if readPath == "." {
		readPath = "rom"
	}
	readPath += "/"

	profileLine := ""
	if fixture.Profile != "" {
		profileLine = "  profile: " + fixture.Profile + "\n"
	}
	entrypointLine := ""
	if fixture.Entrypoint != "" {
		entrypointLine = "  entrypoint: " + fixture.Entrypoint + "\n"
	}

	emulatorData := []byte("engine: " + fixture.Engine + "\n")
	if fixture.Profile != "" {
		emulatorData = append(emulatorData, []byte("profile: "+fixture.Profile+"\n")...)
	}
	manifestData := []byte(`name: ` + name + `
title: "` + title + `"
platform: ` + fixture.Platform + `
licence: freeware
artefact:
  path: ` + artefactPath + `
  sha256: "` + hashHex(artefactData) + `"
runtime:
  engine: ` + fixture.Engine + `
` + profileLine + `  config: emulator.yaml
` + entrypointLine + `verification:
  chain: checksums.sha256
  sbom: sbom.json
  deterministic: true
permissions:
  network: false
  microphone: false
  filesystem:
    read:
      - ` + readPath + `
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
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}
	sbomData, err := BuildSBOM(manifest)
	if err != nil {
		testingT.Fatalf("BuildSBOM returned error: %v", err)
	}
	checksumData := []byte(
		hashHex(manifestData) + "  manifest.yaml\n" +
			hashHex(emulatorData) + "  emulator.yaml\n" +
			hashHex(sbomData) + "  sbom.json\n" +
			hashHex(artefactData) + "  " + artefactPath + "\n",
	)

	return fstest.MapFS{
		"manifest.yaml":    {Data: manifestData},
		"emulator.yaml":    {Data: emulatorData},
		"checksums.sha256": {Data: checksumData},
		"sbom.json":        {Data: sbomData},
		artefactPath:       {Data: artefactData},
	}
}
