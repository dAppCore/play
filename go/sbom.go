package play

import (
	"bytes"

	core "dappco.re/go"
)

const (
	corePlayToolName    = "core-play"
	corePlayToolVersion = "0.2.0"
	sbomTimestamp       = "1980-01-01T00:00:00Z"
)

// SBOMResult reports whether the bundle's CycloneDX document is present and valid.
type SBOMResult struct {
	Path         string
	Present      bool
	Valid        bool
	SerialNumber string
	Issues       ValidationErrors
}

type cycloneDXBOM struct {
	BOMFormat    string               `json:"bomFormat"`
	SpecVersion  string               `json:"specVersion"`
	SerialNumber string               `json:"serialNumber"`
	Version      int                  `json:"version"`
	Metadata     cycloneDXMetadata    `json:"metadata"`
	Components   []cycloneDXComponent `json:"components"`
}

type cycloneDXMetadata struct {
	Timestamp string          `json:"timestamp"`
	Tools     []cycloneDXTool `json:"tools"`
}

type cycloneDXTool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type cycloneDXComponent struct {
	Type    string          `json:"type"`
	Name    string          `json:"name"`
	Version string          `json:"version,omitempty"`
	Hashes  []cycloneDXHash `json:"hashes,omitempty"`
}

type cycloneDXHash struct {
	Algorithm string `json:"alg"`
	Content   string `json:"content"`
}

// BuildSBOM constructs a deterministic CycloneDX 1.5 SBOM for a bundle manifest.
func BuildSBOM(manifest Manifest) ([]byte, error) {
	bom := cycloneDXBOM{
		BOMFormat:    "CycloneDX",
		SpecVersion:  "1.5",
		SerialNumber: "urn:uuid:" + deterministicUUID(bundleIdentityHash(manifest)),
		Version:      1,
		Metadata: cycloneDXMetadata{
			Timestamp: sbomTimestamp,
			Tools: []cycloneDXTool{
				{
					Name:    corePlayToolName,
					Version: corePlayToolVersion,
				},
			},
		},
		Components: []cycloneDXComponent{
			{
				Type:    "application",
				Name:    manifest.Name,
				Version: manifestVersion(manifest),
				Hashes: []cycloneDXHash{
					{
						Algorithm: "SHA-256",
						Content:   bundleIdentityHash(manifest),
					},
				},
			},
			{
				Type: "data",
				Name: manifest.Artefact.Path,
				Hashes: []cycloneDXHash{
					{
						Algorithm: "SHA-256",
						Content:   manifest.Artefact.SHA256,
					},
				},
			},
		},
	}
	if manifest.Verification.Engine.SHA256 != "" {
		bom.Components = append(bom.Components, cycloneDXComponent{
			Type:    "application",
			Name:    defaultString(manifest.Verification.Engine.Name, manifest.Runtime.Engine),
			Version: "runtime",
			Hashes: []cycloneDXHash{
				{
					Algorithm: "SHA-256",
					Content:   manifest.Verification.Engine.SHA256,
				},
			},
		})
	}

	result := core.JSONMarshal(bom)
	if !result.OK {
		return nil, core.E("play.sbom", "failed to serialise CycloneDX SBOM", resultError(result.Value))
	}

	data := cloneBytes(result.Value.([]byte))
	data = append(data, '\n')

	return data, nil
}

func validateSBOM(data []byte, manifest Manifest, path string) SBOMResult {
	result := SBOMResult{
		Path:    path,
		Present: true,
	}

	var bom cycloneDXBOM
	decoded := core.JSONUnmarshal(data, &bom)
	if !decoded.OK {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/parse-failed",
			Field:   path,
			Message: resultMessage(decoded.Value),
		})
		return result
	}

	result.SerialNumber = bom.SerialNumber
	if bom.BOMFormat != "CycloneDX" {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/format-invalid",
			Field:   path,
			Message: "SBOM must use CycloneDX",
		})
	}
	if bom.SpecVersion != "1.5" {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/spec-version-invalid",
			Field:   path,
			Message: "SBOM must use CycloneDX 1.5",
		})
	}
	if bom.Version != 1 {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/version-invalid",
			Field:   path,
			Message: "SBOM version must be 1",
		})
	}
	if !core.HasPrefix(bom.SerialNumber, "urn:uuid:") {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/serial-invalid",
			Field:   path,
			Message: "SBOM serial number must be a UUID URN",
		})
	}
	if !hasCycloneDXHash(bom.Components, "application", manifest.Name, bundleIdentityHash(manifest)) {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/application-hash-missing",
			Field:   path,
			Message: "SBOM must contain the bundle application hash",
		})
	}
	if !hasCycloneDXHash(bom.Components, "data", manifest.Artefact.Path, manifest.Artefact.SHA256) {
		result.Issues = append(result.Issues, ValidationIssue{
			Code:    "sbom/artefact-hash-missing",
			Field:   path,
			Message: "SBOM must contain the artefact hash",
		})
	}

	result.Valid = !result.Issues.HasIssues()
	return result
}

func bundleIdentityHash(manifest Manifest) string {
	var buffer bytes.Buffer
	buffer.WriteString("name=")
	buffer.WriteString(manifest.Name)
	buffer.WriteString("\nversion=")
	buffer.WriteString(manifestVersion(manifest))
	buffer.WriteString("\nplatform=")
	buffer.WriteString(manifest.Platform)
	buffer.WriteString("\nengine=")
	buffer.WriteString(manifest.Runtime.Engine)
	buffer.WriteString("\nartefact.path=")
	buffer.WriteString(manifest.Artefact.Path)
	buffer.WriteString("\nartefact.sha256=")
	buffer.WriteString(manifest.Artefact.SHA256)
	buffer.WriteString("\nengine.sha256=")
	buffer.WriteString(manifest.Verification.Engine.SHA256)
	buffer.WriteString("\n")

	return sha256Hex(buffer.Bytes())
}

func deterministicUUID(hash string) string {
	if len(hash) < 32 {
		return "00000000-0000-0000-0000-000000000000"
	}

	return hash[0:8] + "-" + hash[8:12] + "-" + hash[12:16] + "-" + hash[16:20] + "-" + hash[20:32]
}

func manifestVersion(manifest Manifest) string {
	if manifest.Version != "" {
		return manifest.Version
	}
	if manifest.Year > 0 {
		return core.Sprint(manifest.Year)
	}

	return "1"
}

func hasCycloneDXHash(components []cycloneDXComponent, componentType string, name string, hash string) bool {
	for _, component := range components {
		if component.Type != componentType || component.Name != name {
			continue
		}
		for _, candidate := range component.Hashes {
			if candidate.Algorithm == "SHA-256" && candidate.Content == hash {
				return true
			}
		}
	}

	return false
}

func resultError(value any) error {
	if err, ok := value.(error); ok {
		return err
	}

	return nil
}
