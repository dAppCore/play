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

func verifiedBundleFS() fstest.MapFS {
	romData := []byte("rom")
	emulatorData := []byte("engine: retroarch\nprofile: genesis\n")
	sbomData := []byte("{\"bomFormat\":\"CycloneDX\"}")
	manifestData := []byte(validManifestYAMLWithArtefactHash(hashHex(romData)))

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
