package play

import "testing"

func TestDeterministic_BundleArchive_Good(testingT *testing.T) {
	testingT.Parallel()

	firstHash := deterministicArchiveHash(testingT)
	secondHash := deterministicArchiveHash(testingT)
	if firstHash != secondHash {
		testingT.Fatalf("bundle archive hashes differ: %s != %s", firstHash, secondHash)
	}
}

func TestDeterministic_BundleArchive_Bad(testingT *testing.T) {
	testingT.Parallel()

	rendered := RenderedBundle{}
	_, err := rendered.Archive()
	if err == nil {
		testingT.Fatal("Archive expected an error for an empty rendered bundle")
	}
}

func TestDeterministic_BundleArchive_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	first, err := service.RenderBundle(deterministicBundleRequest("det-test", []byte("rom-one")))
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}
	second, err := service.RenderBundle(deterministicBundleRequest("det-test", []byte("rom-two")))
	if err != nil {
		testingT.Fatalf("RenderBundle returned second error: %v", err)
	}

	firstArchive, err := first.Archive()
	if err != nil {
		testingT.Fatalf("Archive returned error: %v", err)
	}
	secondArchive, err := second.Archive()
	if err != nil {
		testingT.Fatalf("Archive returned second error: %v", err)
	}
	if hashHex(firstArchive) == hashHex(secondArchive) {
		testingT.Fatal("different bundle inputs unexpectedly produced the same archive hash")
	}
}

func deterministicArchiveHash(testingT *testing.T) string {
	testingT.Helper()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(deterministicBundleRequest("det-test", []byte("deterministic-rom")))
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}

	archiveData, err := rendered.Archive()
	if err != nil {
		testingT.Fatalf("Archive returned error: %v", err)
	}

	return hashHex(archiveData)
}

func deterministicBundleRequest(name string, romData []byte) BundleRequest {
	return BundleRequest{
		Name:           name,
		Title:          "Deterministic Test",
		Platform:       "synthetic",
		Licence:        "freeware",
		Engine:         "synthetic",
		ArtefactPath:   "rom/rom.bin",
		ArtefactData:   romData,
		ArtefactSHA256: hashHex(romData),
		ArtefactSize:   int64(len(romData)),
	}
}
