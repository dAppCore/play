package play

import "testing"

func TestScummvm_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := ScummVMEngine{Binary: "scummvm"}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
	if engine.Acceleration().Mode != AccelerationAuto {
		testingT.Fatalf("unexpected acceleration mode: %q", engine.Acceleration().Mode)
	}
}

func TestScummvm_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := ScummVMEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestScummvm_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(scummVMBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	bundle.Manifest.Runtime.Profile = ""
	_, err = ScummVMEngine{Binary: "scummvm"}.PlanLaunch(bundle)
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

func TestScummvm_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(scummVMBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := ScummVMEngine{Binary: "scummvm"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "scummvm" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) != 2 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "--path=game/BASS" {
		testingT.Fatalf("unexpected data path argument: %q", plan.Arguments[0])
	}
	if plan.Arguments[1] != "sky" {
		testingT.Fatalf("unexpected game id argument: %q", plan.Arguments[1])
	}
}
