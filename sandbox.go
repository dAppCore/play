package play

import "path"

// SandboxPolicy describes the host runtime boundaries for a STIM launch.
type SandboxPolicy struct {
	BundleName     string
	Root           string
	Saves          string
	Screenshots    string
	SessionLog     string
	ReadPaths      []string
	WritePaths     []string
	NetworkAllowed bool
}

// ValidateLaunch reports launch plans that widen the prepared sandbox policy.
func (policy SandboxPolicy) ValidateLaunch(plan LaunchPlan) ValidationErrors {
	var issues ValidationErrors
	if plan.NetworkAllowed && !policy.NetworkAllowed {
		issues = append(issues, ValidationIssue{
			Code:    "sandbox/network-denied",
			Field:   "permissions.network",
			Message: "launch plan requests network access outside the sandbox policy",
		})
	}
	if plan.Entrypoint != "" && !sandboxPathAllowed(plan.Entrypoint, policy.ReadPaths) {
		issues = append(issues, ValidationIssue{
			Code:    "sandbox/read-denied",
			Field:   "launch.entrypoint",
			Message: "launch entrypoint is outside the sandbox read allowlist",
		})
	}
	if plan.RuntimeConfig != "" && !sandboxPathAllowed(plan.RuntimeConfig, policy.ReadPaths) {
		issues = append(issues, ValidationIssue{
			Code:    "sandbox/read-denied",
			Field:   "launch.runtime_config",
			Message: "launch runtime config is outside the sandbox read allowlist",
		})
	}

	for _, readPath := range plan.ReadPaths {
		if !sandboxPathAllowed(readPath, policy.ReadPaths) {
			issues = append(issues, ValidationIssue{
				Code:    "sandbox/read-denied",
				Field:   readPath,
				Message: "launch plan requests a read path outside the sandbox policy",
			})
		}
	}
	for _, writePath := range plan.WritePaths {
		if !sandboxPathAllowed(writePath, policy.WritePaths) {
			issues = append(issues, ValidationIssue{
				Code:    "sandbox/write-denied",
				Field:   writePath,
				Message: "launch plan requests a write path outside the sandbox policy",
			})
		}
	}

	return issues
}

// PrepareSandbox creates the save-state layout for a bundle.
func PrepareSandbox(bundle Bundle, home string, writer BundleWriter) (SandboxPolicy, error) {
	if home == "" {
		return SandboxPolicy{}, SandboxError{
			Kind:    "sandbox/home-required",
			Message: "home directory is required",
		}
	}
	if writer == nil {
		return SandboxPolicy{}, SandboxError{
			Kind:    "sandbox/writer-required",
			Message: "sandbox writer is required",
		}
	}
	if bundle.Manifest.Name == "" {
		return SandboxPolicy{}, SandboxError{
			Kind:    "sandbox/name-required",
			Message: "bundle name is required",
		}
	}

	root := path.Join(home, ".core", "play", bundle.Manifest.Name)
	saves := path.Join(root, "saves")
	screenshots := path.Join(root, "screenshots")
	if bundle.Manifest.Save.Path != "" {
		saves = path.Join(root, normaliseBundlePath(bundle.Manifest.Save.Path))
	}
	if bundle.Manifest.Save.Screenshots != "" {
		screenshots = path.Join(root, normaliseBundlePath(bundle.Manifest.Save.Screenshots))
	}

	for _, directory := range []string{root, saves, screenshots} {
		if err := writer.EnsureDirectory(directory); err != nil {
			return SandboxPolicy{}, SandboxError{
				Kind:    "sandbox/directory-create-failed",
				Path:    directory,
				Message: err.Error(),
			}
		}
	}

	return SandboxPolicy{
		BundleName:     bundle.Manifest.Name,
		Root:           root,
		Saves:          saves,
		Screenshots:    screenshots,
		SessionLog:     path.Join(root, "session.log"),
		ReadPaths:      manifestLaunchReadPaths(bundle.Manifest),
		WritePaths:     manifestLaunchWritePaths(bundle.Manifest),
		NetworkAllowed: bundle.Manifest.Permissions.Network,
	}, nil
}

func manifestLaunchReadPaths(manifest Manifest) []string {
	paths := clonePaths(manifest.Permissions.FileSystem.Read)
	paths = appendUniqueBundlePath(paths, manifest.Artefact.Path)
	paths = appendUniqueBundlePath(paths, manifest.Runtime.Config)
	paths = appendUniqueBundlePath(paths, manifest.Runtime.Entrypoint)

	return paths
}

func manifestLaunchWritePaths(manifest Manifest) []string {
	paths := clonePaths(manifest.Permissions.FileSystem.Write)
	paths = appendUniqueBundlePath(paths, manifest.Save.Path)
	paths = appendUniqueBundlePath(paths, manifest.Save.Screenshots)

	return paths
}

func appendUniqueBundlePath(paths []string, value string) []string {
	if !validBundlePath(value) {
		return paths
	}

	cleanValue := normaliseBundlePath(value)
	for _, candidate := range paths {
		if !validBundlePath(candidate) {
			continue
		}
		if normaliseBundlePath(candidate) == cleanValue {
			return paths
		}
	}

	return append(paths, cleanValue)
}

func sandboxPathAllowed(candidate string, allowedPaths []string) bool {
	if !validBundlePath(candidate) {
		return false
	}

	cleanCandidate := normaliseBundlePath(candidate)
	for _, allowedPath := range allowedPaths {
		if !validBundlePath(allowedPath) {
			continue
		}

		cleanAllowed := normaliseBundlePath(allowedPath)
		if cleanCandidate == cleanAllowed || hasBundlePathPrefix(cleanCandidate, cleanAllowed) {
			return true
		}
	}

	return false
}

func hasBundlePathPrefix(candidate string, allowedPath string) bool {
	if len(candidate) <= len(allowedPath) {
		return false
	}
	if candidate[:len(allowedPath)] != allowedPath {
		return false
	}

	return candidate[len(allowedPath)] == '/'
}

// SandboxError reports STIM sandbox preparation failures.
type SandboxError struct {
	Kind    string
	Path    string
	Message string
}

func (sandboxError SandboxError) Error() string {
	if sandboxError.Path == "" {
		return sandboxError.Kind + ": " + sandboxError.Message
	}

	return sandboxError.Kind + ": " + sandboxError.Path + ": " + sandboxError.Message
}
