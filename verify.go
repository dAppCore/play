package play

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
)

// ChecksumEntry describes a single file hash in a checksum chain.
type ChecksumEntry struct {
	Path   string
	SHA256 string
}

// Verify checks a bundle's checksum chain, artefact hash, and engine availability.
func (bundle Bundle) Verify() ValidationErrors {
	return bundle.VerifyWithRegistry(defaultRegistry)
}

// VerifyWithRegistry checks a bundle using the provided engine registry.
func (bundle Bundle) VerifyWithRegistry(registry *Registry) ValidationErrors {
	issues := bundle.Validate()
	if issues.HasIssues() {
		return issues
	}

	checksumData, err := fs.ReadFile(bundle.files, normaliseBundlePath(bundle.Manifest.Verification.Chain))
	if err != nil {
		return append(issues, ValidationIssue{
			Code:    "bundle/verification-chain-missing",
			Field:   "verification.chain",
			Message: "verification chain file is required",
		})
	}

	entries, err := ParseChecksumFile(checksumData)
	if err != nil {
		return append(issues, ValidationIssue{
			Code:    "hash/parse-failed",
			Field:   "verification.chain",
			Message: err.Error(),
		})
	}

	checksumIssues, artefactHash := verifyChecksumEntries(bundle.files, entries, bundle.Manifest.Artefact.Path)
	issues = append(issues, checksumIssues...)

	if artefactHash == "" {
		issues = append(issues, ValidationIssue{
			Code:    "hash/artefact-missing",
			Field:   "artefact.path",
			Message: "artefact hash entry is missing from the verification chain",
		})
	} else if artefactHash != bundle.Manifest.Artefact.SHA256 {
		issues = append(issues, ValidationIssue{
			Code:    "hash/artefact-mismatch",
			Field:   "artefact.sha256",
			Message: "artefact hash does not match the manifest",
		})
	}

	if registry == nil {
		return append(issues, ValidationIssue{
			Code:    "engine/registry-missing",
			Field:   "runtime.engine",
			Message: "engine registry is required",
		})
	}

	engine, found := registry.Resolve(bundle.Manifest.Runtime.Engine)
	if !found {
		return append(issues, ValidationIssue{
			Code:    "engine/unavailable",
			Field:   "runtime.engine",
			Message: "runtime engine is not registered",
		})
	}

	if err := engine.Verify(); err != nil {
		return append(issues, ValidationIssue{
			Code:    "engine/verify-failed",
			Field:   "runtime.engine",
			Message: err.Error(),
		})
	}

	return issues
}

// ParseChecksumFile parses sha256sum-style checksum content.
func ParseChecksumFile(data []byte) ([]ChecksumEntry, error) {
	lines := bytes.Split(data, []byte("\n"))
	entries := make([]ChecksumEntry, 0, len(lines))

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}

		fields := bytes.Fields(trimmed)
		if len(fields) < 2 {
			return nil, ChecksumParseError{
				Kind:    "checksum/invalid-line",
				Message: "checksum lines must contain a sha256 value and path",
			}
		}

		entry := ChecksumEntry{
			SHA256: string(fields[0]),
			Path:   string(fields[1]),
		}
		if !validSHA256(entry.SHA256) {
			return nil, ChecksumParseError{
				Kind:    "checksum/invalid-hash",
				Message: "checksum entry must contain a valid sha256 value",
			}
		}
		if !validBundlePath(entry.Path) {
			return nil, ChecksumParseError{
				Kind:    "checksum/invalid-path",
				Message: "checksum entry path must be a relative bundle path",
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func verifyChecksumEntries(filesystem fs.FS, entries []ChecksumEntry, artefactPath string) (ValidationErrors, string) {
	var issues ValidationErrors
	artefactHash := ""

	for _, entry := range entries {
		expectedPath := normaliseBundlePath(entry.Path)
		data, err := fs.ReadFile(filesystem, expectedPath)
		if err != nil {
			issues = append(issues, ValidationIssue{
				Code:    "hash/file-missing",
				Field:   entry.Path,
				Message: "checksum entry points to a missing file",
			})
			continue
		}

		actualHash := sha256Hex(data)
		if actualHash != entry.SHA256 {
			issues = append(issues, ValidationIssue{
				Code:    "hash/mismatch",
				Field:   entry.Path,
				Message: "file hash does not match the verification chain",
			})
		}
		if expectedPath == normaliseBundlePath(artefactPath) {
			artefactHash = actualHash
		}
	}

	return issues, artefactHash
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ChecksumParseError reports invalid checksum files.
type ChecksumParseError struct {
	Kind    string
	Message string
}

func (checksumParseError ChecksumParseError) Error() string {
	return checksumParseError.Kind + ": " + checksumParseError.Message
}
