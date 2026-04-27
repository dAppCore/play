package play

import (
	"testing"
	"testing/fstest"
)

func TestSnes9x_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := Snes9xEngine{Binary: "snes9x"}
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

func TestSnes9x_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := Snes9xEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestSnes9x_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(snes9xBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}
	bundle.Manifest.Platform = "sega-genesis"

	_, err = Snes9xEngine{Binary: "snes9x"}.PlanLaunch(bundle)
	if err == nil {
		testingT.Fatal("PlanLaunch expected an error for an unsupported platform")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("PlanLaunch returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/platform-unsupported" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}

func TestSnes9x_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(snes9xBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := Snes9xEngine{Binary: "snes9x"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "snes9x" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) != 3 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "-conf" || plan.Arguments[1] != "emulator.yaml" {
		testingT.Fatalf("unexpected config arguments: %v", plan.Arguments)
	}
	if plan.Arguments[2] != "rom/game.sfc" {
		testingT.Fatalf("unexpected artefact argument: %v", plan.Arguments)
	}
}

func snes9xBundleFS(testingT *testing.T) fstest.MapFS {
	testingT.Helper()

	return adapterBundleFS(testingT, adapterBundleFixture{
		Name:         "snes-sample",
		Title:        "SNES Sample",
		Platform:     "snes",
		Engine:       "snes9x",
		ArtefactPath: "rom/game.sfc",
		ArtefactData: []byte("snes-rom"),
	})
}
