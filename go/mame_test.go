package play

import (
	"testing"
	"testing/fstest"
)

func TestMame_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := MAMEEngine{Binary: "mame"}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
	if engine.Acceleration().Mode != AccelerationAuto {
		testingT.Fatalf("unexpected acceleration mode: %q", engine.Acceleration().Mode)
	}
	if len(engine.Platforms()) != 2 {
		testingT.Fatalf("unexpected platform count: %d", len(engine.Platforms()))
	}
}

func TestMame_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := MAMEEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestMame_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(mameBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}
	bundle.Manifest.Runtime.Profile = ""

	_, err = MAMEEngine{Binary: "mame"}.PlanLaunch(bundle)
	if err == nil {
		testingT.Fatal("PlanLaunch expected an error for a missing profile")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("PlanLaunch returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/profile-required" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}

func TestMame_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(mameBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := MAMEEngine{Binary: "mame"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "mame" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) != 5 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "-inipath" || plan.Arguments[1] != "." {
		testingT.Fatalf("unexpected config arguments: %v", plan.Arguments)
	}
	if plan.Arguments[2] != "-rompath" || plan.Arguments[3] != "rom" {
		testingT.Fatalf("unexpected ROM path arguments: %v", plan.Arguments)
	}
	if plan.Arguments[4] != "puckman" {
		testingT.Fatalf("unexpected driver argument: %v", plan.Arguments)
	}
	if !sandboxPathAllowed("emulator.yaml", plan.ReadPaths) {
		testingT.Fatalf("launch plan read paths do not include runtime config: %v", plan.ReadPaths)
	}
}

func mameBundleFS(testingT *testing.T) fstest.MapFS {
	testingT.Helper()

	return adapterBundleFS(testingT, adapterBundleFixture{
		Name:         "puckman-sample",
		Title:        "Puckman Sample",
		Platform:     "arcade",
		Engine:       "mame",
		Profile:      "puckman",
		ArtefactPath: "rom/puckman.zip",
		ArtefactData: []byte("mame-romset"),
	})
}
