package play

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"path"

	"dappco.re/go/core"
)

const maxZipExpansionRatio = 100

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
	if zipExpansionRatio(file) > maxZipExpansionRatio {
		findings = append(findings, ThreatFinding{
			Class:   "zip-expansion",
			Path:    name,
			Message: "ZIP entry has an oversized uncompressed-to-compressed ratio",
		})
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

func zipExpansionRatio(file *zip.File) uint64 {
	compressed := file.CompressedSize64
	if compressed == 0 {
		compressed = 1
	}

	return file.UncompressedSize64 / compressed
}
