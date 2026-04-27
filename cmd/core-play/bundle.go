package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path"

	"dappco.re/go/core"
	"dappco.re/go/play"
)

func runBundle(parsed invocation, out io.Writer) error {
	artefactData, err := os.ReadFile(parsed.ROM)
	if err != nil {
		return err
	}
	engineBinaryPath := ""
	var engineBinaryData []byte
	if parsed.EngineBin != "" {
		engineData, readErr := os.ReadFile(parsed.EngineBin)
		if readErr != nil {
			return readErr
		}
		engineBinaryPath = path.Join("engine", path.Base(parsed.EngineBin))
		engineBinaryData = engineData
	}

	artefactPath := path.Join("rom", path.Base(parsed.ROM))
	service := play.NewService(nil, nil)
	rendered, err := service.RenderBundle(play.BundleRequest{
		Name:             parsed.Name,
		Title:            parsed.Title,
		Author:           parsed.Author,
		Year:             parsed.Year,
		Platform:         parsed.Platform,
		Genre:            parsed.Genre,
		Licence:          parsed.Licence,
		Engine:           parsed.Engine,
		Profile:          parsed.Profile,
		ArtefactPath:     artefactPath,
		ArtefactData:     artefactData,
		ArtefactSHA256:   hashBytes(artefactData),
		ArtefactSize:     int64(len(artefactData)),
		ArtefactSource:   parsed.Source,
		EngineBinaryPath: engineBinaryPath,
		EngineBinaryData: engineBinaryData,
		ResourceLimits: play.ResourceLimits{
			CPUPercent:  parsed.CPU,
			MemoryBytes: parsed.Memory,
		},
		BYOROM: parsed.BYOROM,
	})
	if err != nil {
		return err
	}

	targetRoot := bundleRoot(parsed)
	if parsed.Archive {
		return writeArchiveBundle(rendered, targetRoot, out)
	}
	if err := rendered.Write(localBundleWriter{Root: targetRoot}); err != nil {
		return err
	}

	core.Print(out, "bundle created: %s", outputPath(targetRoot, rendered.Path))
	return nil
}

func writeArchiveBundle(rendered play.RenderedBundle, targetRoot string, out io.Writer) error {
	archiveData, archiveErr := rendered.Archive()
	if archiveErr != nil {
		return archiveErr
	}

	targetPath := outputPath(targetRoot, rendered.Path+".zip")
	if targetRoot != "" && targetRoot != "." {
		if err := os.MkdirAll(targetRoot, 0755); err != nil {
			return err
		}
	}
	if writeErr := os.WriteFile(targetPath, archiveData, 0644); writeErr != nil {
		return writeErr
	}

	core.Print(out, "bundle archive created: %s", targetPath)
	return nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

type localBundleWriter struct {
	Root string
}

func (writer localBundleWriter) EnsureDirectory(targetPath string) error {
	return os.MkdirAll(writer.target(targetPath), 0755)
}

func (writer localBundleWriter) WriteFile(targetPath string, data []byte) error {
	return os.WriteFile(writer.target(targetPath), data, 0644)
}

func (writer localBundleWriter) target(targetPath string) string {
	return outputPath(writer.Root, targetPath)
}

func outputPath(root string, targetPath string) string {
	if root == "" || root == "." {
		return targetPath
	}

	return path.Join(root, targetPath)
}
