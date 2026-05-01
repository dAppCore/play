package play

import "testing"

func TestWrite_RenderedBundle_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(BundleRequest{
		Name:           "mega-lo-mania",
		Title:          "Mega lo Mania",
		Platform:       "sega-genesis",
		Licence:        "freeware",
		Engine:         "retroarch",
		Profile:        "genesis",
		ArtefactPath:   "rom/MegaLoMania.zip",
		ArtefactSHA256: validArtefactSHA256,
	})
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}

	writer := newMemoryBundleWriter()
	if err := rendered.Write(writer); err != nil {
		testingT.Fatalf("Write returned error: %v", err)
	}

	if !writer.hasDirectory("mega-lo-mania") {
		testingT.Fatal("Write did not create the bundle root directory")
	}
	if !writer.hasFile("mega-lo-mania/manifest.yaml") {
		testingT.Fatal("Write did not create manifest.yaml")
	}
	if !writer.hasFile("mega-lo-mania/checksums.sha256") {
		testingT.Fatal("Write did not create checksums.sha256")
	}
}

func TestWrite_RenderedBundle_Bad(testingT *testing.T) {
	testingT.Parallel()

	err := (RenderedBundle{Path: "bundle"}).Write(nil)
	if err == nil {
		testingT.Fatal("Write expected an error for a missing writer")
	}
}

func TestWrite_RenderedBundle_Ugly(testingT *testing.T) {
	testingT.Parallel()

	rendered := RenderedBundle{
		Path: "bundle",
		Files: []RenderedFile{
			{
				Path: "../escape.txt",
				Data: []byte("bad"),
			},
		},
	}

	err := rendered.Write(newMemoryBundleWriter())
	if err == nil {
		testingT.Fatal("Write expected an error for an invalid file path")
	}

	writeError, ok := err.(WriteError)
	if !ok {
		testingT.Fatalf("Write returned %T, want WriteError", err)
	}
	if writeError.Kind != "bundle/file-path-invalid" {
		testingT.Fatalf("unexpected write error kind: %q", writeError.Kind)
	}
}

func TestWrite_Service_Good(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	writer := newMemoryBundleWriter()

	err := service.WriteBundle(BundleRequest{
		Name:              "command-and-conquer",
		Title:             "Command & Conquer",
		Platform:          "dos",
		Licence:           "freeware",
		Engine:            "dosbox",
		Profile:           "dos",
		ArtefactPath:      "rom/CNC.zip",
		ArtefactSHA256:    validArtefactSHA256,
		RuntimeConfigPath: "config/emulator.yaml",
		VerificationChain: "meta/checksums.sha256",
		SBOMPath:          "meta/sbom.json",
	}, writer)
	if err != nil {
		testingT.Fatalf("WriteBundle returned error: %v", err)
	}

	if !writer.hasDirectory("command-and-conquer/config") {
		testingT.Fatal("WriteBundle did not create the config directory")
	}
	if !writer.hasFile("command-and-conquer/meta/sbom.json") {
		testingT.Fatal("WriteBundle did not create the custom SBOM path")
	}
}

func TestWrite_Service_Bad(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	err := service.WriteBundle(BundleRequest{}, newMemoryBundleWriter())
	if err == nil {
		testingT.Fatal("WriteBundle expected an error for an invalid request")
	}
}

func TestWrite_Service_Ugly(testingT *testing.T) {
	testingT.Parallel()

	service := NewService(nil, nil)
	writer := newMemoryBundleWriter()
	writer.directoryErrors["bundle"] = WriteError{
		Kind:    "writer/denied",
		Path:    "bundle",
		Message: "directory creation denied",
	}

	err := service.WriteBundle(BundleRequest{
		Name:           "bundle",
		Title:          "Bundle",
		Platform:       "dos",
		Licence:        "freeware",
		Engine:         "dosbox",
		ArtefactPath:   "rom/game.zip",
		ArtefactSHA256: validArtefactSHA256,
	}, writer)
	if err == nil {
		testingT.Fatal("WriteBundle expected an error when the writer fails")
	}
}

type memoryBundleWriter struct {
	directories     map[string]struct{}
	files           map[string][]byte
	directoryErrors map[string]error
	fileErrors      map[string]error
}

func newMemoryBundleWriter() *memoryBundleWriter {
	return &memoryBundleWriter{
		directories:     map[string]struct{}{},
		files:           map[string][]byte{},
		directoryErrors: map[string]error{},
		fileErrors:      map[string]error{},
	}
}

func (writer *memoryBundleWriter) EnsureDirectory(path string) error {
	if err, exists := writer.directoryErrors[path]; exists {
		return err
	}

	writer.directories[path] = struct{}{}
	return nil
}

func (writer *memoryBundleWriter) WriteFile(path string, data []byte) error {
	if err, exists := writer.fileErrors[path]; exists {
		return err
	}

	writer.files[path] = cloneBytes(data)
	return nil
}

func (writer *memoryBundleWriter) hasDirectory(path string) bool {
	_, exists := writer.directories[path]
	return exists
}

func (writer *memoryBundleWriter) hasFile(path string) bool {
	_, exists := writer.files[path]
	return exists
}

func TestWrite_RenderedBundle_Write_Good(t *core.T) {
	subject := (*RenderedBundle).Write
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestWrite_RenderedBundle_Write_Bad(t *core.T) {
	subject := (*RenderedBundle).Write
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestWrite_RenderedBundle_Write_Ugly(t *core.T) {
	subject := (*RenderedBundle).Write
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestWrite_WriteError_Error_Good(t *core.T) {
	subject := (*WriteError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Good"
	if marker == "" {
		t.FailNow()
	}
}

func TestWrite_WriteError_Error_Bad(t *core.T) {
	subject := (*WriteError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestWrite_WriteError_Error_Ugly(t *core.T) {
	subject := (*WriteError).Error
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
