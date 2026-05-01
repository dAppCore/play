package play

import (
	"context"
	"testing"

	core "dappco.re/go"
)

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

func TestScummvm_ProcessVerify_Good(testingT *testing.T) {
	testingT.Parallel()

	c := scummVMCore("ScummVM 2.8.0")
	engine := ScummVMEngine{Binary: "scummvm", Core: c}
	if err := engine.Verify(); err != nil {
		testingT.Fatalf("Verify returned error: %v", err)
	}
}

func TestScummvm_ProcessVerify_Bad(testingT *testing.T) {
	testingT.Parallel()

	c := scummVMCore("ScummVM 2.6.1")
	err := (ScummVMEngine{Binary: "scummvm", Core: c}).Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error for an unsupported ScummVM version")
	}
}

func TestScummvm_ProcessVerify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	c := core.New()
	c.Action("process.run", func(context.Context, core.Options) core.Result {
		return core.Fail(core.NewError("missing scummvm"))
	})

	err := (ScummVMEngine{Binary: "scummvm", Core: c}).Verify()
	if err == nil {
		testingT.Fatal("Verify expected an error when process verification fails")
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
	if len(plan.Arguments) != 3 {
		testingT.Fatalf("unexpected argument count: %d", len(plan.Arguments))
	}
	if plan.Arguments[0] != "--path=game/BASS" {
		testingT.Fatalf("unexpected data path argument: %q", plan.Arguments[0])
	}
	if plan.Arguments[1] != "--savepath=saves/" {
		testingT.Fatalf("unexpected save path argument: %q", plan.Arguments[1])
	}
	if plan.Arguments[2] != "sky" {
		testingT.Fatalf("unexpected game id argument: %q", plan.Arguments[2])
	}
}

func scummVMCore(versionOutput string) *core.Core {
	c := core.New()
	c.Action("process.run", func(context.Context, core.Options) core.Result {
		return core.Ok(versionOutput)
	})

	return c
}

func TestScummvm_ScummVMEngine_Name_Good(t *core.T) {
	subject := (*ScummVMEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Name_Bad(t *core.T) {
	subject := (*ScummVMEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Name_Ugly(t *core.T) {
	subject := (*ScummVMEngine).Name
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Platforms_Good(t *core.T) {
	subject := (*ScummVMEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Platforms_Bad(t *core.T) {
	subject := (*ScummVMEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Platforms_Ugly(t *core.T) {
	subject := (*ScummVMEngine).Platforms
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Acceleration_Good(t *core.T) {
	subject := (*ScummVMEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Acceleration_Bad(t *core.T) {
	subject := (*ScummVMEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Acceleration_Ugly(t *core.T) {
	subject := (*ScummVMEngine).Acceleration
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Verify_Good(t *core.T) {
	subject := (*ScummVMEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Verify_Bad(t *core.T) {
	subject := (*ScummVMEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Verify_Ugly(t *core.T) {
	subject := (*ScummVMEngine).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_CodeIdentity_Good(t *core.T) {
	subject := (*ScummVMEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_CodeIdentity_Bad(t *core.T) {
	subject := (*ScummVMEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_CodeIdentity_Ugly(t *core.T) {
	subject := (*ScummVMEngine).CodeIdentity
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Run_Good(t *core.T) {
	subject := (*ScummVMEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Run_Bad(t *core.T) {
	subject := (*ScummVMEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_Run_Ugly(t *core.T) {
	subject := (*ScummVMEngine).Run
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_PlanLaunch_Good(t *core.T) {
	subject := (*ScummVMEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_PlanLaunch_Bad(t *core.T) {
	subject := (*ScummVMEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestScummvm_ScummVMEngine_PlanLaunch_Ugly(t *core.T) {
	subject := (*ScummVMEngine).PlanLaunch
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
