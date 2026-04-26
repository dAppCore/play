package play

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestArchive_RenderedBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	rendered := renderedArchiveBundle(testingT)
	first, err := rendered.Archive()
	if err != nil {
		testingT.Fatalf("Archive returned error: %v", err)
	}
	second, err := rendered.Archive()
	if err != nil {
		testingT.Fatalf("Archive returned error on repeat: %v", err)
	}
	if !bytes.Equal(first, second) {
		testingT.Fatal("Archive output is not deterministic")
	}

	reader, err := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if err != nil {
		testingT.Fatalf("zip.NewReader returned error: %v", err)
	}
	if len(reader.File) != 5 {
		testingT.Fatalf("unexpected archive file count: %d", len(reader.File))
	}
	if reader.File[0].Name != "sample/checksums.sha256" {
		testingT.Fatalf("unexpected first archive entry: %q", reader.File[0].Name)
	}
}

func TestArchive_RenderedBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	_, err := (RenderedBundle{}).Archive()
	if err == nil {
		testingT.Fatal("Archive expected an error for missing bundle path")
	}
}

func TestArchive_RenderedBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	_, err := (RenderedBundle{
		Path: "sample",
		Files: []RenderedFile{
			{
				Path: "../escape",
				Data: []byte("bad"),
			},
		},
	}).Archive()
	if err == nil {
		testingT.Fatal("Archive expected an error for an invalid file path")
	}
}

func renderedArchiveBundle(testingT *testing.T) RenderedBundle {
	testingT.Helper()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(BundleRequest{
		Name:           "sample",
		Title:          "Sample",
		Platform:       "synthetic",
		Licence:        "freeware",
		Engine:         "synthetic",
		ArtefactPath:   "rom/rom.bin",
		ArtefactData:   []byte("rom"),
		ArtefactSHA256: hashHex([]byte("rom")),
	})
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}

	return rendered
}
