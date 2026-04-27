package play

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestThreat_VerifyThreat_Good(testingT *testing.T) {
	testingT.Parallel()

	bundle := threatBundle(testingT, testZIP(testZIPEntry{
		Path: "game/data.bin",
		Data: []byte("safe"),
		Mode: 0644,
	}))

	result := verifyThreat(bundle)
	if !result.OK {
		testingT.Fatalf("verifyThreat returned findings: %v", result.Findings)
	}
}

func TestThreat_VerifyThreat_Bad(testingT *testing.T) {
	testingT.Parallel()

	bundle := threatBundle(testingT, testZIP(testZIPEntry{
		Path: "setup.sh",
		Data: []byte("echo bad"),
		Mode: 0755,
	}))

	result := verifyThreat(bundle)
	if result.OK {
		testingT.Fatal("verifyThreat expected an embedded script finding")
	}
	if !hasIssueCode(result.Issues, "threat/embedded-script") {
		testingT.Fatalf("verifyThreat missing embedded-script issue: %v", result.Issues)
	}
}

func TestThreat_VerifyThreat_Ugly(testingT *testing.T) {
	testingT.Parallel()

	expandingBundle := threatBundle(testingT, testZIP(testZIPEntry{
		Path: "game/pad.bin",
		Data: bytes.Repeat([]byte("0"), 1024*1024),
		Mode: 0644,
	}))

	expandingResult := verifyThreat(expandingBundle)
	if !hasIssueCode(expandingResult.Issues, "threat/zip-expansion") {
		testingT.Fatalf("verifyThreat missing zip-expansion issue: %v", expandingResult.Issues)
	}

	deepBundle := threatBundle(testingT, testZIP(testZIPEntry{
		Path: nestedArchivePath(maxZIPEntryDepth + 1),
		Data: []byte("deep"),
		Mode: 0644,
	}))

	deepResult := verifyThreat(deepBundle)
	if !hasIssueCode(deepResult.Issues, "threat/path-depth") {
		testingT.Fatalf("verifyThreat missing path-depth issue: %v", deepResult.Issues)
	}

	oversizedBundle := threatBundle(testingT, testRawZIP(testZIPEntry{
		Path:               "game/huge.bin",
		Mode:               0644,
		UncompressedSize64: maxZIPEntryUncompressedBytes + 1,
	}))

	oversizedResult := verifyThreat(oversizedBundle)
	if !hasIssueCode(oversizedResult.Issues, "threat/entry-size") {
		testingT.Fatalf("verifyThreat missing entry-size issue: %v", oversizedResult.Issues)
	}
}

func TestThreat_MaliciousZIPMatrix_Ugly(testingT *testing.T) {
	testingT.Parallel()

	cases := []struct {
		Name       string
		Artefact   []byte
		WantedCode string
	}{
		{
			Name: "traversal-path",
			Artefact: testZIP(testZIPEntry{
				Path: "../escape.bin",
				Data: []byte("escape"),
				Mode: 0644,
			}),
			WantedCode: "threat/path-invalid",
		},
		{
			Name: "absolute-path",
			Artefact: testZIP(testZIPEntry{
				Path: "/absolute.bin",
				Data: []byte("absolute"),
				Mode: 0644,
			}),
			WantedCode: "threat/path-invalid",
		},
		{
			Name: "normalised-path",
			Artefact: testZIP(testZIPEntry{
				Path: "game/../escape.bin",
				Data: []byte("escape"),
				Mode: 0644,
			}),
			WantedCode: "threat/path-invalid",
		},
		{
			Name: "windows-path",
			Artefact: testZIP(testZIPEntry{
				Path: "game\\payload.bin",
				Data: []byte("payload"),
				Mode: 0644,
			}),
			WantedCode: "threat/path-invalid",
		},
		{
			Name:       "aggregate-expansion",
			Artefact:   testAggregateZIP(),
			WantedCode: "threat/archive-size",
		},
	}

	for _, entry := range cases {
		entry := entry
		testingT.Run(entry.Name, func(testingT *testing.T) {
			testingT.Parallel()

			result := verifyThreat(threatBundle(testingT, entry.Artefact))
			if !hasIssueCode(result.Issues, entry.WantedCode) {
				testingT.Fatalf("verifyThreat missing %s issue: %v", entry.WantedCode, result.Issues)
			}
		})
	}
}

type testZIPEntry struct {
	Path               string
	Data               []byte
	Mode               fs.FileMode
	UncompressedSize64 uint64
	CompressedSize64   uint64
}

func testZIP(entry testZIPEntry) []byte {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	header := &zip.FileHeader{
		Name:   entry.Path,
		Method: zip.Deflate,
	}
	header.SetMode(entry.Mode)

	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		panic(err)
	}
	if _, err := fileWriter.Write(entry.Data); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func testRawZIP(entry testZIPEntry) []byte {
	return testRawZIPArchive(entry)
}

func testRawZIPArchive(entries ...testZIPEntry) []byte {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	for _, entry := range entries {
		header := &zip.FileHeader{
			Name:               entry.Path,
			Method:             zip.Store,
			CompressedSize64:   entry.CompressedSize64,
			UncompressedSize64: entry.UncompressedSize64,
		}
		header.SetMode(entry.Mode)

		fileWriter, err := writer.CreateRaw(header)
		if err != nil {
			panic(err)
		}
		if _, err := fileWriter.Write(entry.Data); err != nil {
			panic(err)
		}
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func testAggregateZIP() []byte {
	entries := make([]testZIPEntry, 0, 9)
	for index := 0; index < 9; index++ {
		entries = append(entries, testZIPEntry{
			Path:               "game/chunk.bin",
			Mode:               0644,
			CompressedSize64:   maxZIPEntryUncompressedBytes,
			UncompressedSize64: maxZIPEntryUncompressedBytes,
		})
	}

	return testRawZIPArchive(entries...)
}

func nestedArchivePath(depth int) string {
	var buffer bytes.Buffer
	for index := 0; index < depth; index++ {
		if index > 0 {
			buffer.WriteByte('/')
		}
		buffer.WriteString("level")
	}
	buffer.WriteString("/rom.bin")

	return buffer.String()
}

func threatBundle(testingT *testing.T, artefactData []byte) Bundle {
	testingT.Helper()

	manifestData := []byte(`name: threat-test
title: "Threat Test"
platform: synthetic
licence: freeware
artefact:
  path: rom/rom.zip
  sha256: "` + hashHex(artefactData) + `"
runtime:
  engine: synthetic
  config: emulator.yaml
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
`)
	bundle, err := LoadBundle(fstest.MapFS{
		"manifest.yaml": {Data: manifestData},
		"rom/rom.zip":   {Data: artefactData},
	}, ".")
	if err != nil {
		testingT.Fatalf("LoadBundle returned error: %v", err)
	}

	return bundle
}
