package play

import "testing"

func TestRetroarch_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := RetroArchEngine{Binary: "retroarch"}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
	if engine.Acceleration().Mode != AccelerationAuto {
		testingT.Fatalf("unexpected acceleration mode: %q", engine.Acceleration().Mode)
	}
}

func TestRetroarch_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := RetroArchEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestRetroarch_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	bundle.Manifest.Runtime.Profile = "unknown-core"
	_, err = RetroArchEngine{Binary: "retroarch"}.PlanLaunch(bundle)
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

func TestRetroarch_PlanLaunch_Genesis(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := RetroArchEngine{Binary: "retroarch"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "retroarch" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) < 3 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "-L" {
		testingT.Fatalf("unexpected first argument: %q", plan.Arguments[0])
	}
	if plan.Arguments[1] != "cores/genesis_plus_gx" {
		testingT.Fatalf("unexpected core path: %q", plan.Arguments[1])
	}
}
