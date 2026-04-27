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

type checksumRequirement struct {
	Field string
	Path  string
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

	checksumIssues, artefactHash := verifyChecksumChain(bundle.files, bundle.Manifest, entries)
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
		if !validBundlePath(entry.Path) || normaliseBundlePath(entry.Path) != entry.Path {
			return nil, ChecksumParseError{
				Kind:    "checksum/invalid-path",
				Message: "checksum entry path must be a canonical relative bundle path",
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func verifyChecksumChain(filesystem fs.FS, manifest Manifest, entries []ChecksumEntry) (ValidationErrors, string) {
	var issues ValidationErrors
	artefactHash := ""
	recordedPaths := map[string]ChecksumEntry{}

	for _, entry := range entries {
		expectedPath := normaliseBundlePath(entry.Path)
		if _, exists := recordedPaths[expectedPath]; exists {
			issues = append(issues, ValidationIssue{
				Code:    "hash/duplicate-entry",
				Field:   entry.Path,
				Message: "checksum chain contains a duplicate path entry",
			})
		} else {
			recordedPaths[expectedPath] = entry
		}

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
		if expectedPath == normaliseBundlePath(manifest.Artefact.Path) {
			artefactHash = actualHash
		}
	}

	issues = append(issues, verifyRequiredChecksumEntries(manifest, recordedPaths)...)
	issues = append(issues, verifyChecksumCoverage(filesystem, manifest, recordedPaths)...)

	return issues, artefactHash
}

func verifyRequiredChecksumEntries(manifest Manifest, recordedPaths map[string]ChecksumEntry) ValidationErrors {
	var issues ValidationErrors
	seenRequirements := map[string]struct{}{}

	for _, requirement := range requiredChecksumEntries(manifest) {
		if !validBundlePath(requirement.Path) {
			continue
		}

		requiredPath := normaliseBundlePath(requirement.Path)
		if _, seen := seenRequirements[requiredPath]; seen {
			continue
		}
		seenRequirements[requiredPath] = struct{}{}

		if _, recorded := recordedPaths[requiredPath]; !recorded {
			issues = append(issues, ValidationIssue{
				Code:    "hash/chain-entry-missing",
				Field:   requirement.Field,
				Message: "required bundle file is missing from the verification chain",
			})
		}
	}

	return issues
}

func requiredChecksumEntries(manifest Manifest) []checksumRequirement {
	requirements := []checksumRequirement{
		{
			Field: "manifest",
			Path:  "manifest.yaml",
		},
		{
			Field: "runtime.config",
			Path:  manifest.Runtime.Config,
		},
		{
			Field: "verification.sbom",
			Path:  manifest.Verification.SBOM,
		},
		{
			Field: "artefact.path",
			Path:  manifest.Artefact.Path,
		},
	}
	return requirements
}

func verifyChecksumCoverage(filesystem fs.FS, manifest Manifest, recordedPaths map[string]ChecksumEntry) ValidationErrors {
	var issues ValidationErrors
	chainPath := normaliseBundlePath(manifest.Verification.Chain)

	_ = fs.WalkDir(filesystem, ".", func(candidatePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return nil
		}

		cleanPath := normaliseBundlePath(candidatePath)
		if cleanPath == "." || cleanPath == chainPath {
			return nil
		}
		if _, recorded := recordedPaths[cleanPath]; recorded {
			return nil
		}

		issues = append(issues, ValidationIssue{
			Code:    "hash/unrecorded-file",
			Field:   cleanPath,
			Message: "bundle file is not recorded in the verification chain",
		})

		return nil
	})

	return issues
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
