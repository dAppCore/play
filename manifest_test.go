package play

import "testing"

const validArtefactSHA256 = "1a0806c20104d3461d8ede70362f16734dbd6a17db24005d1841a7387c9b2405"

func TestManifest_LoadManifest_Good(testingT *testing.T) {
	testingT.Parallel()

	manifest, err := LoadManifest([]byte(validManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned error: %v", err)
	}

	if manifest.Name != "mega-lo-mania" {
		testingT.Fatalf("unexpected manifest name: %q", manifest.Name)
	}
	if manifest.Runtime.Engine != "retroarch" {
		testingT.Fatalf("unexpected engine: %q", manifest.Runtime.Engine)
	}

	preserved, err := LoadManifest([]byte(validPreservationManifestYAML()))
	if err != nil {
		testingT.Fatalf("LoadManifest returned preservation error: %v", err)
	}
	if preserved.Verification.Chain != "checksums.sha256" {
		testingT.Fatalf("unexpected preservation chain: %q", preserved.Verification.Chain)
	}
	if preserved.Verification.SBOM != "sbom.json" {
		testingT.Fatalf("unexpected default SBOM path: %q", preserved.Verification.SBOM)
	}
}

func TestManifest_LoadManifest_Bad(testingT *testing.T) {
	testingT.Parallel()

	_, err := LoadManifest([]byte("name: mega-lo-mania\nunknown: value\n"))
	if err == nil {
		testingT.Fatal("LoadManifest expected an error for unknown fields")
	}
}

func TestManifest_LoadManifest_Ugly(testingT *testing.T) {
	testingT.Parallel()

	_, err := LoadManifest([]byte(validManifestYAML() + "\n---\nname: second\n"))
	if err == nil {
		testingT.Fatal("LoadManifest expected an error for multiple YAML documents")
	}

	parseError, ok := err.(ParseError)
	if !ok {
		testingT.Fatalf("LoadManifest returned %T, want ParseError", err)
	}
	if parseError.Kind != "manifest/multiple-documents" {
		testingT.Fatalf("unexpected parse error kind: %q", parseError.Kind)
	}
}

func FuzzManifest_LoadManifest(fuzzT *testing.F) {
	fuzzT.Add(validManifestYAML())
	fuzzT.Add(validPreservationManifestYAML())
	fuzzT.Add("name: fuzz\nunknown: value\n")
	fuzzT.Add("---\nname: first\n---\nname: second\n")

	fuzzT.Fuzz(func(testingT *testing.T, data string) {
		manifest, err := LoadManifest([]byte(data))
		if err != nil {
			return
		}

		_ = manifest.Validate()
	})
}

func validManifestYAML() string {
	return validManifestYAMLWithArtefactHash(validArtefactSHA256)
}

func validManifestYAMLWithArtefactHash(artefactSHA256 string) string {
	return `name: mega-lo-mania
title: "Mega lo Mania"
author: "Sensible Software"
year: 1991
platform: sega-genesis
genre: strategy
licence: freeware
artefact:
  path: rom/MegaLoMania.zip
  sha256: "` + artefactSHA256 + `"
  size: 554192
  media_type: application/zip
  source: "Rights-cleared redistribution"
runtime:
  engine: retroarch
  profile: genesis
  config: emulator.yaml
  entrypoint: rom/MegaLoMania.zip
  acceleration: auto
  filter: nearest
verification:
  chain: checksums.sha256
  sbom: sbom.json
  deterministic: true
permissions:
  network: false
  microphone: false
  filesystem:
    read:
      - rom/
    write:
      - saves/
      - screenshots/
save:
  path: saves/
  screenshots: screenshots/
distribution:
  mode: catalogue
  byorom: false
`
}

func validPreservationManifestYAML() string {
	return `name: sample-bundle
title: "Synthetic Sample Bundle"
platform: synthetic
licence: freeware
artefact:
  path: rom/rom.bin
  sha256: "` + validArtefactSHA256 + `"
runtime:
  engine: synthetic
  config: emulator.yaml
preservation:
  verified: true
  chain: checksums.sha256
`
}
