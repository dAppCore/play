package play

import (
	"io/fs"
	"path"
)

// Bundle represents a STIM bundle loaded from a filesystem.
type Bundle struct {
	Path     string
	Manifest Manifest

	files fs.FS
}

// LoadBundle loads a STIM bundle from the provided filesystem.
func LoadBundle(filesystem fs.FS, bundlePath string) (Bundle, error) {
	cleanPath, err := cleanBundlePath(bundlePath)
	if err != nil {
		return Bundle{}, err
	}

	bundleFiles := filesystem
	if cleanPath != "." {
		bundleFiles, err = fs.Sub(filesystem, cleanPath)
		if err != nil {
			return Bundle{}, err
		}
	}

	manifestData, err := fs.ReadFile(bundleFiles, "manifest.yaml")
	if err != nil {
		return Bundle{}, err
	}

	manifest, err := LoadManifest(manifestData)
	if err != nil {
		return Bundle{}, err
	}

	return Bundle{
		Path:     cleanPath,
		Manifest: manifest,
		files:    bundleFiles,
	}, nil
}

// Validate reports structural and manifest issues for a loaded bundle.
func (bundle Bundle) Validate() ValidationErrors {
	var issues ValidationErrors
	issues = append(issues, bundle.Manifest.Validate()...)

	requiredFiles := []validationTarget{
		{
			Field:   "runtime.config",
			Code:    "bundle/runtime-config-missing",
			Message: "runtime config file is required",
			Path:    bundle.Manifest.Runtime.Config,
		},
		{
			Field:   "verification.chain",
			Code:    "bundle/verification-chain-missing",
			Message: "verification chain file is required",
			Path:    bundle.Manifest.Verification.Chain,
		},
		{
			Field:   "verification.sbom",
			Code:    "bundle/sbom-missing",
			Message: "SBOM file is required",
			Path:    bundle.Manifest.Verification.SBOM,
		},
		{
			Field:   "artefact.path",
			Code:    "bundle/artefact-missing",
			Message: "artefact file is required",
			Path:    bundle.Manifest.Artefact.Path,
		},
	}

	for _, target := range requiredFiles {
		if !validBundlePath(target.Path) {
			continue
		}
		if _, err := fs.Stat(bundle.files, normaliseBundlePath(target.Path)); err != nil {
			issues = append(issues, ValidationIssue{
				Code:    target.Code,
				Field:   target.Field,
				Message: target.Message,
			})
		}
	}

	return issues
}

type validationTarget struct {
	Field   string
	Code    string
	Message string
	Path    string
}

func cleanBundlePath(bundlePath string) (string, error) {
	if bundlePath == "" || bundlePath == "." || bundlePath == "./" {
		return ".", nil
	}
	if len(bundlePath) > 2 && bundlePath[0] == '.' && bundlePath[1] == '/' {
		bundlePath = bundlePath[2:]
	}

	cleanPath, ok := canonicalBundlePath(bundlePath)
	if !ok {
		return "", PathError{
			Kind:    "bundle/path-invalid",
			Path:    bundlePath,
			Message: "bundle path must stay within the provided filesystem",
		}
	}

	return cleanPath, nil
}

func normaliseBundlePath(value string) string {
	return path.Clean(value)
}

func bundleFileData(bundle Bundle, filePath string) ([]byte, bool) {
	if bundle.files == nil || !validBundlePath(filePath) {
		return nil, false
	}

	data, err := fs.ReadFile(bundle.files, normaliseBundlePath(filePath))
	if err != nil {
		return nil, false
	}

	return data, true
}

// PathError reports invalid bundle or manifest paths.
type PathError struct {
	Kind    string
	Path    string
	Message string
}

func (pathError PathError) Error() string {
	return pathError.Kind + ": " + pathError.Path + ": " + pathError.Message
}
