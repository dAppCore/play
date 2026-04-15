package play

import "testing"

func TestDosbox_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := DOSBoxEngine{Binary: "dosbox"}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
	if engine.Acceleration().Mode != AccelerationAuto {
		testingT.Fatalf("unexpected acceleration mode: %q", engine.Acceleration().Mode)
	}
}

func TestDosbox_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := DOSBoxEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestDosbox_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	_, err = DOSBoxEngine{Binary: "dosbox"}.PlanLaunch(bundle)
	if err == nil {
		testingT.Fatal("PlanLaunch expected an error for a runtime mismatch")
	}

	engineError, ok := err.(EngineError)
	if !ok {
		testingT.Fatalf("PlanLaunch returned %T, want EngineError", err)
	}
	if engineError.Kind != "engine/runtime-mismatch" {
		testingT.Fatalf("unexpected engine error kind: %q", engineError.Kind)
	}
}
