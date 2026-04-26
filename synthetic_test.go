package play

import (
	"bytes"
	"testing"
)

func TestSynthetic_Run_Good(testingT *testing.T) {
	testingT.Parallel()

	var output bytes.Buffer
	err := SyntheticEngine{}.Run("rom.bin", EngineConfig{Output: &output})
	if err != nil {
		testingT.Fatalf("Run returned error: %v", err)
	}
	if !containsLine(output.String(), "SYNTHETIC ENGINE OK") {
		testingT.Fatalf("Run output missing success marker: %q", output.String())
	}
}

func TestSynthetic_Run_Bad(testingT *testing.T) {
	testingT.Parallel()

	err := SyntheticEngine{}.Run("rom.bin", EngineConfig{})
	if err != nil {
		testingT.Fatalf("Run returned error without output: %v", err)
	}
}

func TestSynthetic_Run_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	_, err = SyntheticEngine{}.PlanLaunch(bundle)
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
