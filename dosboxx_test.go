package play

import (
	"testing"
	"testing/fstest"
)

func TestDosboxx_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := DOSBoxXEngine{Binary: "dosbox-x"}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
	if engine.Acceleration().Mode != AccelerationAuto {
		testingT.Fatalf("unexpected acceleration mode: %q", engine.Acceleration().Mode)
	}
	if len(engine.Platforms()) != 4 {
		testingT.Fatalf("unexpected platform count: %d", len(engine.Platforms()))
	}
}

func TestDosboxx_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := DOSBoxXEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestDosboxx_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(dosBoxXBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	bundle.Manifest.Runtime.Profile = "unknown-machine"
	_, err = DOSBoxXEngine{Binary: "dosbox-x"}.PlanLaunch(bundle)
	if err == nil {
		testingT.Fatal("PlanLaunch expected an error for an unsupported profile")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("PlanLaunch returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/profile-unsupported" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}

func TestDosboxx_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(dosBoxXBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := DOSBoxXEngine{Binary: "dosbox-x"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "dosbox-x" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) != 6 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "-conf" || plan.Arguments[1] != "emulator.yaml" {
		testingT.Fatalf("unexpected config arguments: %v", plan.Arguments)
	}
	if plan.Arguments[2] != "-set" || plan.Arguments[3] != "dosbox machine=pc98" {
		testingT.Fatalf("unexpected machine arguments: %v", plan.Arguments)
	}
	if plan.Arguments[4] != "-c" || plan.Arguments[5] != "BOOT rom/pc98.hdi" {
		testingT.Fatalf("unexpected boot arguments: %v", plan.Arguments)
	}
}

func dosBoxXBundleFS() fstest.MapFS {
	romData := []byte("pc98-image")
	emulatorData := []byte("engine: dosbox-x\nprofile: pc-98\n")
	manifestData := []byte(`name: pc98-sample
title: "PC-98 Sample"
platform: pc-98
licence: freeware
artefact:
  path: rom/pc98.hdi
  sha256: "` + hashHex(romData) + `"
runtime:
  engine: dosbox-x
  profile: pc-98
  config: emulator.yaml
  entrypoint: rom/pc98.hdi
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
			hashHex(romData) + "  rom/pc98.hdi\n",
	)

	return fstest.MapFS{
		"manifest.yaml":    {Data: manifestData},
		"emulator.yaml":    {Data: emulatorData},
		"checksums.sha256": {Data: checksumData},
		"sbom.json":        {Data: sbomData},
		"rom/pc98.hdi":     {Data: romData},
	}
}
