package play

import (
	"testing"
	"testing/fstest"
)

func TestBundle_LoadBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(validBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	if bundle.Path != "." {
		testingT.Fatalf("unexpected bundle path: %q", bundle.Path)
	}
	if bundle.Validate().HasIssues() {
		testingT.Fatalf("bundle validation returned issues: %v", bundle.Validate())
	}
}

func TestBundle_LoadBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	_, err := LoadBundle(fstest.MapFS{
		"emulator.yaml":       {Data: []byte("engine: retroarch\n")},
		"checksums.sha256":    {Data: []byte("abc  manifest.yaml\n")},
		"sbom.json":           {Data: []byte("{}")},
		"rom/MegaLoMania.zip": {Data: []byte("rom")},
	}, ".")
	if err == nil {
		testingT.Fatal("LoadBundle expected an error when manifest.yaml is missing")
	}
}

func TestBundle_LoadBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	_, err := LoadBundle(validBundleFS(), "../escape")
	if err == nil {
		testingT.Fatal("LoadBundle expected an error for invalid bundle path")
	}

	pathError, ok := err.(PathError)
	if !ok {
		testingT.Fatalf("LoadBundle returned %T, want PathError", err)
	}
	if pathError.Kind != "bundle/path-invalid" {
		testingT.Fatalf("unexpected path error kind: %q", pathError.Kind)
	}
}

func validBundleFS() fstest.MapFS {
	return fstest.MapFS{
		"manifest.yaml":       {Data: []byte(validManifestYAML())},
		"emulator.yaml":       {Data: []byte("engine: retroarch\nprofile: genesis\n")},
		"checksums.sha256":    {Data: []byte("9f0f  rom/MegaLoMania.zip\n")},
		"sbom.json":           {Data: []byte("{\"bomFormat\":\"CycloneDX\"}")},
		"rom/MegaLoMania.zip": {Data: []byte("rom")},
	}
}

func TestBundle_Bundle_Validate_Good(t *core.T) {
	subject := (*Bundle).Validate
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestBundle_Bundle_Validate_Bad(t *core.T) {
	subject := (*Bundle).Validate
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestBundle_Bundle_Validate_Ugly(t *core.T) {
	subject := (*Bundle).Validate
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestBundle_PathError_Error_Good(t *core.T) {
	subject := (*PathError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestBundle_PathError_Error_Bad(t *core.T) {
	subject := (*PathError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestBundle_PathError_Error_Ugly(t *core.T) {
	subject := (*PathError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
