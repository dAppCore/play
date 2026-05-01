package play

import (
	"testing"
	"testing/fstest"
)

func TestFuse_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := FUSEEngine{Binary: "fuse"}
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

func TestFuse_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := FUSEEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestFuse_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(fuseBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}
	bundle.Manifest.Runtime.Profile = "plus3"

	_, err = FUSEEngine{Binary: "fuse"}.PlanLaunch(bundle)
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

func TestFuse_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(fuseBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := FUSEEngine{Binary: "fuse"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "fuse" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) != 5 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "--settings" || plan.Arguments[1] != "emulator.yaml" {
		testingT.Fatalf("unexpected settings arguments: %v", plan.Arguments)
	}
	if plan.Arguments[2] != "--machine" || plan.Arguments[3] != "128" {
		testingT.Fatalf("unexpected machine arguments: %v", plan.Arguments)
	}
	if plan.Arguments[4] != "rom/game.tap" {
		testingT.Fatalf("unexpected artefact argument: %v", plan.Arguments)
	}
}

func fuseBundleFS(testingT *testing.T) fstest.MapFS {
	testingT.Helper()

	return adapterBundleFS(testingT, adapterBundleFixture{
		Name:         "spectrum-sample",
		Title:        "Spectrum Sample",
		Platform:     "zx-spectrum-128k",
		Engine:       "fuse",
		Profile:      "128k",
		ArtefactPath: "rom/game.tap",
		ArtefactData: []byte("spectrum-tape"),
	})
}
