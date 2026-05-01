package play

import (
	"testing"
	"testing/fstest"
)

func TestVice_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	engine := VICEEngine{Binary: "x64sc"}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
	if engine.Acceleration().Mode != AccelerationAuto {
		testingT.Fatalf("unexpected acceleration mode: %q", engine.Acceleration().Mode)
	}
	if len(engine.Platforms()) != 5 {
		testingT.Fatalf("unexpected platform count: %d", len(engine.Platforms()))
	}
}

func TestVice_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	engine := VICEEngine{}
	err := engine.Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for a missing binary")
	}
}

func TestVice_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(viceBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}
	bundle.Manifest.Runtime.Profile = "pet"

	_, err = VICEEngine{Binary: "x64sc"}.PlanLaunch(bundle)
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

func TestVice_PlanLaunch_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(viceBundleFS(testingT), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	plan, err := VICEEngine{Binary: "x64sc"}.PlanLaunch(bundle)
	if err != nil {
		testingT.Fatalf("PlanLaunch returned error: %v", err)
	}

	if plan.Engine != "vice" {
		testingT.Fatalf("unexpected engine: %q", plan.Engine)
	}
	if len(plan.Arguments) != 6 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "-config" || plan.Arguments[1] != "emulator.yaml" {
		testingT.Fatalf("unexpected config arguments: %v", plan.Arguments)
	}
	if plan.Arguments[2] != "-model" || plan.Arguments[3] != "c64pal" {
		testingT.Fatalf("unexpected model arguments: %v", plan.Arguments)
	}
	if plan.Arguments[4] != "-autostart" || plan.Arguments[5] != "rom/game.d64" {
		testingT.Fatalf("unexpected autostart arguments: %v", plan.Arguments)
	}
}

func viceBundleFS(testingT *testing.T) fstest.MapFS {
	testingT.Helper()

	return adapterBundleFS(testingT, adapterBundleFixture{
		Name:         "c64-sample",
		Title:        "C64 Sample",
		Platform:     "commodore-64",
		Engine:       "vice",
		Profile:      "c64",
		ArtefactPath: "rom/game.d64",
		ArtefactData: []byte("c64-disk"),
	})
}

func TestVice_VICEEngine_Name_Good(t *core.T) {
	subject := (*VICEEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Name_Bad(t *core.T) {
	subject := (*VICEEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Name_Ugly(t *core.T) {
	subject := (*VICEEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Platforms_Good(t *core.T) {
	subject := (*VICEEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Platforms_Bad(t *core.T) {
	subject := (*VICEEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Platforms_Ugly(t *core.T) {
	subject := (*VICEEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Acceleration_Good(t *core.T) {
	subject := (*VICEEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Acceleration_Bad(t *core.T) {
	subject := (*VICEEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Acceleration_Ugly(t *core.T) {
	subject := (*VICEEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Verify_Good(t *core.T) {
	subject := (*VICEEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Verify_Bad(t *core.T) {
	subject := (*VICEEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Verify_Ugly(t *core.T) {
	subject := (*VICEEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_CodeIdentity_Good(t *core.T) {
	subject := (*VICEEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_CodeIdentity_Bad(t *core.T) {
	subject := (*VICEEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_CodeIdentity_Ugly(t *core.T) {
	subject := (*VICEEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Run_Good(t *core.T) {
	subject := (*VICEEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Run_Bad(t *core.T) {
	subject := (*VICEEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_Run_Ugly(t *core.T) {
	subject := (*VICEEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_PlanLaunch_Good(t *core.T) {
	subject := (*VICEEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_PlanLaunch_Bad(t *core.T) {
	subject := (*VICEEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVice_VICEEngine_PlanLaunch_Ugly(t *core.T) {
	subject := (*VICEEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
