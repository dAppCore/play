package play

import (
	"archive/zip"
	"bytes"
	"sort"
	"time"
)

var deterministicArchiveTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

// Archive renders a deterministic zip archive for a bundle.
func (rendered RenderedBundle) Archive() ([]byte, error) {
	rootPath, err := cleanWriteRoot(rendered.Path)
	if err != nil {
		return nil, err
	}

	files := cloneRenderedFiles(rendered.Files)
	sort.Slice(files, func(first int, second int) bool {
		return files[first].Path < files[second].Path
	})

	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for _, file := range files {
		if !validBundlePath(file.Path) {
			return nil, WriteError{
				Kind:    "bundle/file-path-invalid",
				Path:    file.Path,
				Message: "rendered file path must be a relative bundle path",
			}
		}

		header := &zip.FileHeader{
			Name:   joinWritePath(rootPath, file.Path),
			Method: zip.Deflate,
		}
		header.SetMode(0644)
		header.SetModTime(deterministicArchiveTime)

		writer, err := archive.CreateHeader(header)
		if err != nil {
			_ = archive.Close()
			return nil, WriteError{
				Kind:    "bundle/archive-file-create-failed",
				Path:    file.Path,
				Message: err.Error(),
			}
		}
		if _, err := writer.Write(file.Data); err != nil {
			_ = archive.Close()
			return nil, WriteError{
				Kind:    "bundle/archive-file-write-failed",
				Path:    file.Path,
				Message: err.Error(),
			}
		}
	}
	if err := archive.Close(); err != nil {
		return nil, WriteError{
			Kind:    "bundle/archive-close-failed",
			Path:    rootPath,
			Message: err.Error(),
		}
	}

	return cloneBytes(buffer.Bytes()), nil
}

func cloneRenderedFiles(files []RenderedFile) []RenderedFile {
	if len(files) == 0 {
		return nil
	}

	cloned := make([]RenderedFile, len(files))
	for index, file := range files {
		cloned[index] = RenderedFile{
			Path: file.Path,
			Data: cloneBytes(file.Data),
		}
	}

	return cloned
}
