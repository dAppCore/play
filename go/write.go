package play

import "path"

// BundleWriter materialises rendered bundles without binding to a concrete host filesystem.
// EnsureDirectory should behave like mkdir -p for the provided relative path.
type BundleWriter interface {
	EnsureDirectory(path string) error
	WriteFile(path string, data []byte) error
}

// Write materialises a rendered bundle through the provided writer.
func (rendered RenderedBundle) Write(writer BundleWriter) error {
	if writer == nil {
		return WriteError{
			Kind:    "bundle/writer-missing",
			Path:    rendered.Path,
			Message: "bundle writer is required",
		}
	}

	rootPath, err := cleanWriteRoot(rendered.Path)
	if err != nil {
		return err
	}
	if rootPath != "." {
		if err := writer.EnsureDirectory(rootPath); err != nil {
			return WriteError{
				Kind:    "bundle/root-create-failed",
				Path:    rootPath,
				Message: err.Error(),
			}
		}
	}

	createdDirectories := map[string]struct{}{}
	if rootPath != "." {
		createdDirectories[rootPath] = struct{}{}
	}

	for _, file := range rendered.Files {
		if !validBundlePath(file.Path) {
			return WriteError{
				Kind:    "bundle/file-path-invalid",
				Path:    file.Path,
				Message: "rendered file path must be a relative bundle path",
			}
		}

		fullPath := joinWritePath(rootPath, file.Path)
		dirPath := path.Dir(fullPath)
		if dirPath != "." {
			if _, exists := createdDirectories[dirPath]; !exists {
				if err := writer.EnsureDirectory(dirPath); err != nil {
					return WriteError{
						Kind:    "bundle/directory-create-failed",
						Path:    dirPath,
						Message: err.Error(),
					}
				}
				createdDirectories[dirPath] = struct{}{}
			}
		}

		if err := writer.WriteFile(fullPath, cloneBytes(file.Data)); err != nil {
			return WriteError{
				Kind:    "bundle/file-write-failed",
				Path:    fullPath,
				Message: err.Error(),
			}
		}
	}

	return nil
}

func cleanWriteRoot(rootPath string) (string, error) {
	if rootPath == "" {
		return "", WriteError{
			Kind:    "bundle/path-required",
			Message: "bundle path is required",
		}
	}

	return cleanBundlePath(rootPath)
}

func joinWritePath(rootPath string, filePath string) string {
	if rootPath == "." {
		return filePath
	}

	return path.Join(rootPath, filePath)
}

func cloneBytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	cloned := make([]byte, len(data))
	copy(cloned, data)

	return cloned
}

// WriteError reports bundle materialisation failures.
type WriteError struct {
	Kind    string
	Path    string
	Message string
}

func (writeError WriteError) Error() string {
	if writeError.Path == "" {
		return writeError.Kind + ": " + writeError.Message
	}

	return writeError.Kind + ": " + writeError.Path + ": " + writeError.Message
}
