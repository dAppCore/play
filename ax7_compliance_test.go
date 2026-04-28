package play

import core "dappco.re/go"

type stubengine struct {
	name      string
	platforms []string
	verifyErr error
}

func (engine stubengine) Name() string {
	return engine.name
}

func (engine stubengine) Platforms() []string {
	return clonePaths(engine.platforms)
}

func (engine stubengine) Run(string, EngineConfig) error {
	return engine.verifyErr
}

func (engine stubengine) Verify() error {
	return engine.verifyErr
}

type failingwriter struct {
	failEnsure bool
	failWrite  bool
}

func (writer failingwriter) EnsureDirectory(string) error {
	if writer.failEnsure {
		return core.NewError("ensure failed")
	}

	return nil
}

func (writer failingwriter) WriteFile(string, []byte) error {
	if writer.failWrite {
		return core.NewError("write failed")
	}

	return nil
}

func ax7ProcessCore(output string) *core.Core {
	c := core.New()
	c.Action("process.run", func(core.Context, core.Options) core.Result {
		return core.Ok(output)
	})

	return c
}

func ax7EngineConfig(profile string) EngineConfig {
	return EngineConfig{
		Core:    ax7ProcessCore("process ok"),
		Context: core.Background(),
		Profile: profile,
		Output:  core.NewBuffer(),
	}
}

func ax7Bundle(testingT *core.T, filesystem core.FS) Bundle {
	testingT.Helper()

	bundle, err := LoadBundle(filesystem, ".")
	core.RequireNoError(testingT, err)

	return bundle
}

func ax7Registry(testingT *core.T, engine Engine) *Registry {
	testingT.Helper()

	registry := NewRegistry()
	core.RequireNoError(testingT, registry.Register(engine))

	return registry
}

func ax7BundleRequest(name string) BundleRequest {
	return BundleRequest{
		Name:           name,
		Title:          "AX7 Bundle",
		Platform:       "synthetic",
		Licence:        "freeware",
		Engine:         "synthetic",
		ArtefactPath:   "rom/game.bin",
		ArtefactData:   []byte("rom"),
		ArtefactSHA256: hashHex([]byte("rom")),
		ArtefactSize:   3,
	}
}

func TestAX7_RenderedBundle_Archive_Good(testingT *core.T) {
	rendered := renderedArchiveBundle(testingT)
	data, err := rendered.Archive()
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, data)
}

func TestAX7_RenderedBundle_Archive_Bad(testingT *core.T) {
	rendered := RenderedBundle{Path: "", Files: []RenderedFile{{Path: "manifest.yaml", Data: []byte("x")}}}
	data, err := rendered.Archive()
	core.AssertError(testingT, err, "bundle/path-required")
	core.AssertNil(testingT, data)
}

func TestAX7_RenderedBundle_Archive_Ugly(testingT *core.T) {
	rendered := RenderedBundle{Path: "game", Files: []RenderedFile{{Path: "../escape", Data: []byte("x")}}}
	data, err := rendered.Archive()
	core.AssertError(testingT, err, "bundle/file-path-invalid")
	core.AssertNil(testingT, data)
}

func TestAX7_Bundle_Validate_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, validBundleFS())
	issues := bundle.Validate()
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_Bundle_Validate_Bad(testingT *core.T) {
	bundle := Bundle{}
	issues := bundle.Validate()
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_Bundle_Validate_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, validBundleFS())
	bundle.Manifest.Artefact.Path = "../escape"
	issues := bundle.Validate()
	core.AssertTrue(testingT, hasIssueCode(issues, "manifest/artefact-path-invalid"))
}

func TestAX7_PathError_Error_Good(testingT *core.T) {
	err := PathError{Kind: "bundle/missing", Path: "games/demo", Message: "not found"}
	got := err.Error()
	core.AssertContains(testingT, got, "games/demo")
}

func TestAX7_PathError_Error_Bad(testingT *core.T) {
	err := PathError{Kind: "bundle/missing", Message: "not found"}
	got := err.Error()
	core.AssertEqual(testingT, "bundle/missing: : not found", got)
}

func TestAX7_PathError_Error_Ugly(testingT *core.T) {
	err := PathError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_Catalogue_Walk_Good(testingT *core.T) {
	catalogue := Catalogue{Bundles: catalogueBundleFS(testingT), Registry: ax7Registry(testingT, SyntheticEngine{})}
	summaries, err := catalogue.Walk(".")
	core.AssertNoError(testingT, err)
	core.AssertLen(testingT, summaries, 2)
}

func TestAX7_Catalogue_Walk_Bad(testingT *core.T) {
	catalogue := Catalogue{}
	summaries, err := catalogue.Walk(".")
	core.AssertError(testingT, err, "bundle/filesystem-missing")
	core.AssertNil(testingT, summaries)
}

func TestAX7_Catalogue_Walk_Ugly(testingT *core.T) {
	catalogue := Catalogue{Bundles: catalogueBundleFS(testingT), BasePath: "library", Registry: ax7Registry(testingT, SyntheticEngine{})}
	summaries, err := catalogue.Walk(".")
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, summaries[0].Path, "library")
}

func TestAX7_Catalogue_Print_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (Catalogue{}).Print(output, []BundleSummary{{Name: "demo", Title: "Demo", Platform: "synthetic", Engine: "synthetic"}})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "demo")
}

func TestAX7_Catalogue_Print_Bad(testingT *core.T) {
	output := core.NewBuffer()
	err := (Catalogue{}).Print(output, nil)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "NAME")
}

func TestAX7_Catalogue_Print_Ugly(testingT *core.T) {
	output := core.NewBuffer()
	err := (Catalogue{}).Print(output, []BundleSummary{{Name: "long-name", Size: 42, Verified: true}})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "[Y]")
}

func TestAX7_Catalogue_PrintJSON_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (Catalogue{}).PrintJSON(output, []BundleSummary{{Name: "demo"}})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "demo")
}

func TestAX7_Catalogue_PrintJSON_Bad(testingT *core.T) {
	output := core.NewBuffer()
	err := (Catalogue{}).PrintJSON(output, nil)
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "null")
}

func TestAX7_Catalogue_PrintJSON_Ugly(testingT *core.T) {
	output := core.NewBuffer()
	err := (Catalogue{}).PrintJSON(output, []BundleSummary{{Name: "demo", Size: 1}})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "\"size\":1")
}

func TestAX7_EngineError_Error_Good(testingT *core.T) {
	err := EngineError{Kind: "engine/binary-required", Name: "retroarch", Message: "binary path is required"}
	got := err.Error()
	core.AssertContains(testingT, got, "retroarch")
}

func TestAX7_EngineError_Error_Bad(testingT *core.T) {
	err := EngineError{Kind: "engine/nil", Message: "engine is required"}
	got := err.Error()
	core.AssertEqual(testingT, "engine/nil: engine is required", got)
}

func TestAX7_EngineError_Error_Ugly(testingT *core.T) {
	err := EngineError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_NewRegistry_Good(testingT *core.T) {
	registry := NewRegistry()
	names := registry.Names()
	core.AssertEmpty(testingT, names)
}

func TestAX7_NewRegistry_Bad(testingT *core.T) {
	registry := NewRegistry()
	_, found := registry.Resolve("missing")
	core.AssertFalse(testingT, found)
}

func TestAX7_NewRegistry_Ugly(testingT *core.T) {
	first := NewRegistry()
	second := NewRegistry()
	core.AssertNotEqual(testingT, core.Sprintf("%p", first), core.Sprintf("%p", second))
}

func TestAX7_Registry_Register_Good(testingT *core.T) {
	registry := NewRegistry()
	err := registry.Register(stubengine{name: "ax7-registry-good", platforms: []string{"synthetic"}})
	core.AssertNoError(testingT, err)
}

func TestAX7_Registry_Register_Bad(testingT *core.T) {
	registry := NewRegistry()
	err := registry.Register(nil)
	core.AssertError(testingT, err, "engine/nil")
}

func TestAX7_Registry_Register_Ugly(testingT *core.T) {
	registry := NewRegistry()
	core.RequireNoError(testingT, registry.Register(stubengine{name: "ax7-duplicate", platforms: []string{"synthetic"}}))
	err := registry.Register(stubengine{name: "ax7-duplicate", platforms: []string{"synthetic"}})
	core.AssertError(testingT, err, "engine/duplicate")
}

func TestAX7_Registry_Resolve_Good(testingT *core.T) {
	registry := ax7Registry(testingT, stubengine{name: "ax7-resolve", platforms: []string{"synthetic"}})
	engine, found := registry.Resolve("ax7-resolve")
	core.AssertTrue(testingT, found)
	core.AssertEqual(testingT, "ax7-resolve", engine.Name())
}

func TestAX7_Registry_Resolve_Bad(testingT *core.T) {
	registry := NewRegistry()
	engine, found := registry.Resolve("missing")
	core.AssertFalse(testingT, found)
	core.AssertNil(testingT, engine)
}

func TestAX7_Registry_Resolve_Ugly(testingT *core.T) {
	var registry *Registry
	engine, found := registry.Resolve("missing")
	core.AssertFalse(testingT, found)
	core.AssertNil(testingT, engine)
}

func TestAX7_Registry_Names_Good(testingT *core.T) {
	registry := NewRegistry()
	core.RequireNoError(testingT, registry.Register(stubengine{name: "b", platforms: []string{"synthetic"}}))
	core.RequireNoError(testingT, registry.Register(stubengine{name: "a", platforms: []string{"synthetic"}}))
	core.AssertEqual(testingT, []string{"a", "b"}, registry.Names())
}

func TestAX7_Registry_Names_Bad(testingT *core.T) {
	registry := NewRegistry()
	names := registry.Names()
	core.AssertEmpty(testingT, names)
}

func TestAX7_Registry_Names_Ugly(testingT *core.T) {
	var registry *Registry
	names := registry.Names()
	core.AssertNil(testingT, names)
}

func TestAX7_RegisterEngine_Good(testingT *core.T) {
	err := RegisterEngine(stubengine{name: "ax7-register-engine-good", platforms: []string{"synthetic"}})
	core.AssertNoError(testingT, err)
	_, found := ResolveEngine("ax7-register-engine-good")
	core.AssertTrue(testingT, found)
}

func TestAX7_RegisterEngine_Bad(testingT *core.T) {
	err := RegisterEngine(nil)
	core.AssertError(testingT, err, "engine/nil")
	core.AssertNotNil(testingT, err)
}

func TestAX7_RegisterEngine_Ugly(testingT *core.T) {
	err := RegisterEngine(stubengine{name: "", platforms: []string{"synthetic"}})
	core.AssertError(testingT, err, "engine/name-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_ResolveEngine_Good(testingT *core.T) {
	core.RequireNoError(testingT, RegisterEngine(stubengine{name: "ax7-resolve-engine-good", platforms: []string{"synthetic"}}))
	engine, found := ResolveEngine("ax7-resolve-engine-good")
	core.AssertTrue(testingT, found)
	core.AssertEqual(testingT, "ax7-resolve-engine-good", engine.Name())
}

func TestAX7_ResolveEngine_Bad(testingT *core.T) {
	engine, found := ResolveEngine("ax7-resolve-engine-missing")
	core.AssertFalse(testingT, found)
	core.AssertNil(testingT, engine)
}

func TestAX7_ResolveEngine_Ugly(testingT *core.T) {
	engine, found := ResolveEngine("")
	core.AssertFalse(testingT, found)
	core.AssertNil(testingT, engine)
}

func TestAX7_RegisteredEngines_Good(testingT *core.T) {
	core.RequireNoError(testingT, RegisterEngine(stubengine{name: "ax7-registered-engines-good", platforms: []string{"synthetic"}}))
	names := RegisteredEngines()
	core.AssertContains(testingT, names, "ax7-registered-engines-good")
}

func TestAX7_RegisteredEngines_Bad(testingT *core.T) {
	names := RegisteredEngines()
	core.AssertNotContains(testingT, names, "ax7-registered-engines-missing")
	core.AssertNotNil(testingT, names)
}

func TestAX7_RegisteredEngines_Ugly(testingT *core.T) {
	first := RegisteredEngines()
	second := RegisteredEngines()
	core.AssertEqual(testingT, first, second)
}

func TestAX7_FrameBuffer_Clone_Good(testingT *core.T) {
	frame := validRGBAFrame()
	clone := frame.Clone()
	core.AssertEqual(testingT, frame.Width, clone.Width)
	core.AssertEqual(testingT, len(frame.Data), len(clone.Data))
}

func TestAX7_FrameBuffer_Clone_Bad(testingT *core.T) {
	frame := FrameBuffer{}
	clone := frame.Clone()
	core.AssertNil(testingT, clone.Data)
}

func TestAX7_FrameBuffer_Clone_Ugly(testingT *core.T) {
	frame := validRGBAFrame()
	clone := frame.Clone()
	frame.Data[0] = 9
	core.AssertNotEqual(testingT, frame.Data[0], clone.Data[0])
}

func TestAX7_FrameBuffer_Validate_Good(testingT *core.T) {
	frame := validRGBAFrame()
	issues := frame.Validate()
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_FrameBuffer_Validate_Bad(testingT *core.T) {
	frame := FrameBuffer{}
	issues := frame.Validate()
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_FrameBuffer_Validate_Ugly(testingT *core.T) {
	frame := validRGBAFrame()
	frame.Format = PixelFormat("bad")
	issues := frame.Validate()
	core.AssertTrue(testingT, hasIssueCode(issues, "frame/format-invalid"))
}

func TestAX7_ResourceLimits_IsZero_Good(testingT *core.T) {
	limits := ResourceLimits{}
	got := limits.IsZero()
	core.AssertTrue(testingT, got)
}

func TestAX7_ResourceLimits_IsZero_Bad(testingT *core.T) {
	limits := ResourceLimits{CPUPercent: 1}
	got := limits.IsZero()
	core.AssertFalse(testingT, got)
}

func TestAX7_ResourceLimits_IsZero_Ugly(testingT *core.T) {
	limits := ResourceLimits{MemoryBytes: -1}
	got := limits.IsZero()
	core.AssertFalse(testingT, got)
}

func TestAX7_ManifestMigration_Migrated_Good(testingT *core.T) {
	migration := ManifestMigration{Applied: []string{"manifest/legacy-format-version"}}
	got := migration.Migrated()
	core.AssertTrue(testingT, got)
}

func TestAX7_ManifestMigration_Migrated_Bad(testingT *core.T) {
	migration := ManifestMigration{}
	got := migration.Migrated()
	core.AssertFalse(testingT, got)
}

func TestAX7_ManifestMigration_Migrated_Ugly(testingT *core.T) {
	migration := ManifestMigration{Applied: []string{""}}
	got := migration.Migrated()
	core.AssertTrue(testingT, got)
}

func TestAX7_MigrateManifest_Good(testingT *core.T) {
	manifest := Manifest{Name: "demo", Verification: Verification{Chain: "checksums.sha256"}}
	migrated, migration, err := MigrateManifest(manifest)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, CurrentManifestFormatVersion, migrated.FormatVersion)
	core.AssertTrue(testingT, migration.Migrated())
}

func TestAX7_MigrateManifest_Bad(testingT *core.T) {
	manifest := Manifest{FormatVersion: "future"}
	_, _, err := MigrateManifest(manifest)
	core.AssertError(testingT, err, "manifest/format-version-unsupported")
}

func TestAX7_MigrateManifest_Ugly(testingT *core.T) {
	manifest := Manifest{FormatVersion: CurrentManifestFormatVersion, Preservation: Preservation{Chain: "chain.txt"}}
	migrated, migration, err := MigrateManifest(manifest)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "chain.txt", migrated.Verification.Chain)
	core.AssertTrue(testingT, migration.Migrated())
}

func TestAX7_ParseError_Error_Good(testingT *core.T) {
	err := ParseError{Kind: "manifest/multiple-documents", Message: "manifest must contain exactly one YAML document"}
	got := err.Error()
	core.AssertContains(testingT, got, "manifest/multiple-documents")
}

func TestAX7_ParseError_Error_Bad(testingT *core.T) {
	err := ParseError{Kind: "manifest/empty", Message: "empty manifest"}
	got := err.Error()
	core.AssertEqual(testingT, "manifest/empty: empty manifest", got)
}

func TestAX7_ParseError_Error_Ugly(testingT *core.T) {
	err := ParseError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_FramePipeline_Process_Good(testingT *core.T) {
	pipeline := FramePipeline{Primary: acceleratedFrameProcessor{name: "metal", available: true}}
	result, err := pipeline.Process(validRGBAFrame(), FramePolicy{Mode: AccelerationAuto})
	core.AssertNoError(testingT, err)
	core.AssertTrue(testingT, result.Accelerated)
}

func TestAX7_FramePipeline_Process_Bad(testingT *core.T) {
	pipeline := FramePipeline{}
	_, err := pipeline.Process(FrameBuffer{}, FramePolicy{})
	core.AssertError(testingT, err, "frame/")
}

func TestAX7_FramePipeline_Process_Ugly(testingT *core.T) {
	pipeline := FramePipeline{Primary: acceleratedFrameProcessor{name: "metal", available: false}, Fallback: acceleratedFrameProcessor{name: "cpu"}}
	result, err := pipeline.Process(validRGBAFrame(), FramePolicy{Mode: AccelerationAuto})
	core.AssertNoError(testingT, err)
	core.AssertTrue(testingT, result.Fallback)
}

func TestAX7_FrameProcessor_Name_Good(testingT *core.T) {
	processor := identityFrameProcessor{}
	got := processor.Name()
	core.AssertEqual(testingT, "identity", got)
}

func TestAX7_FrameProcessor_Name_Bad(testingT *core.T) {
	processor := identityFrameProcessor{}
	got := processor.Name()
	core.AssertNotEqual(testingT, "metal", got)
}

func TestAX7_FrameProcessor_Name_Ugly(testingT *core.T) {
	processor := identityFrameProcessor{}
	first := processor.Name()
	second := processor.Name()
	core.AssertEqual(testingT, first, second)
}

func TestAX7_FrameProcessor_Available_Good(testingT *core.T) {
	processor := identityFrameProcessor{}
	got := processor.Available()
	core.AssertTrue(testingT, got)
}

func TestAX7_FrameProcessor_Available_Bad(testingT *core.T) {
	processor := acceleratedFrameProcessor{name: "metal", available: false}
	got := processor.Available()
	core.AssertFalse(testingT, got)
}

func TestAX7_FrameProcessor_Available_Ugly(testingT *core.T) {
	processor := identityFrameProcessor{}
	first := processor.Available()
	second := processor.Available()
	core.AssertEqual(testingT, first, second)
}

func TestAX7_FrameProcessor_Supports_Good(testingT *core.T) {
	processor := identityFrameProcessor{}
	got := processor.Supports(validRGBAFrame(), FramePolicy{})
	core.AssertTrue(testingT, got)
}

func TestAX7_FrameProcessor_Supports_Bad(testingT *core.T) {
	processor := acceleratedFrameProcessor{name: "metal", available: true}
	got := processor.Supports(FrameBuffer{}, FramePolicy{})
	core.AssertTrue(testingT, got)
}

func TestAX7_FrameProcessor_Supports_Ugly(testingT *core.T) {
	processor := identityFrameProcessor{}
	got := processor.Supports(FrameBuffer{}, FramePolicy{Mode: AccelerationRequired})
	core.AssertTrue(testingT, got)
}

func TestAX7_FrameProcessor_Process_Good(testingT *core.T) {
	processor := identityFrameProcessor{}
	frame, err := processor.Process(validRGBAFrame(), FramePolicy{})
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, 320, frame.Width)
}

func TestAX7_FrameProcessor_Process_Bad(testingT *core.T) {
	processor := acceleratedFrameProcessor{processErr: core.NewError("processor failed")}
	_, err := processor.Process(validRGBAFrame(), FramePolicy{})
	core.AssertError(testingT, err, "processor failed")
}

func TestAX7_FrameProcessor_Process_Ugly(testingT *core.T) {
	processor := identityFrameProcessor{}
	frame, err := processor.Process(FrameBuffer{}, FramePolicy{})
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, 0, frame.Width)
}

func TestAX7_Manifest_FramePolicy_Good(testingT *core.T) {
	manifest := Manifest{Runtime: Runtime{Acceleration: AccelerationRequired, Filter: FrameFilterCRT}}
	policy := manifest.FramePolicy()
	core.AssertEqual(testingT, AccelerationRequired, policy.Mode)
	core.AssertEqual(testingT, FrameFilterCRT, policy.Filter)
}

func TestAX7_Manifest_FramePolicy_Bad(testingT *core.T) {
	manifest := Manifest{Runtime: Runtime{Acceleration: AccelerationMode("bad"), Filter: FrameFilter("bad")}}
	policy := manifest.FramePolicy()
	core.AssertEqual(testingT, AccelerationAuto, policy.Mode)
	core.AssertEqual(testingT, FrameFilterNone, policy.Filter)
}

func TestAX7_Manifest_FramePolicy_Ugly(testingT *core.T) {
	manifest := Manifest{}
	policy := manifest.FramePolicy()
	core.AssertEqual(testingT, AccelerationAuto, policy.Mode)
	core.AssertEqual(testingT, FrameFilterNone, policy.Filter)
}

func TestAX7_PipelineError_Error_Good(testingT *core.T) {
	err := PipelineError{Kind: "frame/fallback-failed", Message: "cpu failed"}
	got := err.Error()
	core.AssertContains(testingT, got, "cpu failed")
}

func TestAX7_PipelineError_Error_Bad(testingT *core.T) {
	err := PipelineError{Kind: "frame/acceleration-required", Message: "missing"}
	got := err.Error()
	core.AssertEqual(testingT, "frame/acceleration-required: missing", got)
}

func TestAX7_PipelineError_Error_Ugly(testingT *core.T) {
	err := PipelineError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_BundlePlan_Render_Good(testingT *core.T) {
	plan, issues := NewService(nil, nil).PlanBundle(ax7BundleRequest("render-good"))
	core.RequireTrue(testingT, !issues.HasIssues())
	rendered, err := plan.Render()
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, rendered.Files)
}

func TestAX7_BundlePlan_Render_Bad(testingT *core.T) {
	plan := BundlePlan{}
	rendered, err := plan.Render()
	core.AssertError(testingT, err, "manifest/")
	core.AssertEmpty(testingT, rendered.Files)
}

func TestAX7_BundlePlan_Render_Ugly(testingT *core.T) {
	request := ax7BundleRequest("render-ugly")
	request.EngineBinaryPath = "engine/synthetic"
	request.EngineBinaryData = []byte("engine")
	request.EngineBinarySHA256 = hashHex([]byte("engine"))
	plan, issues := NewService(nil, nil).PlanBundle(request)
	core.RequireTrue(testingT, !issues.HasIssues())
	rendered, err := plan.Render()
	core.AssertNoError(testingT, err)
	core.AssertNotNil(testingT, renderedFileData(rendered, "engine/synthetic"))
}

func TestAX7_SandboxPolicy_ValidateLaunch_Good(testingT *core.T) {
	policy := SandboxPolicy{ReadPaths: []string{"rom/"}, WritePaths: []string{"saves/"}, Resources: ResourceLimits{CPUPercent: 50}}
	issues := policy.ValidateLaunch(LaunchPlan{Entrypoint: "rom/game.bin", WritePaths: []string{"saves/state"}, Resources: ResourceLimits{CPUPercent: 25}})
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_SandboxPolicy_ValidateLaunch_Bad(testingT *core.T) {
	policy := SandboxPolicy{}
	issues := policy.ValidateLaunch(LaunchPlan{NetworkAllowed: true})
	core.AssertTrue(testingT, hasIssueCode(issues, "sandbox/network-denied"))
}

func TestAX7_SandboxPolicy_ValidateLaunch_Ugly(testingT *core.T) {
	policy := SandboxPolicy{ReadPaths: []string{"rom/"}, WritePaths: []string{"saves/"}, Resources: ResourceLimits{MemoryBytes: 128}}
	issues := policy.ValidateLaunch(LaunchPlan{ReadPaths: []string{"../escape"}, WritePaths: []string{"screenshots/"}, Resources: ResourceLimits{MemoryBytes: 256}})
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_SandboxError_Error_Good(testingT *core.T) {
	err := SandboxError{Kind: "sandbox/directory-create-failed", Path: "saves", Message: "denied"}
	got := err.Error()
	core.AssertContains(testingT, got, "saves")
}

func TestAX7_SandboxError_Error_Bad(testingT *core.T) {
	err := SandboxError{Kind: "sandbox/home-required", Message: "home directory is required"}
	got := err.Error()
	core.AssertEqual(testingT, "sandbox/home-required: home directory is required", got)
}

func TestAX7_SandboxError_Error_Ugly(testingT *core.T) {
	err := SandboxError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_Shield_Verify_Good(testingT *core.T) {
	bundle := shieldBundle(testingT, []byte("shield-rom"))
	report := (Shield{Registry: ax7Registry(testingT, SyntheticEngine{})}).Verify(bundle)
	core.AssertTrue(testingT, report.OverallOK)
}

func TestAX7_Shield_Verify_Bad(testingT *core.T) {
	bundle := shieldBundle(testingT, []byte("shield-rom"))
	report := (Shield{Registry: NewRegistry()}).Verify(bundle)
	core.AssertFalse(testingT, report.OverallOK)
}

func TestAX7_Shield_Verify_Ugly(testingT *core.T) {
	bundle := shieldBundle(testingT, []byte("shield-rom"))
	bundle.Manifest.Artefact.Path = "../escape"
	report := (Shield{Registry: ax7Registry(testingT, SyntheticEngine{})}).Verify(bundle)
	core.AssertFalse(testingT, report.OverallOK)
}

func TestAX7_ShieldReport_Issues_Good(testingT *core.T) {
	report := ShieldReport{SBOM: SBOMResult{Valid: true}, Code: CodeResult{OK: true}, Content: ContentResult{OK: true}, Threat: ThreatResult{OK: true}}
	issues := report.Issues()
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_ShieldReport_Issues_Bad(testingT *core.T) {
	report := ShieldReport{SBOM: SBOMResult{Issues: ValidationErrors{{Code: "sbom/missing"}}}}
	issues := report.Issues()
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_ShieldReport_Issues_Ugly(testingT *core.T) {
	report := ShieldReport{Threat: ThreatResult{Issues: ValidationErrors{{Code: "threat/test"}}}}
	issues := report.Issues()
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_NewService_Good(testingT *core.T) {
	registry := NewRegistry()
	service := NewService(validBundleFS(), registry)
	core.AssertEqual(testingT, registry, service.Registry)
}

func TestAX7_NewService_Bad(testingT *core.T) {
	service := NewService(nil, nil)
	core.AssertNil(testingT, service.Bundles)
	core.AssertNotNil(testingT, service.Registry)
}

func TestAX7_NewService_Ugly(testingT *core.T) {
	service := NewService(validBundleFS(), nil)
	core.AssertNotNil(testingT, service.Bundles)
	core.AssertNotNil(testingT, service.Registry)
}

func TestAX7_Service_ListBundles_Good(testingT *core.T) {
	service := NewService(listBundleFS(), ax7Registry(testingT, SyntheticEngine{}))
	summaries, err := service.ListBundles(ListRequest{Root: "."})
	core.AssertNoError(testingT, err)
	core.AssertLen(testingT, summaries, 2)
}

func TestAX7_Service_ListBundles_Bad(testingT *core.T) {
	service := NewService(nil, NewRegistry())
	summaries, err := service.ListBundles(ListRequest{Root: "."})
	core.AssertError(testingT, err, "bundle/filesystem-missing")
	core.AssertNil(testingT, summaries)
}

func TestAX7_Service_ListBundles_Ugly(testingT *core.T) {
	service := NewService(listBundleFS(), NewRegistry())
	summaries, err := service.ListBundles(ListRequest{Root: "missing"})
	core.AssertError(testingT, err)
	core.AssertNil(testingT, summaries)
}

func TestAX7_Service_VerifyBundle_Good(testingT *core.T) {
	service := NewService(verifiedBundleFS(), ax7Registry(testingT, RetroArchEngine{Binary: "retroarch"}))
	result, err := service.VerifyBundle(VerifyRequest{BundlePath: "."})
	core.AssertNoError(testingT, err)
	core.AssertTrue(testingT, result.Verified)
}

func TestAX7_Service_VerifyBundle_Bad(testingT *core.T) {
	service := NewService(nil, NewRegistry())
	_, err := service.VerifyBundle(VerifyRequest{BundlePath: "."})
	core.AssertError(testingT, err, "bundle/filesystem-missing")
}

func TestAX7_Service_VerifyBundle_Ugly(testingT *core.T) {
	service := NewService(verifiedBundleFS(), NewRegistry())
	result, err := service.VerifyBundle(VerifyRequest{BundlePath: "."})
	core.AssertNoError(testingT, err)
	core.AssertFalse(testingT, result.Verified)
}

func TestAX7_Service_PreparePlay_Good(testingT *core.T) {
	service := NewService(verifiedBundleFS(), ax7Registry(testingT, RetroArchEngine{Binary: "retroarch"}))
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	core.AssertNoError(testingT, err)
	core.AssertTrue(testingT, plan.Ready)
}

func TestAX7_Service_PreparePlay_Bad(testingT *core.T) {
	service := NewService(nil, NewRegistry())
	_, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	core.AssertError(testingT, err, "bundle/filesystem-missing")
}

func TestAX7_Service_PreparePlay_Ugly(testingT *core.T) {
	service := NewService(unsupportedProfileBundleFS(), ax7Registry(testingT, RetroArchEngine{Binary: "retroarch"}))
	plan, err := service.PreparePlay(PlayRequest{BundlePath: "."})
	core.AssertNoError(testingT, err)
	core.AssertFalse(testingT, plan.Ready)
}

func TestAX7_Service_PlanBundle_Good(testingT *core.T) {
	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(ax7BundleRequest("plan-good"))
	core.AssertFalse(testingT, issues.HasIssues())
	core.AssertEqual(testingT, "plan-good", plan.Path)
}

func TestAX7_Service_PlanBundle_Bad(testingT *core.T) {
	service := NewService(nil, nil)
	_, issues := service.PlanBundle(BundleRequest{})
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_Service_PlanBundle_Ugly(testingT *core.T) {
	request := ax7BundleRequest("plan-ugly")
	request.ArtefactData[0] = 'x'
	service := NewService(nil, nil)
	plan, issues := service.PlanBundle(request)
	core.AssertFalse(testingT, issues.HasIssues())
	core.AssertEqual(testingT, []byte("xom"), plan.ArtefactData)
}

func TestAX7_Service_RenderBundle_Good(testingT *core.T) {
	rendered, err := NewService(nil, nil).RenderBundle(ax7BundleRequest("render-service-good"))
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, rendered.Files)
}

func TestAX7_Service_RenderBundle_Bad(testingT *core.T) {
	rendered, err := NewService(nil, nil).RenderBundle(BundleRequest{})
	core.AssertError(testingT, err, "manifest/")
	core.AssertEmpty(testingT, rendered.Files)
}

func TestAX7_Service_RenderBundle_Ugly(testingT *core.T) {
	request := ax7BundleRequest("render-service-ugly")
	request.BYOROM = true
	rendered, err := NewService(nil, nil).RenderBundle(request)
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, rendered.Files)
}

func TestAX7_Service_WriteBundle_Good(testingT *core.T) {
	writer := newMemoryBundleWriter()
	err := NewService(nil, nil).WriteBundle(ax7BundleRequest("write-service-good"), writer)
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, writer.files)
}

func TestAX7_Service_WriteBundle_Bad(testingT *core.T) {
	err := NewService(nil, nil).WriteBundle(BundleRequest{}, newMemoryBundleWriter())
	core.AssertError(testingT, err, "manifest/")
	core.AssertNotNil(testingT, err)
}

func TestAX7_Service_WriteBundle_Ugly(testingT *core.T) {
	err := NewService(nil, nil).WriteBundle(ax7BundleRequest("write-service-ugly"), failingwriter{failWrite: true})
	core.AssertError(testingT, err, "bundle/file-write-failed")
	core.AssertNotNil(testingT, err)
}

func TestAX7_Manifest_Validate_Good(testingT *core.T) {
	manifest, err := LoadManifest([]byte(validManifestYAML()))
	core.RequireNoError(testingT, err)
	issues := manifest.Validate()
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_Manifest_Validate_Bad(testingT *core.T) {
	manifest := Manifest{}
	issues := manifest.Validate()
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_Manifest_Validate_Ugly(testingT *core.T) {
	manifest, err := LoadManifest([]byte(validManifestYAML()))
	core.RequireNoError(testingT, err)
	manifest.Resources.CPUPercent = -1
	issues := manifest.Validate()
	core.AssertTrue(testingT, hasIssueCode(issues, "manifest/resources-cpu-invalid"))
}

func TestAX7_ValidationIssue_Error_Good(testingT *core.T) {
	issue := ValidationIssue{Code: "manifest/name-required", Field: "name", Message: "name is required"}
	got := issue.Error()
	core.AssertEqual(testingT, "manifest/name-required: name: name is required", got)
}

func TestAX7_ValidationIssue_Error_Bad(testingT *core.T) {
	issue := ValidationIssue{Code: "manifest/name-required"}
	got := issue.Error()
	core.AssertEqual(testingT, "manifest/name-required", got)
}

func TestAX7_ValidationIssue_Error_Ugly(testingT *core.T) {
	issue := ValidationIssue{Code: "manifest/name-required", Message: "name is required"}
	got := issue.Error()
	core.AssertEqual(testingT, "manifest/name-required: name is required", got)
}

func TestAX7_ValidationErrors_Error_Good(testingT *core.T) {
	issues := ValidationErrors{{Code: "one"}, {Code: "two"}}
	got := issues.Error()
	core.AssertEqual(testingT, "one; two", got)
}

func TestAX7_ValidationErrors_Error_Bad(testingT *core.T) {
	issues := ValidationErrors{}
	got := issues.Error()
	core.AssertEqual(testingT, "", got)
}

func TestAX7_ValidationErrors_Error_Ugly(testingT *core.T) {
	issues := ValidationErrors{{Code: "one", Field: "field", Message: "message"}}
	got := issues.Error()
	core.AssertContains(testingT, got, "field")
}

func TestAX7_ValidationErrors_HasIssues_Good(testingT *core.T) {
	issues := ValidationErrors{{Code: "one"}}
	got := issues.HasIssues()
	core.AssertTrue(testingT, got)
}

func TestAX7_ValidationErrors_HasIssues_Bad(testingT *core.T) {
	issues := ValidationErrors{}
	got := issues.HasIssues()
	core.AssertFalse(testingT, got)
}

func TestAX7_ValidationErrors_HasIssues_Ugly(testingT *core.T) {
	var issues ValidationErrors
	got := issues.HasIssues()
	core.AssertFalse(testingT, got)
}

func TestAX7_Bundle_Verify_Good(testingT *core.T) {
	core.RequireNoError(testingT, RegisterEngine(RetroArchEngine{Binary: "retroarch"}))
	bundle := ax7Bundle(testingT, verifiedBundleFS())
	issues := bundle.Verify()
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_Bundle_Verify_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, brokenChecksumBundleFS())
	issues := bundle.VerifyWithRegistry(ax7Registry(testingT, RetroArchEngine{Binary: "retroarch"}))
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_Bundle_Verify_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, verifiedBundleFS())
	issues := bundle.VerifyWithRegistry(nil)
	core.AssertTrue(testingT, hasIssueCode(issues, "engine/registry-missing"))
}

func TestAX7_Bundle_VerifyWithRegistry_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, verifiedBundleFS())
	issues := bundle.VerifyWithRegistry(ax7Registry(testingT, RetroArchEngine{Binary: "retroarch"}))
	core.AssertFalse(testingT, issues.HasIssues())
}

func TestAX7_Bundle_VerifyWithRegistry_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, verifiedBundleFS())
	issues := bundle.VerifyWithRegistry(NewRegistry())
	core.AssertTrue(testingT, hasIssueCode(issues, "engine/unavailable"))
}

func TestAX7_Bundle_VerifyWithRegistry_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, brokenChecksumBundleFS())
	issues := bundle.VerifyWithRegistry(ax7Registry(testingT, RetroArchEngine{Binary: "retroarch"}))
	core.AssertTrue(testingT, issues.HasIssues())
}

func TestAX7_ParseChecksumFile_Good(testingT *core.T) {
	entries, err := ParseChecksumFile([]byte(validArtefactSHA256 + "  rom/game.bin\n"))
	core.AssertNoError(testingT, err)
	core.AssertLen(testingT, entries, 1)
}

func TestAX7_ParseChecksumFile_Bad(testingT *core.T) {
	entries, err := ParseChecksumFile([]byte("not-a-checksum\n"))
	core.AssertError(testingT, err, "checksum/invalid-line")
	core.AssertNil(testingT, entries)
}

func TestAX7_ParseChecksumFile_Ugly(testingT *core.T) {
	entries, err := ParseChecksumFile([]byte("# comment\n\n" + validArtefactSHA256 + "  rom/game.bin\n"))
	core.AssertNoError(testingT, err)
	core.AssertLen(testingT, entries, 1)
}

func TestAX7_ChecksumParseError_Error_Good(testingT *core.T) {
	err := ChecksumParseError{Kind: "checksum/invalid-hash", Message: "checksum entry must contain a valid sha256 value"}
	got := err.Error()
	core.AssertContains(testingT, got, "checksum/invalid-hash")
}

func TestAX7_ChecksumParseError_Error_Bad(testingT *core.T) {
	err := ChecksumParseError{Kind: "checksum/invalid-line", Message: "checksum lines must contain a sha256 value and path"}
	got := err.Error()
	core.AssertEqual(testingT, "checksum/invalid-line: checksum lines must contain a sha256 value and path", got)
}

func TestAX7_ChecksumParseError_Error_Ugly(testingT *core.T) {
	err := ChecksumParseError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_RenderedBundle_Write_Good(testingT *core.T) {
	rendered := shieldRendered(testingT, []byte("rom"))
	writer := newMemoryBundleWriter()
	err := rendered.Write(writer)
	core.AssertNoError(testingT, err)
	core.AssertNotEmpty(testingT, writer.files)
}

func TestAX7_RenderedBundle_Write_Bad(testingT *core.T) {
	rendered := RenderedBundle{Path: "game"}
	err := rendered.Write(nil)
	core.AssertError(testingT, err, "bundle/writer-missing")
}

func TestAX7_RenderedBundle_Write_Ugly(testingT *core.T) {
	rendered := RenderedBundle{Path: "game", Files: []RenderedFile{{Path: "rom/game.bin", Data: []byte("rom")}}}
	err := rendered.Write(failingwriter{failEnsure: true})
	core.AssertError(testingT, err, "bundle/root-create-failed")
}

func TestAX7_WriteError_Error_Good(testingT *core.T) {
	err := WriteError{Kind: "bundle/file-write-failed", Path: "rom/game.bin", Message: "denied"}
	got := err.Error()
	core.AssertContains(testingT, got, "rom/game.bin")
}

func TestAX7_WriteError_Error_Bad(testingT *core.T) {
	err := WriteError{Kind: "bundle/path-required", Message: "bundle path is required"}
	got := err.Error()
	core.AssertEqual(testingT, "bundle/path-required: bundle path is required", got)
}

func TestAX7_WriteError_Error_Ugly(testingT *core.T) {
	err := WriteError{}
	got := err.Error()
	core.AssertContains(testingT, got, ": ")
}

func TestAX7_DOSBoxEngine_Name_Good(testingT *core.T) {
	engine := DOSBoxEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "dosbox", got)
}

func TestAX7_DOSBoxEngine_Name_Bad(testingT *core.T) {
	engine := DOSBoxEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "retroarch", got)
}

func TestAX7_DOSBoxEngine_Name_Ugly(testingT *core.T) {
	engine := DOSBoxEngine{Binary: "custom-dosbox"}
	got := engine.Name()
	core.AssertEqual(testingT, "dosbox", got)
}

func TestAX7_DOSBoxEngine_Platforms_Good(testingT *core.T) {
	engine := DOSBoxEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "dos")
}

func TestAX7_DOSBoxEngine_Platforms_Bad(testingT *core.T) {
	engine := DOSBoxEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "snes")
}

func TestAX7_DOSBoxEngine_Platforms_Ugly(testingT *core.T) {
	engine := DOSBoxEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 1)
}

func TestAX7_DOSBoxEngine_Acceleration_Good(testingT *core.T) {
	engine := DOSBoxEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_DOSBoxEngine_Acceleration_Bad(testingT *core.T) {
	engine := DOSBoxEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationRequired, acceleration.Mode)
}

func TestAX7_DOSBoxEngine_Acceleration_Ugly(testingT *core.T) {
	engine := DOSBoxEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterBilinear)
}

func TestAX7_DOSBoxEngine_Verify_Good(testingT *core.T) {
	engine := DOSBoxEngine{Binary: "dosbox"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_DOSBoxEngine_Verify_Bad(testingT *core.T) {
	engine := DOSBoxEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_DOSBoxEngine_Verify_Ugly(testingT *core.T) {
	engine := DOSBoxEngine{Binary: "bin/dosbox"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_DOSBoxEngine_CodeIdentity_Good(testingT *core.T) {
	engine := DOSBoxEngine{Binary: "dosbox", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_DOSBoxEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := DOSBoxEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "dosbox", identity.Path)
}

func TestAX7_DOSBoxEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := DOSBoxEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_DOSBoxEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (DOSBoxEngine{Binary: "dosbox"}).Run("rom/game.exe", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_DOSBoxEngine_Run_Bad(testingT *core.T) {
	err := (DOSBoxEngine{}).Run("rom/game.exe", ax7EngineConfig(""))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_DOSBoxEngine_Run_Ugly(testingT *core.T) {
	err := (DOSBoxEngine{Binary: "dosbox"}).Run("rom/game.exe", EngineConfig{})
	core.AssertError(testingT, err, "engine/process-unavailable")
	core.AssertNotNil(testingT, err)
}

func TestAX7_DOSBoxEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, dosBundleFS())
	plan, err := (DOSBoxEngine{Binary: "dosbox"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "dosbox", plan.Engine)
}

func TestAX7_DOSBoxEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, dosBundleFS())
	_, err := (DOSBoxEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_DOSBoxEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, dosBundleFS())
	bundle.Manifest.Platform = "snes"
	_, err := (DOSBoxEngine{Binary: "dosbox"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/platform-unsupported")
}

func TestAX7_DOSBoxXEngine_Name_Good(testingT *core.T) {
	engine := DOSBoxXEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "dosbox-x", got)
}

func TestAX7_DOSBoxXEngine_Name_Bad(testingT *core.T) {
	engine := DOSBoxXEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "dosbox", got)
}

func TestAX7_DOSBoxXEngine_Name_Ugly(testingT *core.T) {
	engine := DOSBoxXEngine{Binary: "custom"}
	got := engine.Name()
	core.AssertEqual(testingT, "dosbox-x", got)
}

func TestAX7_DOSBoxXEngine_Platforms_Good(testingT *core.T) {
	engine := DOSBoxXEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "windows-9x")
}

func TestAX7_DOSBoxXEngine_Platforms_Bad(testingT *core.T) {
	engine := DOSBoxXEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "arcade")
}

func TestAX7_DOSBoxXEngine_Platforms_Ugly(testingT *core.T) {
	engine := DOSBoxXEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 4)
}

func TestAX7_DOSBoxXEngine_Acceleration_Good(testingT *core.T) {
	engine := DOSBoxXEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_DOSBoxXEngine_Acceleration_Bad(testingT *core.T) {
	engine := DOSBoxXEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationOff, acceleration.Mode)
}

func TestAX7_DOSBoxXEngine_Acceleration_Ugly(testingT *core.T) {
	engine := DOSBoxXEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterCRT)
}

func TestAX7_DOSBoxXEngine_Verify_Good(testingT *core.T) {
	engine := DOSBoxXEngine{Binary: "dosbox-x"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_DOSBoxXEngine_Verify_Bad(testingT *core.T) {
	engine := DOSBoxXEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_DOSBoxXEngine_Verify_Ugly(testingT *core.T) {
	engine := DOSBoxXEngine{Binary: "bin/dosbox-x"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_DOSBoxXEngine_CodeIdentity_Good(testingT *core.T) {
	engine := DOSBoxXEngine{Binary: "dosbox-x", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_DOSBoxXEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := DOSBoxXEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "dosbox-x", identity.Path)
}

func TestAX7_DOSBoxXEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := DOSBoxXEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_DOSBoxXEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (DOSBoxXEngine{Binary: "dosbox-x"}).Run("rom/game.exe", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Profile: "dos", Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_DOSBoxXEngine_Run_Bad(testingT *core.T) {
	err := (DOSBoxXEngine{}).Run("rom/game.exe", ax7EngineConfig("dos"))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_DOSBoxXEngine_Run_Ugly(testingT *core.T) {
	err := (DOSBoxXEngine{Binary: "dosbox-x"}).Run("rom/game.exe", EngineConfig{Profile: "unknown"})
	core.AssertError(testingT, err, "engine/profile-unsupported")
	core.AssertNotNil(testingT, err)
}

func TestAX7_DOSBoxXEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, dosBoxXBundleFS())
	plan, err := (DOSBoxXEngine{Binary: "dosbox-x"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "dosbox-x", plan.Engine)
}

func TestAX7_DOSBoxXEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, dosBoxXBundleFS())
	_, err := (DOSBoxXEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_DOSBoxXEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, dosBoxXBundleFS())
	bundle.Manifest.Runtime.Profile = "unknown"
	_, err := (DOSBoxXEngine{Binary: "dosbox-x"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/profile-unsupported")
}

func TestAX7_FUSEEngine_Name_Good(testingT *core.T) {
	engine := FUSEEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "fuse", got)
}

func TestAX7_FUSEEngine_Name_Bad(testingT *core.T) {
	engine := FUSEEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "vice", got)
}

func TestAX7_FUSEEngine_Name_Ugly(testingT *core.T) {
	engine := FUSEEngine{Binary: "custom"}
	got := engine.Name()
	core.AssertEqual(testingT, "fuse", got)
}

func TestAX7_FUSEEngine_Platforms_Good(testingT *core.T) {
	engine := FUSEEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "zx-spectrum")
}

func TestAX7_FUSEEngine_Platforms_Bad(testingT *core.T) {
	engine := FUSEEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "commodore-64")
}

func TestAX7_FUSEEngine_Platforms_Ugly(testingT *core.T) {
	engine := FUSEEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 4)
}

func TestAX7_FUSEEngine_Acceleration_Good(testingT *core.T) {
	engine := FUSEEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_FUSEEngine_Acceleration_Bad(testingT *core.T) {
	engine := FUSEEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationRequired, acceleration.Mode)
}

func TestAX7_FUSEEngine_Acceleration_Ugly(testingT *core.T) {
	engine := FUSEEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterCRT)
}

func TestAX7_FUSEEngine_Verify_Good(testingT *core.T) {
	engine := FUSEEngine{Binary: "fuse"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_FUSEEngine_Verify_Bad(testingT *core.T) {
	engine := FUSEEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_FUSEEngine_Verify_Ugly(testingT *core.T) {
	engine := FUSEEngine{Binary: "bin/fuse"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_FUSEEngine_CodeIdentity_Good(testingT *core.T) {
	engine := FUSEEngine{Binary: "fuse", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_FUSEEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := FUSEEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "fuse", identity.Path)
}

func TestAX7_FUSEEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := FUSEEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_FUSEEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (FUSEEngine{Binary: "fuse"}).Run("rom/game.tap", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Profile: "48k", Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_FUSEEngine_Run_Bad(testingT *core.T) {
	err := (FUSEEngine{}).Run("rom/game.tap", ax7EngineConfig("48k"))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_FUSEEngine_Run_Ugly(testingT *core.T) {
	err := (FUSEEngine{Binary: "fuse"}).Run("rom/game.tap", EngineConfig{Profile: "unknown"})
	core.AssertError(testingT, err, "engine/profile-unsupported")
	core.AssertNotNil(testingT, err)
}

func TestAX7_FUSEEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, fuseBundleFS(testingT))
	plan, err := (FUSEEngine{Binary: "fuse"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "fuse", plan.Engine)
}

func TestAX7_FUSEEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, fuseBundleFS(testingT))
	_, err := (FUSEEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_FUSEEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, fuseBundleFS(testingT))
	bundle.Manifest.Runtime.Profile = "unknown"
	_, err := (FUSEEngine{Binary: "fuse"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/profile-unsupported")
}

func TestAX7_MAMEEngine_Name_Good(testingT *core.T) {
	engine := MAMEEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "mame", got)
}

func TestAX7_MAMEEngine_Name_Bad(testingT *core.T) {
	engine := MAMEEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "fuse", got)
}

func TestAX7_MAMEEngine_Name_Ugly(testingT *core.T) {
	engine := MAMEEngine{Binary: "custom"}
	got := engine.Name()
	core.AssertEqual(testingT, "mame", got)
}

func TestAX7_MAMEEngine_Platforms_Good(testingT *core.T) {
	engine := MAMEEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "arcade")
}

func TestAX7_MAMEEngine_Platforms_Bad(testingT *core.T) {
	engine := MAMEEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "dos")
}

func TestAX7_MAMEEngine_Platforms_Ugly(testingT *core.T) {
	engine := MAMEEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 2)
}

func TestAX7_MAMEEngine_Acceleration_Good(testingT *core.T) {
	engine := MAMEEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_MAMEEngine_Acceleration_Bad(testingT *core.T) {
	engine := MAMEEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationOff, acceleration.Mode)
}

func TestAX7_MAMEEngine_Acceleration_Ugly(testingT *core.T) {
	engine := MAMEEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterScanline)
}

func TestAX7_MAMEEngine_Verify_Good(testingT *core.T) {
	engine := MAMEEngine{Binary: "mame"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_MAMEEngine_Verify_Bad(testingT *core.T) {
	engine := MAMEEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_MAMEEngine_Verify_Ugly(testingT *core.T) {
	engine := MAMEEngine{Binary: "bin/mame"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_MAMEEngine_CodeIdentity_Good(testingT *core.T) {
	engine := MAMEEngine{Binary: "mame", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_MAMEEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := MAMEEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "mame", identity.Path)
}

func TestAX7_MAMEEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := MAMEEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_MAMEEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (MAMEEngine{Binary: "mame"}).Run("rom/pacman.zip", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Profile: "pacman", Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_MAMEEngine_Run_Bad(testingT *core.T) {
	err := (MAMEEngine{}).Run("rom/pacman.zip", ax7EngineConfig("pacman"))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_MAMEEngine_Run_Ugly(testingT *core.T) {
	err := (MAMEEngine{Binary: "mame"}).Run("rom/pacman.zip", EngineConfig{})
	core.AssertError(testingT, err, "engine/profile-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_MAMEEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, mameBundleFS(testingT))
	plan, err := (MAMEEngine{Binary: "mame"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "mame", plan.Engine)
}

func TestAX7_MAMEEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, mameBundleFS(testingT))
	_, err := (MAMEEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_MAMEEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, mameBundleFS(testingT))
	bundle.Manifest.Runtime.Profile = ""
	_, err := (MAMEEngine{Binary: "mame"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/profile-required")
}

func TestAX7_RetroArchEngine_Name_Good(testingT *core.T) {
	engine := RetroArchEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "retroarch", got)
}

func TestAX7_RetroArchEngine_Name_Bad(testingT *core.T) {
	engine := RetroArchEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "snes9x", got)
}

func TestAX7_RetroArchEngine_Name_Ugly(testingT *core.T) {
	engine := RetroArchEngine{Binary: "custom"}
	got := engine.Name()
	core.AssertEqual(testingT, "retroarch", got)
}

func TestAX7_RetroArchEngine_Platforms_Good(testingT *core.T) {
	engine := RetroArchEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "sega-genesis")
}

func TestAX7_RetroArchEngine_Platforms_Bad(testingT *core.T) {
	engine := RetroArchEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "scummvm")
}

func TestAX7_RetroArchEngine_Platforms_Ugly(testingT *core.T) {
	engine := RetroArchEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "gba")
}

func TestAX7_RetroArchEngine_Acceleration_Good(testingT *core.T) {
	engine := RetroArchEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_RetroArchEngine_Acceleration_Bad(testingT *core.T) {
	engine := RetroArchEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationRequired, acceleration.Mode)
}

func TestAX7_RetroArchEngine_Acceleration_Ugly(testingT *core.T) {
	engine := RetroArchEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterCRT)
}

func TestAX7_RetroArchEngine_Verify_Good(testingT *core.T) {
	engine := RetroArchEngine{Binary: "retroarch"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_RetroArchEngine_Verify_Bad(testingT *core.T) {
	engine := RetroArchEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_RetroArchEngine_Verify_Ugly(testingT *core.T) {
	engine := RetroArchEngine{Binary: "bin/retroarch"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_RetroArchEngine_CodeIdentity_Good(testingT *core.T) {
	engine := RetroArchEngine{Binary: "retroarch", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_RetroArchEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := RetroArchEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "retroarch", identity.Path)
}

func TestAX7_RetroArchEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := RetroArchEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_RetroArchEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (RetroArchEngine{Binary: "retroarch"}).Run("rom/game.zip", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Profile: "genesis", Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_RetroArchEngine_Run_Bad(testingT *core.T) {
	err := (RetroArchEngine{}).Run("rom/game.zip", ax7EngineConfig("genesis"))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_RetroArchEngine_Run_Ugly(testingT *core.T) {
	err := (RetroArchEngine{Binary: "retroarch"}).Run("rom/game.zip", EngineConfig{Profile: "unknown"})
	core.AssertError(testingT, err, "engine/profile-unsupported")
	core.AssertNotNil(testingT, err)
}

func TestAX7_RetroArchEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, validBundleFS())
	plan, err := (RetroArchEngine{Binary: "retroarch"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "retroarch", plan.Engine)
}

func TestAX7_RetroArchEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, validBundleFS())
	_, err := (RetroArchEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_RetroArchEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, validBundleFS())
	bundle.Manifest.Runtime.Profile = "unknown"
	_, err := (RetroArchEngine{Binary: "retroarch"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/profile-unsupported")
}

func TestAX7_ScummVMEngine_Name_Good(testingT *core.T) {
	engine := ScummVMEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "scummvm", got)
}

func TestAX7_ScummVMEngine_Name_Bad(testingT *core.T) {
	engine := ScummVMEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "retroarch", got)
}

func TestAX7_ScummVMEngine_Name_Ugly(testingT *core.T) {
	engine := ScummVMEngine{Binary: "custom"}
	got := engine.Name()
	core.AssertEqual(testingT, "scummvm", got)
}

func TestAX7_ScummVMEngine_Platforms_Good(testingT *core.T) {
	engine := ScummVMEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "scummvm")
}

func TestAX7_ScummVMEngine_Platforms_Bad(testingT *core.T) {
	engine := ScummVMEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "arcade")
}

func TestAX7_ScummVMEngine_Platforms_Ugly(testingT *core.T) {
	engine := ScummVMEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 2)
}

func TestAX7_ScummVMEngine_Acceleration_Good(testingT *core.T) {
	engine := ScummVMEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_ScummVMEngine_Acceleration_Bad(testingT *core.T) {
	engine := ScummVMEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationRequired, acceleration.Mode)
}

func TestAX7_ScummVMEngine_Acceleration_Ugly(testingT *core.T) {
	engine := ScummVMEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterBilinear)
}

func TestAX7_ScummVMEngine_Verify_Good(testingT *core.T) {
	engine := ScummVMEngine{Binary: "scummvm"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_ScummVMEngine_Verify_Bad(testingT *core.T) {
	engine := ScummVMEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_ScummVMEngine_Verify_Ugly(testingT *core.T) {
	engine := ScummVMEngine{Binary: "scummvm", Core: scummVMCore("ScummVM 2.6.1")}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/version-unsupported")
}

func TestAX7_ScummVMEngine_CodeIdentity_Good(testingT *core.T) {
	engine := ScummVMEngine{Binary: "scummvm", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_ScummVMEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := ScummVMEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "scummvm", identity.Path)
}

func TestAX7_ScummVMEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := ScummVMEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_ScummVMEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (ScummVMEngine{Binary: "scummvm"}).Run("game/BASS/sky.dsk", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Profile: "sky", SaveRoot: "saves/", Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_ScummVMEngine_Run_Bad(testingT *core.T) {
	err := (ScummVMEngine{}).Run("game/BASS/sky.dsk", ax7EngineConfig("sky"))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_ScummVMEngine_Run_Ugly(testingT *core.T) {
	err := (ScummVMEngine{Binary: "scummvm"}).Run("game/BASS/sky.dsk", EngineConfig{})
	core.AssertError(testingT, err, "engine/profile-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_ScummVMEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, scummVMBundleFS())
	plan, err := (ScummVMEngine{Binary: "scummvm"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "scummvm", plan.Engine)
}

func TestAX7_ScummVMEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, scummVMBundleFS())
	_, err := (ScummVMEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_ScummVMEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, scummVMBundleFS())
	bundle.Manifest.Runtime.Profile = ""
	_, err := (ScummVMEngine{Binary: "scummvm"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/profile-required")
}

func TestAX7_Snes9xEngine_Name_Good(testingT *core.T) {
	engine := Snes9xEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "snes9x", got)
}

func TestAX7_Snes9xEngine_Name_Bad(testingT *core.T) {
	engine := Snes9xEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "mame", got)
}

func TestAX7_Snes9xEngine_Name_Ugly(testingT *core.T) {
	engine := Snes9xEngine{Binary: "custom"}
	got := engine.Name()
	core.AssertEqual(testingT, "snes9x", got)
}

func TestAX7_Snes9xEngine_Platforms_Good(testingT *core.T) {
	engine := Snes9xEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "snes")
}

func TestAX7_Snes9xEngine_Platforms_Bad(testingT *core.T) {
	engine := Snes9xEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "nes")
}

func TestAX7_Snes9xEngine_Platforms_Ugly(testingT *core.T) {
	engine := Snes9xEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 2)
}

func TestAX7_Snes9xEngine_Acceleration_Good(testingT *core.T) {
	engine := Snes9xEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_Snes9xEngine_Acceleration_Bad(testingT *core.T) {
	engine := Snes9xEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationRequired, acceleration.Mode)
}

func TestAX7_Snes9xEngine_Acceleration_Ugly(testingT *core.T) {
	engine := Snes9xEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterScanline)
}

func TestAX7_Snes9xEngine_Verify_Good(testingT *core.T) {
	engine := Snes9xEngine{Binary: "snes9x"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_Snes9xEngine_Verify_Bad(testingT *core.T) {
	engine := Snes9xEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_Snes9xEngine_Verify_Ugly(testingT *core.T) {
	engine := Snes9xEngine{Binary: "bin/snes9x"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_Snes9xEngine_CodeIdentity_Good(testingT *core.T) {
	engine := Snes9xEngine{Binary: "snes9x", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_Snes9xEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := Snes9xEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "snes9x", identity.Path)
}

func TestAX7_Snes9xEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := Snes9xEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_Snes9xEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (Snes9xEngine{Binary: "snes9x"}).Run("rom/game.sfc", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_Snes9xEngine_Run_Bad(testingT *core.T) {
	err := (Snes9xEngine{}).Run("rom/game.sfc", ax7EngineConfig(""))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_Snes9xEngine_Run_Ugly(testingT *core.T) {
	err := (Snes9xEngine{Binary: "snes9x"}).Run("rom/game.sfc", EngineConfig{})
	core.AssertError(testingT, err, "engine/process-unavailable")
	core.AssertNotNil(testingT, err)
}

func TestAX7_Snes9xEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, snes9xBundleFS(testingT))
	plan, err := (Snes9xEngine{Binary: "snes9x"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "snes9x", plan.Engine)
}

func TestAX7_Snes9xEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, snes9xBundleFS(testingT))
	_, err := (Snes9xEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_Snes9xEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, snes9xBundleFS(testingT))
	bundle.Manifest.Platform = "dos"
	_, err := (Snes9xEngine{Binary: "snes9x"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/platform-unsupported")
}

func TestAX7_SyntheticEngine_Name_Good(testingT *core.T) {
	engine := SyntheticEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "synthetic", got)
}

func TestAX7_SyntheticEngine_Name_Bad(testingT *core.T) {
	engine := SyntheticEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "retroarch", got)
}

func TestAX7_SyntheticEngine_Name_Ugly(testingT *core.T) {
	engine := SyntheticEngine{}
	first := engine.Name()
	second := engine.Name()
	core.AssertEqual(testingT, first, second)
}

func TestAX7_SyntheticEngine_Platforms_Good(testingT *core.T) {
	engine := SyntheticEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "synthetic")
}

func TestAX7_SyntheticEngine_Platforms_Bad(testingT *core.T) {
	engine := SyntheticEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "dos")
}

func TestAX7_SyntheticEngine_Platforms_Ugly(testingT *core.T) {
	engine := SyntheticEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 1)
}

func TestAX7_SyntheticEngine_Verify_Good(testingT *core.T) {
	engine := SyntheticEngine{}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_SyntheticEngine_Verify_Bad(testingT *core.T) {
	engine := SyntheticEngine{}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_SyntheticEngine_Verify_Ugly(testingT *core.T) {
	engine := SyntheticEngine{}
	err := engine.Verify()
	core.AssertNil(testingT, err)
}

func TestAX7_SyntheticEngine_CodeIdentity_Good(testingT *core.T) {
	engine := SyntheticEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "synthetic", identity.Name)
}

func TestAX7_SyntheticEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := SyntheticEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "synthetic", identity.Path)
}

func TestAX7_SyntheticEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := SyntheticEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_SyntheticEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := SyntheticEngine{}.Run("rom/game.bin", EngineConfig{Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "SYNTHETIC ENGINE OK")
}

func TestAX7_SyntheticEngine_Run_Bad(testingT *core.T) {
	err := SyntheticEngine{}.Run("rom/game.bin", EngineConfig{})
	core.AssertNoError(testingT, err)
	core.AssertNil(testingT, err)
}

func TestAX7_SyntheticEngine_Run_Ugly(testingT *core.T) {
	output := core.NewBuffer()
	err := SyntheticEngine{}.Run("", EngineConfig{Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "SYNTHETIC ENGINE OK")
}

func TestAX7_SyntheticEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := shieldBundle(testingT, []byte("rom"))
	plan, err := SyntheticEngine{}.PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "synthetic", plan.Engine)
}

func TestAX7_SyntheticEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := shieldBundle(testingT, []byte("rom"))
	bundle.Manifest.Runtime.Engine = "retroarch"
	_, err := SyntheticEngine{}.PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/runtime-mismatch")
}

func TestAX7_SyntheticEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := shieldBundle(testingT, []byte("rom"))
	bundle.Manifest.Platform = "dos"
	_, err := SyntheticEngine{}.PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/platform-unsupported")
}

func TestAX7_VICEEngine_Name_Good(testingT *core.T) {
	engine := VICEEngine{}
	got := engine.Name()
	core.AssertEqual(testingT, "vice", got)
}

func TestAX7_VICEEngine_Name_Bad(testingT *core.T) {
	engine := VICEEngine{}
	got := engine.Name()
	core.AssertNotEqual(testingT, "fuse", got)
}

func TestAX7_VICEEngine_Name_Ugly(testingT *core.T) {
	engine := VICEEngine{Binary: "x64sc"}
	got := engine.Name()
	core.AssertEqual(testingT, "vice", got)
}

func TestAX7_VICEEngine_Platforms_Good(testingT *core.T) {
	engine := VICEEngine{}
	platforms := engine.Platforms()
	core.AssertContains(testingT, platforms, "commodore-64")
}

func TestAX7_VICEEngine_Platforms_Bad(testingT *core.T) {
	engine := VICEEngine{}
	platforms := engine.Platforms()
	core.AssertNotContains(testingT, platforms, "zx-spectrum")
}

func TestAX7_VICEEngine_Platforms_Ugly(testingT *core.T) {
	engine := VICEEngine{}
	platforms := engine.Platforms()
	core.AssertLen(testingT, platforms, 5)
}

func TestAX7_VICEEngine_Acceleration_Good(testingT *core.T) {
	engine := VICEEngine{}
	acceleration := engine.Acceleration()
	core.AssertEqual(testingT, AccelerationAuto, acceleration.Mode)
}

func TestAX7_VICEEngine_Acceleration_Bad(testingT *core.T) {
	engine := VICEEngine{}
	acceleration := engine.Acceleration()
	core.AssertNotEqual(testingT, AccelerationRequired, acceleration.Mode)
}

func TestAX7_VICEEngine_Acceleration_Ugly(testingT *core.T) {
	engine := VICEEngine{}
	acceleration := engine.Acceleration()
	core.AssertContains(testingT, acceleration.PreferredFilters, FrameFilterCRT)
}

func TestAX7_VICEEngine_Verify_Good(testingT *core.T) {
	engine := VICEEngine{Binary: "x64sc"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_VICEEngine_Verify_Bad(testingT *core.T) {
	engine := VICEEngine{}
	err := engine.Verify()
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_VICEEngine_Verify_Ugly(testingT *core.T) {
	engine := VICEEngine{Binary: "bin/x64sc"}
	err := engine.Verify()
	core.AssertNoError(testingT, err)
}

func TestAX7_VICEEngine_CodeIdentity_Good(testingT *core.T) {
	engine := VICEEngine{Binary: "x64sc", BinarySHA256: validArtefactSHA256}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, validArtefactSHA256, identity.SHA256)
}

func TestAX7_VICEEngine_CodeIdentity_Bad(testingT *core.T) {
	engine := VICEEngine{}
	identity := engine.CodeIdentity()
	core.AssertEqual(testingT, "vice", identity.Path)
}

func TestAX7_VICEEngine_CodeIdentity_Ugly(testingT *core.T) {
	engine := VICEEngine{}
	identity := engine.CodeIdentity()
	core.AssertLen(testingT, identity.SHA256, 64)
}

func TestAX7_VICEEngine_Run_Good(testingT *core.T) {
	output := core.NewBuffer()
	err := (VICEEngine{Binary: "x64sc"}).Run("rom/game.d64", EngineConfig{Core: ax7ProcessCore("ran"), Context: core.Background(), Profile: "c64", Output: output})
	core.AssertNoError(testingT, err)
	core.AssertContains(testingT, output.String(), "ran")
}

func TestAX7_VICEEngine_Run_Bad(testingT *core.T) {
	err := (VICEEngine{}).Run("rom/game.d64", ax7EngineConfig("c64"))
	core.AssertError(testingT, err, "engine/binary-required")
	core.AssertNotNil(testingT, err)
}

func TestAX7_VICEEngine_Run_Ugly(testingT *core.T) {
	err := (VICEEngine{Binary: "x64sc"}).Run("rom/game.d64", EngineConfig{Profile: "unknown"})
	core.AssertError(testingT, err, "engine/profile-unsupported")
	core.AssertNotNil(testingT, err)
}

func TestAX7_VICEEngine_PlanLaunch_Good(testingT *core.T) {
	bundle := ax7Bundle(testingT, viceBundleFS(testingT))
	plan, err := (VICEEngine{Binary: "x64sc"}).PlanLaunch(bundle)
	core.AssertNoError(testingT, err)
	core.AssertEqual(testingT, "vice", plan.Engine)
}

func TestAX7_VICEEngine_PlanLaunch_Bad(testingT *core.T) {
	bundle := ax7Bundle(testingT, viceBundleFS(testingT))
	_, err := (VICEEngine{}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/binary-required")
}

func TestAX7_VICEEngine_PlanLaunch_Ugly(testingT *core.T) {
	bundle := ax7Bundle(testingT, viceBundleFS(testingT))
	bundle.Manifest.Runtime.Profile = "unknown"
	_, err := (VICEEngine{Binary: "x64sc"}).PlanLaunch(bundle)
	core.AssertError(testingT, err, "engine/profile-unsupported")
}
