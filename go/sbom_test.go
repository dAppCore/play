package play

import "testing"

func TestSbom_BuildSBOM_Good(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	data, err := BuildSBOM(manifest)
	if err != nil {
		testingT.Fatalf("BuildSBOM returned error: %v", err)
	}

	result := validateSBOM(data, manifest, "sbom.json")
	if !result.Valid {
		testingT.Fatalf("validateSBOM returned issues: %v", result.Issues)
	}
	if result.SerialNumber == "" {
		testingT.Fatal("validateSBOM expected a deterministic serial number")
	}
}

func TestSbom_BuildSBOM_Bad(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	result := validateSBOM([]byte("{"), manifest, "sbom.json")
	if !hasIssueCode(result.Issues, "sbom/parse-failed") {
		testingT.Fatalf("validateSBOM missing sbom/parse-failed issue: %v", result.Issues)
	}
}

func TestSbom_BuildSBOM_Ugly(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	result := validateSBOM([]byte(`{"bomFormat":"CycloneDX","specVersion":"1.5","serialNumber":"urn:uuid:broken","version":1,"metadata":{"timestamp":"1980-01-01T00:00:00Z","tools":[]},"components":[]}`), manifest, "sbom.json")
	if !hasIssueCode(result.Issues, "sbom/application-hash-missing") {
		testingT.Fatalf("validateSBOM missing application hash issue: %v", result.Issues)
	}
}
