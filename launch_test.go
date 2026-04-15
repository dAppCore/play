package play

import (
	"testing"
	"testing/fstest"
)

func TestLaunch_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(dosBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := PlanLaunch(bundle, DOSBoxEngine{Binary: "dosbox"})
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "dosbox" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if plan.Executable != "dosbox" {
		testingT.Fatalf("unexpected executable: %q", plan.Executable)
	}
}

func TestLaunch_PlanLaunch_Bad(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(dosBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	_, err = PlanLaunch(bundle, stubEngine{name: "dosbox", platforms: []string{"dos"}})
	if err == nil {
		testingT.Fatal("PlanLaunch expected an error for a non-planning engine")
	}
}

func TestLaunch_PlanLaunch_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(dosBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	_, err = PlanLaunch(bundle, DOSBoxEngine{})
	if err == nil {
		testingT.Fatal("PlanLaunch expected an error for an unconfigured DOSBox engine")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("PlanLaunch returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/binary-required" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}

func dosBundleFS() fstest.MapFS {
	romData := []byte("dos-rom")
	emulatorData := []byte("engine: dosbox\nprofile: dos\n")
	sbomData := []byte("{\"bomFormat\":\"CycloneDX\"}")
	manifestData := []byte(`name: command-and-conquer
title: "Command & Conquer"
platform: dos
licence: freeware
artefact:
  path: rom/CNC.zip
  sha256: "` + hashHex(romData) + `"
runtime:
  engine: dosbox
  profile: dos
  config: emulator.yaml
  entrypoint: rom/CNC.zip
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
      - screenshots/
save:
  path: saves/
  screenshots: screenshots/
distribution:
  mode: catalogue
`)
	checksumData := []byte(
		hashHex(manifestData) + "  manifest.yaml\n" +
			hashHex(emulatorData) + "  emulator.yaml\n" +
			hashHex(sbomData) + "  sbom.json\n" +
			hashHex(romData) + "  rom/CNC.zip\n",
	)

	return fstest.MapFS{
		"manifest.yaml":    {Data: manifestData},
		"emulator.yaml":    {Data: emulatorData},
		"checksums.sha256": {Data: checksumData},
		"sbom.json":        {Data: sbomData},
		"rom/CNC.zip":      {Data: romData},
	}
}
