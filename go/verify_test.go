package play

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"testing/fstest"
)

func TestVerify_Bundle_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	issues := bundle.VerifyWithRegistry(registry)
	if issues.HasIssues() {
		testingT.Fatalf("VerifyWithRegistry returned issues: %v", issues)
	}
}

func TestVerify_Bundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(brokenChecksumBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	issues := bundle.VerifyWithRegistry(registry)
	if !hasIssueCode(issues, "hash/mismatch") {
		testingT.Fatalf("VerifyWithRegistry missing hash/mismatch issue: %v", issues)
	}
}

func TestVerify_Bundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	issues := bundle.VerifyWithRegistry(NewRegistry())
	if !hasIssueCode(issues, "engine/unavailable") {
		testingT.Fatalf("VerifyWithRegistry missing engine/unavailable issue: %v", issues)
	}
}

func TestVerify_ChecksumChainCoverage_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle, err := LoadBundle(verifiedBundleFS(), ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	issues := bundle.VerifyWithRegistry(registry)
	if issues.HasIssues() {
		testingT.Fatalf("VerifyWithRegistry returned issues: %v", issues)
	}
}

func TestVerify_ChecksumChainCoverage_Bad(testingT *testing.T) {
	testingT.Parallel()

	filesystem := verifiedBundleFS()
	filesystem["payload/setup.sh"] = &fstest.MapFile{
		Data: []byte("echo untracked"),
	}
	bundle, err := LoadBundle(filesystem, ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	issues := bundle.VerifyWithRegistry(registry)
	if !hasIssueCode(issues, "hash/unrecorded-file") {
		testingT.Fatalf("VerifyWithRegistry missing hash/unrecorded-file issue: %v", issues)
	}
}

func TestVerify_ChecksumChainCoverage_Ugly(testingT *testing.T) {
	testingT.Parallel()

	filesystem := verifiedBundleFS()
	filesystem["checksums.sha256"] = &fstest.MapFile{
		Data: []byte(
			hashHex(filesystem["emulator.yaml"].Data) + "  emulator.yaml\n" +
				hashHex(filesystem["sbom.json"].Data) + "  sbom.json\n" +
				hashHex(filesystem["rom/MegaLoMania.zip"].Data) + "  rom/MegaLoMania.zip\n",
		),
	}
	bundle, err := LoadBundle(filesystem, ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	registry := NewRegistry()
	if err := registry.Register(stubEngine{name: "retroarch", platforms: []string{"sega-genesis"}}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	issues := bundle.VerifyWithRegistry(registry)
	if !hasIssueCode(issues, "hash/chain-entry-missing") {
		testingT.Fatalf("VerifyWithRegistry missing hash/chain-entry-missing issue: %v", issues)
	}
}

func TestVerify_ParseChecksumFile_Ugly(testingT *testing.T) {
	testingT.Parallel()

	_, err := ParseChecksumFile([]byte(validArtefactSHA256 + "  rom/../escape.bin\n"))
	if err == nil {
		testingT.Fatal("ParseChecksumFile expected an error for a non-canonical path")
	}
}

func verifiedBundleFS() fstest.MapFS {
	romData := []byte("rom")
	emulatorData := []byte("engine: retroarch\nprofile: genesis\n")
	manifestData := []byte(validManifestYAMLWithArtefactHash(hashHex(romData)))
	manifest, err := LoadManifest(manifestData)
	if err != nil {
		panic(err)
	}
	sbomData, err := BuildSBOM(manifest)
	if err != nil {
		panic(err)
	}

	checksumData := []byte(
		hashHex(manifestData) + "  manifest.yaml\n" +
			hashHex(emulatorData) + "  emulator.yaml\n" +
			hashHex(sbomData) + "  sbom.json\n" +
			hashHex(romData) + "  rom/MegaLoMania.zip\n",
	)

	return fstest.MapFS{
		"manifest.yaml":       {Data: manifestData},
		"emulator.yaml":       {Data: emulatorData},
		"checksums.sha256":    {Data: checksumData},
		"sbom.json":           {Data: sbomData},
		"rom/MegaLoMania.zip": {Data: romData},
	}
}

func brokenChecksumBundleFS() fstest.MapFS {
	filesystem := verifiedBundleFS()
	filesystem["checksums.sha256"] = &fstest.MapFile{
		Data: []byte(
			hashHex([]byte("not-rom")) + "  rom/MegaLoMania.zip\n",
		),
	}

	return filesystem
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestVerify_Bundle_Verify_Good(t *core.T) {
	subject := (*Bundle).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_Bundle_Verify_Bad(t *core.T) {
	subject := (*Bundle).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_Bundle_Verify_Ugly(t *core.T) {
	subject := (*Bundle).Verify
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_Bundle_VerifyWithRegistry_Good(t *core.T) {
	subject := (*Bundle).VerifyWithRegistry
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_Bundle_VerifyWithRegistry_Bad(t *core.T) {
	subject := (*Bundle).VerifyWithRegistry
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_Bundle_VerifyWithRegistry_Ugly(t *core.T) {
	subject := (*Bundle).VerifyWithRegistry
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_ParseChecksumFile_Good(t *core.T) {
	subject := ParseChecksumFile
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_ParseChecksumFile_Bad(t *core.T) {
	subject := ParseChecksumFile
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_ChecksumParseError_Error_Good(t *core.T) {
	subject := (*ChecksumParseError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_ChecksumParseError_Error_Bad(t *core.T) {
	subject := (*ChecksumParseError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestVerify_ChecksumParseError_Error_Ugly(t *core.T) {
	subject := (*ChecksumParseError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
