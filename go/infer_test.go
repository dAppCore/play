package play

import "testing"

func TestInfer_PlanBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "mega-lo-mania",
		Title:          "Mega lo Mania",
		Platform:       "sega-genesis",
		Licence:        "freeware",
		ArtefactPath:   "rom/MegaLoMania.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Engine != "retroarch" {
		testingT.Fatalf("unexpected inferred engine: %q", plan.Manifest.Runtime.Engine)
	}
	if plan.Manifest.Runtime.Profile != "genesis" {
		testingT.Fatalf("unexpected inferred profile: %q", plan.Manifest.Runtime.Profile)
	}
	if plan.Manifest.Runtime.Acceleration != AccelerationAuto {
		testingT.Fatalf("unexpected inferred acceleration: %q", plan.Manifest.Runtime.Acceleration)
	}
	if plan.Manifest.Runtime.Filter != FrameFilterNearest {
		testingT.Fatalf("unexpected inferred filter: %q", plan.Manifest.Runtime.Filter)
	}
}

func TestInfer_PlanBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	_, issues := service.PlanBundle(BundleRequest{
		Name:           "beneath-a-steel-sky",
		Title:          "Beneath a Steel Sky",
		Platform:       "scummvm",
		Licence:        "freeware",
		ArtefactPath:   "game/BASS/sky.dsk",
		ArtefactSHA256: validArtefactSHA256,
	})
	if !hasIssueCode(issues, "manifest/runtime-profile-required") {
		testingT.Fatalf("PlanBundle missing manifest/runtime-profile-required issue: %v", issues)
	}
}

func TestInfer_PlanBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "custom-genesis",
		Title:          "Custom Genesis",
		Platform:       "sega-genesis",
		Licence:        "freeware",
		Engine:         "retroarch",
		Profile:        "custom-core",
		Acceleration:   AccelerationRequired,
		Filter:         FrameFilterCRT,
		ArtefactPath:   "rom/game.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Engine != "retroarch" {
		testingT.Fatalf("unexpected runtime engine: %q", plan.Manifest.Runtime.Engine)
	}
	if plan.Manifest.Runtime.Profile != "custom-core" {
		testingT.Fatalf("explicit runtime profile was not preserved: %q", plan.Manifest.Runtime.Profile)
	}
	if plan.Manifest.Runtime.Acceleration != AccelerationRequired {
		testingT.Fatalf("explicit acceleration was not preserved: %q", plan.Manifest.Runtime.Acceleration)
	}
	if plan.Manifest.Runtime.Filter != FrameFilterCRT {
		testingT.Fatalf("explicit filter was not preserved: %q", plan.Manifest.Runtime.Filter)
	}
}

func TestInfer_DOSDefaults_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "command-and-conquer",
		Title:          "Command & Conquer",
		Platform:       "dos",
		Licence:        "freeware",
		ArtefactPath:   "rom/CNC.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Engine != "dosbox" {
		testingT.Fatalf("unexpected inferred engine: %q", plan.Manifest.Runtime.Engine)
	}
	if plan.Manifest.Runtime.Profile != "dos" {
		testingT.Fatalf("unexpected inferred profile: %q", plan.Manifest.Runtime.Profile)
	}
	if plan.Manifest.Runtime.Filter != FrameFilterNearest {
		testingT.Fatalf("unexpected inferred filter: %q", plan.Manifest.Runtime.Filter)
	}
}

func TestInfer_DOSBoxXDefaults_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "pc98-sample",
		Title:          "PC-98 Sample",
		Platform:       "pc-98",
		Licence:        "freeware",
		ArtefactPath:   "rom/pc98.hdi",
		ArtefactSHA256: validArtefactSHA256,
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Engine != "dosbox-x" {
		testingT.Fatalf("unexpected inferred engine: %q", plan.Manifest.Runtime.Engine)
	}
	if plan.Manifest.Runtime.Profile != "pc-98" {
		testingT.Fatalf("unexpected inferred profile: %q", plan.Manifest.Runtime.Profile)
	}
	if plan.Manifest.Runtime.Filter != FrameFilterNearest {
		testingT.Fatalf("unexpected inferred filter: %q", plan.Manifest.Runtime.Filter)
	}
}

func TestInfer_AdditionalAdapterDefaults_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "c64-sample",
		Title:          "C64 Sample",
		Platform:       "commodore-64",
		Licence:        "freeware",
		ArtefactPath:   "rom/game.d64",
		ArtefactSHA256: validArtefactSHA256,
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Engine != "vice" {
		testingT.Fatalf("unexpected inferred engine: %q", plan.Manifest.Runtime.Engine)
	}
	if plan.Manifest.Runtime.Profile != "c64" {
		testingT.Fatalf("unexpected inferred profile: %q", plan.Manifest.Runtime.Profile)
	}
	if plan.Manifest.Runtime.Filter != FrameFilterNearest {
		testingT.Fatalf("unexpected inferred filter: %q", plan.Manifest.Runtime.Filter)
	}
}

func TestInfer_MAMEProfile_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	_, issues := service.PlanBundle(BundleRequest{
		Name:           "arcade-sample",
		Title:          "Arcade Sample",
		Platform:       "arcade",
		Licence:        "freeware",
		ArtefactPath:   "rom/puckman.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if !hasIssueCode(issues, "manifest/runtime-profile-required") {
		testingT.Fatalf("PlanBundle missing manifest/runtime-profile-required issue: %v", issues)
	}
}

func TestInfer_FUSEDefaults_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(BundleRequest{
		Name:           "spectrum-sample",
		Title:          "Spectrum Sample",
		Platform:       "zx-spectrum-128k",
		Licence:        "freeware",
		ArtefactPath:   "rom/game.tap",
		ArtefactSHA256: validArtefactSHA256,
	})
	if issues.HasIssues() {
		testingT.Fatalf("PlanBundle returned issues: %v", issues)
	}

	if plan.Manifest.Runtime.Engine != "fuse" {
		testingT.Fatalf("unexpected inferred engine: %q", plan.Manifest.Runtime.Engine)
	}
	if plan.Manifest.Runtime.Profile != "128k" {
		testingT.Fatalf("unexpected inferred profile: %q", plan.Manifest.Runtime.Profile)
	}
}
