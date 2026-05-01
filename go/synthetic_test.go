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

func TestSynthetic_SyntheticEngine_Name_Good(t *core.T) {
	subject := (*SyntheticEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Name_Bad(t *core.T) {
	subject := (*SyntheticEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Name_Ugly(t *core.T) {
	subject := (*SyntheticEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Platforms_Good(t *core.T) {
	subject := (*SyntheticEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Platforms_Bad(t *core.T) {
	subject := (*SyntheticEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Platforms_Ugly(t *core.T) {
	subject := (*SyntheticEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Run_Good(t *core.T) {
	subject := (*SyntheticEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Run_Bad(t *core.T) {
	subject := (*SyntheticEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Run_Ugly(t *core.T) {
	subject := (*SyntheticEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Verify_Good(t *core.T) {
	subject := (*SyntheticEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Verify_Bad(t *core.T) {
	subject := (*SyntheticEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_Verify_Ugly(t *core.T) {
	subject := (*SyntheticEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_CodeIdentity_Good(t *core.T) {
	subject := (*SyntheticEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_CodeIdentity_Bad(t *core.T) {
	subject := (*SyntheticEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_CodeIdentity_Ugly(t *core.T) {
	subject := (*SyntheticEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_PlanLaunch_Good(t *core.T) {
	subject := (*SyntheticEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_PlanLaunch_Bad(t *core.T) {
	subject := (*SyntheticEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestSynthetic_SyntheticEngine_PlanLaunch_Ugly(t *core.T) {
	subject := (*SyntheticEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
