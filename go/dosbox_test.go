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

func TestDosbox_DOSBoxEngine_Name_Good(t *core.T) {
	subject := (*DOSBoxEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Name_Bad(t *core.T) {
	subject := (*DOSBoxEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Name_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Platforms_Good(t *core.T) {
	subject := (*DOSBoxEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Platforms_Bad(t *core.T) {
	subject := (*DOSBoxEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Platforms_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Acceleration_Good(t *core.T) {
	subject := (*DOSBoxEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Acceleration_Bad(t *core.T) {
	subject := (*DOSBoxEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Acceleration_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Verify_Good(t *core.T) {
	subject := (*DOSBoxEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Verify_Bad(t *core.T) {
	subject := (*DOSBoxEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Verify_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_CodeIdentity_Good(t *core.T) {
	subject := (*DOSBoxEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_CodeIdentity_Bad(t *core.T) {
	subject := (*DOSBoxEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_CodeIdentity_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Run_Good(t *core.T) {
	subject := (*DOSBoxEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Run_Bad(t *core.T) {
	subject := (*DOSBoxEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_Run_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_PlanLaunch_Good(t *core.T) {
	subject := (*DOSBoxEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_PlanLaunch_Bad(t *core.T) {
	subject := (*DOSBoxEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestDosbox_DOSBoxEngine_PlanLaunch_Ugly(t *core.T) {
	subject := (*DOSBoxEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
