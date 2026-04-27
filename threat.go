package play

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"path"

	"dappco.re/go/core"
)

const (
	maxZipExpansionRatio           uint64 = 100
	maxZIPEntryDepth                      = 16
	maxZIPEntryUncompressedBytes   uint64 = 512 * 1024 * 1024
	maxZIPArchiveUncompressedBytes uint64 = 4 * 1024 * 1024 * 1024
)

// ThreatFinding describes a suspicious artefact entry found during scanning.
type ThreatFinding struct {
	Class   string
	Path    string
	Message string
}

// ThreatResult reports whether the artefact contains suspicious payloads.
type ThreatResult struct {
	Path     string
	Scanned  bool
	Findings []ThreatFinding
	OK       bool
	Issues   ValidationErrors
}

func verifyThreat(bundle Bundle) ThreatResult {
	result := ThreatResult{
		Path: bundle.Manifest.Artefact.Path,
	}

	data, ok := bundleFileData(bundle, bundle.Manifest.Artefact.Path)
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "threat/artefact-missing",
			Field:   "artefact.path",
			Message: "artefact file is required for threat scanning",
		})
		result.OK = false
		return result
	}
	if !looksLikeZIP(data) {
		result.Scanned = true
		result.OK = true
		return result
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "threat/zip-invalid",
			Field:   "artefact.path",
			Message: err.Error(),
		})
		result.OK = false
		return result
	}

	result.Scanned = true
	result.Findings = append(result.Findings, scanZIPArchive(bundle.Manifest.Artefact.Path, reader.File)...)
	for _, file := range reader.File {
		result.Findings = append(result.Findings, scanZIPEntry(file)...)
	}
	for _, finding := range result.Findings {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "threat/" + finding.Class,
			Field:   finding.Path,
			Message: finding.Message,
		})
	}

	result.OK = len(result.Findings) == 0 && !result.Issues.HasIssues()
	return result
}

func scanZIPEntry(file *zip.File) []ThreatFinding {
	var findings []ThreatFinding
	name := file.Name
	if file.FileInfo().IsDir() {
		return findings
	}
	if !validArchiveEntryPath(name) {
		findings = append(findings, ThreatFinding{
			Class:   "path-invalid",
			Path:    name,
			Message: "ZIP entry path escapes the artefact root",
		})
	}
	if file.FileInfo().Mode()&0111 != 0 {
		findings = append(findings, ThreatFinding{
			Class:   "executable-entry",
			Path:    name,
			Message: "ZIP entry has executable bits set",
		})
	}
	if suspiciousScriptPath(name) {
		findings = append(findings, ThreatFinding{
			Class:   "embedded-script",
			Path:    name,
			Message: "ZIP entry contains an embedded script or executable payload",
		})
	}
	if archivePathDepth(name) > maxZIPEntryDepth {
		findings = append(findings, ThreatFinding{
			Class:   "path-depth",
			Path:    name,
			Message: "ZIP entry is nested too deeply",
		})
	}
	if file.UncompressedSize64 > maxZIPEntryUncompressedBytes {
		findings = append(findings, ThreatFinding{
			Class:   "entry-size",
			Path:    name,
			Message: "ZIP entry is too large to scan safely",
		})
	}
	if zipExpansionRatio(file) > maxZipExpansionRatio {
		findings = append(findings, ThreatFinding{
			Class:   "zip-expansion",
			Path:    name,
			Message: "ZIP entry has an oversized uncompressed-to-compressed ratio",
		})
	}

	return findings
}

func scanZIPArchive(artefactPath string, files []*zip.File) []ThreatFinding {
	var findings []ThreatFinding
	var total uint64
	totalExceeded := false

	for _, file := range files {
		if file.FileInfo().IsDir() || totalExceeded {
			continue
		}

		size := file.UncompressedSize64
		if size > maxZIPArchiveUncompressedBytes {
			findings = append(findings, ThreatFinding{
				Class:   "archive-size",
				Path:    artefactPath,
				Message: "ZIP artefact expands beyond the safe scan limit",
			})
			totalExceeded = true
			continue
		}
		if total > maxZIPArchiveUncompressedBytes-size || total+size > maxZIPArchiveUncompressedBytes {
			findings = append(findings, ThreatFinding{
				Class:   "archive-size",
				Path:    artefactPath,
				Message: "ZIP artefact expands beyond the safe scan limit",
			})
			totalExceeded = true
			continue
		}

		total += size
	}

	return findings
}

func looksLikeZIP(data []byte) bool {
	return len(data) >= 4 && data[0] == 'P' && data[1] == 'K' && data[2] == 3 && data[3] == 4
}

func validArchiveEntryPath(name string) bool {
	cleanName := path.Clean(name)
	if cleanName == "." || cleanName != name {
		return false
	}
	if core.HasPrefix(cleanName, "/") || core.HasPrefix(cleanName, "../") || core.Contains(cleanName, "\\") {
		return false
	}

	return fs.ValidPath(cleanName)
}

func suspiciousScriptPath(name string) bool {
	lower := core.Lower(name)
	extension := path.Ext(lower)
	switch extension {
	case ".bat", ".cmd", ".com", ".exe", ".jar", ".js", ".lua", ".pl", ".ps1", ".py", ".rb", ".sh", ".vbs":
		return true
	default:
		return false
	}
}

func archivePathDepth(name string) int {
	if name == "" {
		return 0
	}

	depth := 1
	for index := 0; index < len(name); index++ {
		if name[index] == '/' {
			depth++
		}
	}

	return depth
}

func zipExpansionRatio(file *zip.File) uint64 {
	compressed := file.CompressedSize64
	if compressed == 0 {
		compressed = 1
	}

	return file.UncompressedSize64 / compressed
}
