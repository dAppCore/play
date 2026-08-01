package main

import (
	"context"
	"os"
	"path"

	core "dappco.re/go"
	"dappco.re/go/play"
)

// renderBundleToDisk materialises a verified RetroArch bundle under root and
// returns the bundle's directory name. The rendered tree carries a manifest,
// SBOM and checksum chain, so verify and shield-verify succeed once a matching
// engine is registered.
//
//	root := testingT.TempDir()
//	name := renderBundleToDisk(testingT, root)
//	_ = runInfo(invocation{Root: root, Bundle: name}, core.NewBuffer())
func renderBundleToDisk(testingT *core.T, root string) string {
	testingT.Helper()

	artefact := []byte("mega-lo-mania-rom")
	service := play.NewService(nil, nil)
	rendered, err := service.RenderBundle(play.BundleRequest{
		Name:           "mega-lo-mania",
		Title:          "Mega lo Mania",
		Author:         "Sensible Software",
		Year:           1991,
		Platform:       "sega-genesis",
		Genre:          "strategy",
		Licence:        "freeware",
		Engine:         "retroarch",
		Profile:        "genesis",
		ArtefactPath:   "rom/MegaLoMania.zip",
		ArtefactData:   artefact,
		ArtefactSHA256: hashBytes(artefact),
		ArtefactSize:   int64(len(artefact)),
		ArtefactSource: "Rights-cleared redistribution",
		ResourceLimits: play.ResourceLimits{CPUPercent: 75, MemoryBytes: 268435456},
	})
	if err != nil {
		testingT.Fatalf("RenderBundle returned error: %v", err)
	}
	if err := rendered.Write(localBundleWriter{Root: root}); err != nil {
		testingT.Fatalf("Write returned error: %v", err)
	}

	return rendered.Path
}

func registerRetroArch(testingT *core.T) {
	testingT.Helper()
	if _, found := play.ResolveEngine("retroarch"); found {
		return
	}
	if err := play.RegisterEngine(play.RetroArchEngine{Binary: "retroarch"}); err != nil {
		testingT.Fatalf("RegisterEngine returned error: %v", err)
	}
}

func TestCommand_RunInfo_Good(testingT *core.T) {
	root := testingT.TempDir()
	name := renderBundleToDisk(testingT, root)

	out := core.NewBuffer()
	err := runInfo(invocation{Root: root, Bundle: name}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "mega-lo-mania")
}

func TestCommand_RunInfo_Bad(testingT *core.T) {
	out := core.NewBuffer()
	err := runInfo(invocation{Root: testingT.TempDir(), Bundle: "missing"}, out)
	core.AssertError(testingT, err)
}

func TestCommand_RunInfo_Ugly(testingT *core.T) {
	root := testingT.TempDir()
	name := renderBundleToDisk(testingT, root)
	// Empty manifest filename still resolves to the bundle directory; reading a
	// directory as a file is the ugly path that must surface an error.
	out := core.NewBuffer()
	err := printBundleFile(root, name, "", out)
	core.AssertError(testingT, err)
}

func TestCommand_RunList_Good(testingT *core.T) {
	root := testingT.TempDir()
	renderBundleToDisk(testingT, root)

	out := core.NewBuffer()
	err := runList(invocation{Root: root}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "mega-lo-mania")
}

func TestCommand_RunList_Bad(testingT *core.T) {
	out := core.NewBuffer()
	err := runList(invocation{Root: path.Join(testingT.TempDir(), "missing")}, out)
	core.AssertError(testingT, err)
}

func TestCommand_RunList_Ugly(testingT *core.T) {
	root := testingT.TempDir()
	renderBundleToDisk(testingT, root)

	out := core.NewBuffer()
	err := runList(invocation{Root: root, JSON: true}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "mega-lo-mania")
}

func TestCommand_RunVerify_Good(testingT *core.T) {
	registerRetroArch(testingT)
	root := testingT.TempDir()
	name := renderBundleToDisk(testingT, root)

	out := core.NewBuffer()
	err := runVerify(invocation{Operation: operationVerify, Root: root, Bundle: name}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "ShieldReport")
	core.AssertContains(testingT, out.String(), "OverallOK: true")
}

func TestCommand_RunVerify_Bad(testingT *core.T) {
	out := core.NewBuffer()
	err := runVerify(invocation{Operation: operationVerify, Root: path.Join(testingT.TempDir(), "missing"), Bundle: "."}, out)
	core.AssertError(testingT, err)
}

func TestCommand_RunVerify_Ugly(testingT *core.T) {
	// A rendered bundle with no engine registered for an unknown engine name
	// still verifies its SBOM, code, content and threat surfaces. Drive the
	// printShieldReport unverified branch directly with a failing report.
	out := core.NewBuffer()
	printShieldReport(out, play.ShieldReport{OverallOK: false})
	core.AssertContains(testingT, out.String(), "OverallOK: false")
}

func TestCommand_RunEngines_Good(testingT *core.T) {
	registerRetroArch(testingT)

	out := core.NewBuffer()
	err := runEngines(out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "ENGINE")
	core.AssertContains(testingT, out.String(), "retroarch")
}

func TestCommand_RunBundle_Good(testingT *core.T) {
	source := testingT.TempDir()
	romPath := path.Join(source, "MegaLoMania.zip")
	if err := os.WriteFile(romPath, []byte("rom-bytes"), 0o644); err != nil {
		testingT.Fatalf("WriteFile returned error: %v", err)
	}

	out := core.NewBuffer()
	target := testingT.TempDir()
	err := runBundle(invocation{
		Root:     target,
		Name:     "mega-lo-mania",
		Title:    "Mega lo Mania",
		Platform: "sega-genesis",
		Licence:  "freeware",
		Engine:   "retroarch",
		Profile:  "genesis",
		ROM:      romPath,
	}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "bundle created")
	core.AssertTrue(testingT, core.Stat(path.Join(target, "mega-lo-mania", "manifest.yaml")).OK)
}

func TestCommand_RunBundle_Bad(testingT *core.T) {
	out := core.NewBuffer()
	err := runBundle(invocation{
		Name: "mega-lo-mania",
		ROM:  path.Join(testingT.TempDir(), "missing.zip"),
	}, out)
	core.AssertError(testingT, err)
}

func TestCommand_RunBundle_Ugly(testingT *core.T) {
	source := testingT.TempDir()
	romPath := path.Join(source, "MegaLoMania.zip")
	if err := os.WriteFile(romPath, []byte("rom-bytes"), 0o644); err != nil {
		testingT.Fatalf("WriteFile returned error: %v", err)
	}

	out := core.NewBuffer()
	target := testingT.TempDir()
	err := runBundle(invocation{
		Root:     target,
		Name:     "mega-lo-mania",
		Title:    "Mega lo Mania",
		Platform: "sega-genesis",
		Licence:  "freeware",
		Engine:   "retroarch",
		Profile:  "genesis",
		ROM:      romPath,
		Archive:  true,
	}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "bundle archive created")
	core.AssertTrue(testingT, core.Stat(path.Join(target, "mega-lo-mania.zip")).OK)
}

func TestCommand_Run_Good(testingT *core.T) {
	registerRetroArch(testingT)
	root := testingT.TempDir()
	renderBundleToDisk(testingT, root)

	c := core.New(core.WithOption("name", "core-play"))
	out := core.NewBuffer()
	err := run(context.Background(), c, []string{"list", "--root", root}, out)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, out.String(), "mega-lo-mania")
}

func TestCommand_Run_Bad(testingT *core.T) {
	c := core.New(core.WithOption("name", "core-play"))
	out := core.NewBuffer()
	err := run(context.Background(), c, []string{"list", "extra-arg"}, out)
	core.AssertError(testingT, err)
}

func TestCommand_Run_Ugly(testingT *core.T) {
	c := core.New(core.WithOption("name", "core-play"))
	out := core.NewBuffer()
	err := run(context.Background(), c, []string{"engines"}, out)
	core.AssertNoError(testingT, err)
}
