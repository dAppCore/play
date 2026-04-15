package play

import "testing"

func TestValidate_Manifest_Good(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	issues := manifest.Validate()
	if issues.HasIssues() {
		testingT.Fatalf("Validate returned issues: %v", issues)
	}
}

func TestValidate_Manifest_Bad(testingT *testing.T) {
	testingT.Parallel()

	manifest := Manifest{}
	issues := manifest.Validate()

	if !issues.HasIssues() {
		testingT.Fatal("Validate expected issues for an empty manifest")
	}
	if !hasIssueCode(issues, "manifest/name-required") {
		testingT.Fatal("Validate missing manifest/name-required issue")
	}
	if !hasIssueCode(issues, "manifest/runtime-engine-required") {
		testingT.Fatal("Validate missing manifest/runtime-engine-required issue")
	}
}

func TestValidate_Manifest_Ugly(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	manifest.Artefact.Path = "../rom/MegaLoMania.zip"
	manifest.Artefact.SHA256 = "short"
	manifest.Runtime.Config = "/absolute/emulator.yaml"
	manifest.Runtime.Acceleration = AccelerationMode("warp")
	manifest.Runtime.Filter = FrameFilter("phosphor-plus")
	manifest.Distribution.Mode = ""
	manifest.Distribution.BYOROM = true

	issues := manifest.Validate()
	if !hasIssueCode(issues, "manifest/artefact-path-invalid") {
		testingT.Fatal("Validate missing manifest/artefact-path-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/runtime-config-invalid") {
		testingT.Fatal("Validate missing manifest/runtime-config-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/runtime-acceleration-invalid") {
		testingT.Fatal("Validate missing manifest/runtime-acceleration-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/runtime-filter-invalid") {
		testingT.Fatal("Validate missing manifest/runtime-filter-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/artefact-sha256-invalid") {
		testingT.Fatal("Validate missing manifest/artefact-sha256-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/distribution-mode-required") {
		testingT.Fatal("Validate missing manifest/distribution-mode-required issue")
	}
}

func hasIssueCode(issues ValidationErrors, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}

	return false
}
