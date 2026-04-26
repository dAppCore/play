package play

import "testing"

func TestShield_Verify_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle := shieldBundle(testingT, []byte("shield-rom"))
	registry := NewRegistry()
	if err := registry.Register(SyntheticEngine{}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	report := (Shield{Registry: registry}).Verify(bundle)
	if !report.OverallOK {
		testingT.Fatalf("Shield Verify returned issues: %v", report.Issues())
	}
	if !report.SBOM.Valid || !report.Code.OK || !report.Content.OK || !report.Threat.OK {
		testingT.Fatalf("Shield surfaces were not all OK: %+v", report)
	}
}

func TestShield_Verify_Bad(testingT *testing.T) {
	testingT.Parallel()

	bundle := shieldBundle(testingT, []byte("shield-rom"))
	tamperedFiles := renderedBundleFS(shieldRendered(testingT, []byte("shield-rom")), "rom/rom.bin", []byte("shield-rom"))
	tamperedFiles["rom/rom.bin"].Data = []byte("tampered")
	bundle.files = tamperedFiles
	registry := NewRegistry()
	if err := registry.Register(SyntheticEngine{}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	report := (Shield{Registry: registry}).Verify(bundle)
	if report.OverallOK {
		testingT.Fatal("Shield Verify expected a content failure")
	}
	if !hasIssueCode(report.Content.Issues, "hash/mismatch") {
		testingT.Fatalf("Shield Verify missing hash/mismatch issue: %v", report.Content.Issues)
	}
}

func TestShield_Verify_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle := shieldBundle(testingT, []byte("shield-rom"))
	bundle.Manifest.Verification.Engine.SHA256 = hashHex([]byte("wrong-engine"))
	registry := NewRegistry()
	if err := registry.Register(SyntheticEngine{}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	report := (Shield{Registry: registry}).Verify(bundle)
	if report.Code.OK {
		testingT.Fatal("Shield Verify expected a code-integrity failure")
	}
	if !hasIssueCode(report.Code.Issues, "code/hash-mismatch") {
		testingT.Fatalf("Shield Verify missing code/hash-mismatch issue: %v", report.Code.Issues)
	}
}

func shieldBundle(testingT *testing.T, artefactData []byte) Bundle {
	testingT.Helper()

	rendered := shieldRendered(testingT, artefactData)
	bundle, err := LoadBundle(renderedBundleFS(rendered, "rom/rom.bin", artefactData), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	return bundle
}

func shieldRendered(testingT *testing.T, artefactData []byte) RenderedBundle {
	testingT.Helper()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(BundleRequest{
		Name:           "shield-test",
		Title:          "Shield Test",
		Platform:       "synthetic",
		Licence:        "freeware",
		Engine:         "synthetic",
		ArtefactPath:   "rom/rom.bin",
		ArtefactData:   artefactData,
		ArtefactSHA256: hashHex(artefactData),
		ArtefactSize:   int64(len(artefactData)),
	})
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}

	return rendered
}
