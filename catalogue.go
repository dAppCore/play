package play

import (
	"io"
	"io/fs"
	"path"
	"sort"

	"dappco.re/go/core"
)

// BundleSummary describes a bundle for list and verify surfaces.
type BundleSummary struct {
	Name     string `json:"name"`
	Title    string `json:"title,omitempty"`
	Platform string `json:"platform"`
	Engine   string `json:"engine"`
	Size     int64  `json:"size"`
	Year     int    `json:"year,omitempty"`
	Verified bool   `json:"verified"`
	Path     string `json:"path"`
}

// Catalogue indexes multiple STIM bundles under one filesystem root.
type Catalogue struct {
	Bundles  fs.FS
	Registry *Registry
	BasePath string
}

// Walk discovers all bundles under rootDir and returns stable summaries.
func (catalogue Catalogue) Walk(rootDir string) ([]BundleSummary, error) {
	rootPath, err := cleanBundlePath(defaultString(rootDir, "."))
	if err != nil {
		return nil, err
	}
	if catalogue.Bundles == nil {
		return nil, PathError{
			Kind:    "bundle/filesystem-missing",
			Path:    rootPath,
			Message: "bundle filesystem is required",
		}
	}

	summaries := make([]BundleSummary, 0)
	if hasManifest(catalogue.Bundles, rootPath) {
		summary, loadErr := catalogue.summary(rootPath)
		if loadErr != nil {
			return nil, loadErr
		}
		summaries = append(summaries, summary)
	}

	err = fs.WalkDir(catalogue.Bundles, rootPath, func(candidatePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() || candidatePath == rootPath {
			return nil
		}
		if !hasManifest(catalogue.Bundles, candidatePath) {
			return nil
		}

		summary, loadErr := catalogue.summary(candidatePath)
		if loadErr != nil {
			return loadErr
		}
		summaries = append(summaries, summary)

		return fs.SkipDir
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(summaries, func(first int, second int) bool {
		return summaries[first].Name < summaries[second].Name
	})

	return summaries, nil
}

// Print writes a deterministic table for bundle summaries.
func (catalogue Catalogue) Print(writer io.Writer, summaries []BundleSummary) error {
	widths := catalogueWidths(summaries)
	core.Print(writer, "%s  %s  %s  %s  %s  %s",
		padRight("NAME", widths.Name),
		padRight("PLATFORM", widths.Platform),
		padRight("ENGINE", widths.Engine),
		padRight("YEAR", widths.Year),
		padLeft("SIZE", widths.Size),
		"VERIFIED",
	)
	for _, summary := range summaries {
		core.Print(writer, "%s  %s  %s  %s  %s  %s",
			padRight(summary.Name, widths.Name),
			padRight(summary.Platform, widths.Platform),
			padRight(summary.Engine, widths.Engine),
			padRight(yearText(summary.Year), widths.Year),
			padLeft(core.Sprint(summary.Size), widths.Size),
			verifiedMarker(summary.Verified),
		)
	}

	return nil
}

// PrintJSON writes machine-readable bundle summaries.
func (catalogue Catalogue) PrintJSON(writer io.Writer, summaries []BundleSummary) error {
	result := core.JSONMarshal(summaries)
	if !result.OK {
		return core.E("play.catalogue", "failed to serialise catalogue", resultError(result.Value))
	}
	core.Print(writer, "%s", string(result.Value.([]byte)))

	return nil
}

func (catalogue Catalogue) summary(bundlePath string) (BundleSummary, error) {
	bundle, err := LoadBundle(catalogue.Bundles, bundlePath)
	if err != nil {
		return BundleSummary{}, err
	}

	summary := summariseBundle(bundle)
	summary.Size = catalogue.bundleSize(bundlePath)
	summary.Path = catalogue.summaryPath(bundlePath)
	summary.Verified = (Shield{Registry: catalogue.registry()}).Verify(bundle).OverallOK

	return summary, nil
}

func (catalogue Catalogue) registry() *Registry {
	if catalogue.Registry == nil {
		return defaultRegistry
	}

	return catalogue.Registry
}

func (catalogue Catalogue) bundleSize(bundlePath string) int64 {
	var size int64
	_ = fs.WalkDir(catalogue.Bundles, bundlePath, func(_ string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		size += info.Size()

		return nil
	})

	return size
}

func (catalogue Catalogue) summaryPath(bundlePath string) string {
	if catalogue.BasePath == "" || catalogue.BasePath == "." {
		return bundlePath
	}
	if bundlePath == "." {
		return catalogue.BasePath
	}

	return path.Join(catalogue.BasePath, bundlePath)
}

type catalogueColumnWidths struct {
	Name     int
	Platform int
	Engine   int
	Year     int
	Size     int
}

func catalogueWidths(summaries []BundleSummary) catalogueColumnWidths {
	widths := catalogueColumnWidths{
		Name:     len("NAME"),
		Platform: len("PLATFORM"),
		Engine:   len("ENGINE"),
		Year:     len("YEAR"),
		Size:     len("SIZE"),
	}
	for _, summary := range summaries {
		widths.Name = maxInt(widths.Name, len(summary.Name))
		widths.Platform = maxInt(widths.Platform, len(summary.Platform))
		widths.Engine = maxInt(widths.Engine, len(summary.Engine))
		widths.Year = maxInt(widths.Year, len(yearText(summary.Year)))
		widths.Size = maxInt(widths.Size, len(core.Sprint(summary.Size)))
	}

	return widths
}

func padRight(value string, width int) string {
	for len(value) < width {
		value += " "
	}

	return value
}

func padLeft(value string, width int) string {
	for len(value) < width {
		value = " " + value
	}

	return value
}

func yearText(year int) string {
	if year == 0 {
		return "-"
	}

	return core.Sprint(year)
}

func verifiedMarker(verified bool) string {
	if verified {
		return "[Y]"
	}

	return "[N]"
}

func maxInt(first int, second int) int {
	if first > second {
		return first
	}

	return second
}
