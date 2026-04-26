package play

import "io/fs"

// ContentResult reports hash-chain and artefact integrity status.
type ContentResult struct {
	ChainPath      string
	ArtefactPath   string
	ArtefactSHA256 string
	HashChainOK    bool
	ArtefactHashOK bool
	OK             bool
	Issues         ValidationErrors
}

// ShieldReport aggregates all Shield surfaces for a STIM bundle.
type ShieldReport struct {
	SBOM      SBOMResult
	Code      CodeResult
	Content   ContentResult
	Threat    ThreatResult
	OverallOK bool
}

// Shield verifies SBOM, code, content, and threat surfaces.
type Shield struct {
	Registry *Registry
}

// Verify checks every Shield surface for the supplied bundle.
func (shield Shield) Verify(bundle Bundle) ShieldReport {
	registry := shield.Registry
	if registry == nil {
		registry = defaultRegistry
	}

	sbom := verifySBOM(bundle)
	code := verifyCode(bundle, registry)
	content := verifyContent(bundle)
	threat := verifyThreat(bundle)

	return ShieldReport{
		SBOM:      sbom,
		Code:      code,
		Content:   content,
		Threat:    threat,
		OverallOK: sbom.Valid && code.OK && content.OK && threat.OK,
	}
}

// Issues returns all Shield validation failures in surface order.
func (report ShieldReport) Issues() ValidationErrors {
	var issues ValidationErrors
	issues = append(issues, report.SBOM.Issues...)
	issues = append(issues, report.Code.Issues...)
	issues = append(issues, report.Content.Issues...)
	issues = append(issues, report.Threat.Issues...)

	return issues
}

func verifySBOM(bundle Bundle) SBOMResult {
	sbomPath := normaliseBundlePath(bundle.Manifest.Verification.SBOM)
	result := SBOMResult{
		Path: sbomPath,
	}

	data, ok := bundleFileData(bundle, sbomPath)
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/missing",
			Field:   "verification.sbom",
			Message: "SBOM file is required",
		})
		return result
	}

	return validateSBOM(data, bundle.Manifest, sbomPath)
}

func verifyContent(bundle Bundle) ContentResult {
	result := ContentResult{
		ChainPath:      bundle.Manifest.Verification.Chain,
		ArtefactPath:   bundle.Manifest.Artefact.Path,
		ArtefactSHA256: bundle.Manifest.Artefact.SHA256,
	}
	issues := bundle.Validate()
	if issues.HasIssues() {
		result.Issues = append(result.Issues, issues...)
		result.OK = false
		return result
	}

	checksumData, err := fs.ReadFile(bundle.files, normaliseBundlePath(bundle.Manifest.Verification.Chain))
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "content/verification-chain-missing",
			Field:   "verification.chain",
			Message: "verification chain file is required",
		})
		result.OK = false
		return result
	}

	entries, err := ParseChecksumFile(checksumData)
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "content/hash-parse-failed",
			Field:   "verification.chain",
			Message: err.Error(),
		})
		result.OK = false
		return result
	}

	checksumIssues, artefactHash := verifyChecksumEntries(bundle.files, entries, bundle.Manifest.Artefact.Path)
	result.Issues = append(result.Issues, checksumIssues...)
	result.HashChainOK = !checksumIssues.HasIssues()

	if artefactHash == "" {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "content/artefact-hash-missing",
			Field:   "artefact.path",
			Message: "artefact hash entry is missing from the verification chain",
		})
	} else if artefactHash != bundle.Manifest.Artefact.SHA256 {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "content/artefact-mismatch",
			Field:   "artefact.sha256",
			Message: "artefact hash does not match the manifest",
		})
	} else {
		result.ArtefactHashOK = true
	}

	result.OK = result.HashChainOK && result.ArtefactHashOK && !result.Issues.HasIssues()
	return result
}
