package play

import (
	"io/fs"
	"path"
	"sort"
)

// Service provides CLI-facing bundle operations without binding to a concrete shell.
type Service struct {
	Bundles  fs.FS
	Registry *Registry
}

// BundleSummary describes a bundle for list and verify surfaces.
type BundleSummary struct {
	Path     string
	Name     string
	Title    string
	Platform string
	Engine   string
}

// VerifyResult describes bundle verification output.
type VerifyResult struct {
	Summary  BundleSummary
	Issues   ValidationErrors
	Verified bool
}

// PlayPlan describes a prepared bundle launch.
type PlayPlan struct {
	Summary  BundleSummary
	Manifest Manifest
	Engine   Engine
	Launch   *LaunchPlan
	Issues   ValidationErrors
	Ready    bool
}

// BundlePlan describes a planned bundle before any files are written.
type BundlePlan struct {
	Path         string
	Manifest     Manifest
	ArtefactData []byte
}

// NewService creates a service from a bundle filesystem and engine registry.
func NewService(bundles fs.FS, registry *Registry) Service {
	if registry == nil {
		registry = defaultRegistry
	}

	return Service{
		Bundles:  bundles,
		Registry: registry,
	}
}

// ListBundles discovers bundles beneath a root directory.
func (service Service) ListBundles(request ListRequest) ([]BundleSummary, error) {
	rootPath, err := cleanBundlePath(request.Root)
	if err != nil {
		return nil, err
	}
	if service.Bundles == nil {
		return nil, PathError{
			Kind:    "bundle/filesystem-missing",
			Path:    rootPath,
			Message: "bundle filesystem is required",
		}
	}

	summaries := make([]BundleSummary, 0)
	if hasManifest(service.Bundles, rootPath) {
		bundle, loadErr := LoadBundle(service.Bundles, rootPath)
		if loadErr != nil {
			return nil, loadErr
		}
		summaries = append(summaries, summariseBundle(bundle))
	}

	entries, err := fs.ReadDir(service.Bundles, rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidatePath := joinBundlePath(rootPath, entry.Name())
		if !hasManifest(service.Bundles, candidatePath) {
			continue
		}

		bundle, loadErr := LoadBundle(service.Bundles, candidatePath)
		if loadErr != nil {
			return nil, loadErr
		}
		summaries = append(summaries, summariseBundle(bundle))
	}

	sort.Slice(summaries, func(first int, second int) bool {
		return summaries[first].Path < summaries[second].Path
	})

	return summaries, nil
}

// VerifyBundle loads and verifies a bundle.
func (service Service) VerifyBundle(request VerifyRequest) (VerifyResult, error) {
	if service.Bundles == nil {
		return VerifyResult{}, PathError{
			Kind:    "bundle/filesystem-missing",
			Path:    request.BundlePath,
			Message: "bundle filesystem is required",
		}
	}

	bundle, err := LoadBundle(service.Bundles, request.BundlePath)
	if err != nil {
		return VerifyResult{}, err
	}

	issues := bundle.VerifyWithRegistry(service.registry())

	return VerifyResult{
		Summary:  summariseBundle(bundle),
		Issues:   issues,
		Verified: !issues.HasIssues(),
	}, nil
}

// PreparePlay loads a bundle and resolves its engine for launch.
func (service Service) PreparePlay(request PlayRequest) (PlayPlan, error) {
	if service.Bundles == nil {
		return PlayPlan{}, PathError{
			Kind:    "bundle/filesystem-missing",
			Path:    request.BundlePath,
			Message: "bundle filesystem is required",
		}
	}

	bundle, err := LoadBundle(service.Bundles, request.BundlePath)
	if err != nil {
		return PlayPlan{}, err
	}

	registry := service.registry()
	issues := bundle.VerifyWithRegistry(registry)
	engine, found := registry.Resolve(bundle.Manifest.Runtime.Engine)
	var launch *LaunchPlan
	if found && !issues.HasIssues() {
		launchPlan, launchErr := PlanLaunch(bundle, engine)
		if launchErr == nil {
			launch = &launchPlan
		}
	}

	return PlayPlan{
		Summary:  summariseBundle(bundle),
		Manifest: bundle.Manifest,
		Engine:   engine,
		Launch:   launch,
		Issues:   issues,
		Ready:    found && !issues.HasIssues(),
	}, nil
}

// PlanBundle builds a first-pass manifest plan for bundle creation.
func (service Service) PlanBundle(request BundleRequest) (BundlePlan, ValidationErrors) {
	runtimeDefaults := inferRuntimeDefaults(request)
	runtimeConfigPath := defaultString(request.RuntimeConfigPath, "emulator.yaml")
	verificationChainPath := defaultString(request.VerificationChain, "checksums.sha256")
	sbomPath := defaultString(request.SBOMPath, "sbom.json")
	savePath := defaultString(request.SavePath, "saves/")
	screenshotPath := defaultString(request.ScreenshotPath, "screenshots/")
	distributionMode := defaultString(request.DistributionMode, "catalogue")

	manifest := Manifest{
		Name:     request.Name,
		Title:    request.Title,
		Author:   request.Author,
		Year:     request.Year,
		Platform: request.Platform,
		Genre:    request.Genre,
		Licence:  request.Licence,
		Artefact: Artefact{
			Path:      request.ArtefactPath,
			SHA256:    request.ArtefactSHA256,
			Size:      request.ArtefactSize,
			MediaType: request.ArtefactMediaType,
			Source:    request.ArtefactSource,
		},
		Runtime: Runtime{
			Engine:       runtimeDefaults.Engine,
			Profile:      runtimeDefaults.Profile,
			Config:       runtimeConfigPath,
			Entrypoint:   request.Entrypoint,
			Acceleration: runtimeDefaults.Acceleration,
			Filter:       runtimeDefaults.Filter,
		},
		Verification: Verification{
			Chain:         verificationChainPath,
			SBOM:          sbomPath,
			Deterministic: true,
		},
		Preservation: Preservation{
			Verified: true,
			Chain:    verificationChainPath,
		},
		Permissions: Permissions{
			FileSystem: FileSystemPermissions{
				Read: []string{
					"rom/",
				},
				Write: []string{
					savePath,
					screenshotPath,
				},
			},
		},
		Save: Save{
			Path:        savePath,
			Screenshots: screenshotPath,
		},
		Distribution: Distribution{
			Mode:   distributionMode,
			BYOROM: request.BYOROM,
		},
	}

	issues := manifest.Validate()
	if issues.HasIssues() {
		return BundlePlan{}, issues
	}

	return BundlePlan{
		Path:         request.Name,
		Manifest:     manifest,
		ArtefactData: cloneBytes(request.ArtefactData),
	}, nil
}

// RenderBundle plans and renders bundle files without writing them to disk.
func (service Service) RenderBundle(request BundleRequest) (RenderedBundle, error) {
	plan, issues := service.PlanBundle(request)
	if issues.HasIssues() {
		return RenderedBundle{}, issues
	}

	return plan.Render()
}

// WriteBundle plans, renders, and materialises a bundle through the provided writer.
func (service Service) WriteBundle(request BundleRequest, writer BundleWriter) error {
	rendered, err := service.RenderBundle(request)
	if err != nil {
		return err
	}

	return rendered.Write(writer)
}

func hasManifest(filesystem fs.FS, bundlePath string) bool {
	bundleFiles, err := bundleSubFS(filesystem, bundlePath)
	if err != nil {
		return false
	}

	info, err := fs.Stat(bundleFiles, "manifest.yaml")
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func bundleSubFS(filesystem fs.FS, bundlePath string) (fs.FS, error) {
	if filesystem == nil {
		return nil, PathError{
			Kind:    "bundle/filesystem-missing",
			Path:    bundlePath,
			Message: "bundle filesystem is required",
		}
	}
	if bundlePath == "." {
		return filesystem, nil
	}

	return fs.Sub(filesystem, bundlePath)
}

func summariseBundle(bundle Bundle) BundleSummary {
	return BundleSummary{
		Path:     bundle.Path,
		Name:     bundle.Manifest.Name,
		Title:    bundle.Manifest.Title,
		Platform: bundle.Manifest.Platform,
		Engine:   bundle.Manifest.Runtime.Engine,
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func joinBundlePath(rootPath string, child string) string {
	if rootPath == "." {
		return child
	}

	return path.Join(rootPath, child)
}

func (service Service) registry() *Registry {
	if service.Registry == nil {
		return defaultRegistry
	}

	return service.Registry
}
