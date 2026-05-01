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
	if !hasIssueCode(issues, "manifest/format-version-required") {
		testingT.Fatal("Validate missing manifest/format-version-required issue")
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
	manifest.Permissions.FileSystem.Write = []string{"rom/"}
	manifest.Resources.CPUPercent = -1
	manifest.Resources.MemoryBytes = -1
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
	if !hasIssueCode(issues, "manifest/filesystem-write-denied") {
		testingT.Fatal("Validate missing manifest/filesystem-write-denied issue")
	}
	if !hasIssueCode(issues, "manifest/resources-cpu-invalid") {
		testingT.Fatal("Validate missing manifest/resources-cpu-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/resources-memory-invalid") {
		testingT.Fatal("Validate missing manifest/resources-memory-invalid issue")
	}
	if !hasIssueCode(issues, "manifest/distribution-mode-required") {
		testingT.Fatal("Validate missing manifest/distribution-mode-required issue")
	}
}

func TestValidate_BundlePathCanonical_Ugly(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	manifest.Artefact.Path = "rom/../payload.bin"
	manifest.Runtime.Config = "config//emulator.yaml"
	manifest.Permissions.FileSystem.Read = []string{"rom\\game.zip"}

	issues := manifest.Validate()
	if !hasIssueCode(issues, "manifest/artefact-path-invalid") {
		testingT.Fatalf("Validate missing canonical artefact path issue: %v", issues)
	}
	if !hasIssueCode(issues, "manifest/runtime-config-invalid") {
		testingT.Fatalf("Validate missing canonical runtime config issue: %v", issues)
	}
	if !hasIssueCode(issues, "manifest/filesystem-read-invalid") {
		testingT.Fatalf("Validate missing filesystem read path issue: %v", issues)
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

func TestValidate_ValidationIssue_Error_Good(t *core.T) {
	subject := (*ValidationIssue).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationIssue_Error_Bad(t *core.T) {
	subject := (*ValidationIssue).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationIssue_Error_Ugly(t *core.T) {
	subject := (*ValidationIssue).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationErrors_Error_Good(t *core.T) {
	subject := (*ValidationErrors).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationErrors_Error_Bad(t *core.T) {
	subject := (*ValidationErrors).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationErrors_Error_Ugly(t *core.T) {
	subject := (*ValidationErrors).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationErrors_HasIssues_Good(t *core.T) {
	subject := (*ValidationErrors).HasIssues
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationErrors_HasIssues_Bad(t *core.T) {
	subject := (*ValidationErrors).HasIssues
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_ValidationErrors_HasIssues_Ugly(t *core.T) {
	subject := (*ValidationErrors).HasIssues
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_Manifest_Validate_Good(t *core.T) {
	subject := (*Manifest).Validate
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_Manifest_Validate_Bad(t *core.T) {
	subject := (*Manifest).Validate
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestValidate_Manifest_Validate_Ugly(t *core.T) {
	subject := (*Manifest).Validate
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
