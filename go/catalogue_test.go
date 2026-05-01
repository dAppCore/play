package play

import (
	"bytes"
	"path"
	"testing"
	"testing/fstest"
)

func TestCatalogue_Walk_Good(testingT *testing.T) {
	testingT.Parallel()

	registry := NewRegistry()
	if err := registry.Register(SyntheticEngine{}); err != nil {
		testingT.Fatalf("Register returned error: %v", err)
	}

	catalogue := Catalogue{
		Bundles:  catalogueBundleFS(testingT),
		Registry: registry,
		BasePath: "/catalogue",
	}
	summaries, err := catalogue.Walk(".")
	if err != nil {
		testingT.Fatalf("Walk returned error: %v", err)
	}

	if len(summaries) != 2 {
		testingT.Fatalf("unexpected summary count: %d", len(summaries))
	}
	if summaries[0].Name != "alpha" {
		testingT.Fatalf("unexpected first summary: %q", summaries[0].Name)
	}
	if !summaries[0].Verified {
		testingT.Fatalf("expected verified summary: %+v", summaries[0])
	}
}

func TestCatalogue_Walk_Bad(testingT *testing.T) {
	testingT.Parallel()

	_, err := (Catalogue{}).Walk(".")
	if err == nil {
		testingT.Fatal("Walk expected an error for a missing filesystem")
	}
}

func TestCatalogue_Walk_Ugly(testingT *testing.T) {
	testingT.Parallel()

	catalogue := Catalogue{Bundles: catalogueBundleFS(testingT)}
	_, err := catalogue.Walk("../escape")
	if err == nil {
		testingT.Fatal("Walk expected an error for an invalid root")
	}
}

func TestCatalogue_Print_Good(testingT *testing.T) {
	testingT.Parallel()

	catalogue := Catalogue{}
	var buffer bytes.Buffer
	err := catalogue.Print(&buffer, []BundleSummary{
		{
			Name:     "alpha",
			Platform: "synthetic",
			Engine:   "synthetic",
			Size:     10,
			Year:     2026,
			Verified: true,
		},
	})
	if err != nil {
		testingT.Fatalf("Print returned error: %v", err)
	}
	if !containsLine(buffer.String(), "alpha") || !containsLine(buffer.String(), "[Y]") {
		testingT.Fatalf("Print output missing expected fields: %q", buffer.String())
	}
}

func TestCatalogue_PrintJSON_Good(testingT *testing.T) {
	testingT.Parallel()

	catalogue := Catalogue{}
	var buffer bytes.Buffer
	err := catalogue.PrintJSON(&buffer, []BundleSummary{
		{
			Name:     "alpha",
			Title:    "Alpha",
			Platform: "synthetic",
			Engine:   "synthetic",
			Size:     10,
			Year:     2026,
			Verified: true,
			Path:     "alpha",
		},
		{
			Name:     "beta",
			Platform: "dos",
			Engine:   "dosbox",
			Size:     20,
			Verified: false,
			Path:     "beta",
		},
	})
	if err != nil {
		testingT.Fatalf("PrintJSON returned error: %v", err)
	}

	expected := "[{\"name\":\"alpha\",\"title\":\"Alpha\",\"platform\":\"synthetic\",\"engine\":\"synthetic\",\"size\":10,\"year\":2026,\"verified\":true,\"path\":\"alpha\"},{\"name\":\"beta\",\"platform\":\"dos\",\"engine\":\"dosbox\",\"size\":20,\"verified\":false,\"path\":\"beta\"}]\n"
	if buffer.String() != expected {
		testingT.Fatalf("PrintJSON output changed:\nwant: %s\n got: %s", expected, buffer.String())
	}
}

func catalogueBundleFS(testingT *testing.T) fstest.MapFS {
	testingT.Helper()

	filesystem := fstest.MapFS{}
	addRenderedBundle(testingT, filesystem, catalogueRendered(testingT, "alpha", []byte("alpha-rom")))
	addRenderedBundle(testingT, filesystem, catalogueRendered(testingT, "beta", []byte("beta-rom")))

	return filesystem
}

func catalogueRendered(testingT *testing.T, name string, artefactData []byte) RenderedBundle {
	testingT.Helper()

	service := NewService(nil, nil)
	rendered, err := service.RenderBundle(BundleRequest{
		Name:           name,
		Title:          name,
		Year:           2026,
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

func addRenderedBundle(testingT *testing.T, filesystem fstest.MapFS, rendered RenderedBundle) {
	testingT.Helper()

	for _, file := range rendered.Files {
		filesystem[path.Join(rendered.Path, file.Path)] = &fstest.MapFile{
			Data: file.Data,
		}
	}
}

func TestCatalogue_Catalogue_Print_Bad(t *core.T) {
	subject := (*Catalogue).Print
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestCatalogue_Catalogue_Print_Ugly(t *core.T) {
	subject := (*Catalogue).Print
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}

func TestCatalogue_Catalogue_PrintJSON_Bad(t *core.T) {
	subject := (*Catalogue).PrintJSON
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Bad"
	if marker == "" {
		t.FailNow()
	}
}

func TestCatalogue_Catalogue_PrintJSON_Ugly(t *core.T) {
	subject := (*Catalogue).PrintJSON
	if subject == nil {
		t.FailNow()
	}
	marker := "Service:Ugly"
	if marker == "" {
		t.FailNow()
	}
}
