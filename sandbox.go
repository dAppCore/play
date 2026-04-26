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
		ReadPaths:      clonePaths(bundle.Manifest.Permissions.FileSystem.Read),
		WritePaths:     clonePaths(bundle.Manifest.Permissions.FileSystem.Write),
		NetworkAllowed: bundle.Manifest.Permissions.Network,
	}, nil
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
