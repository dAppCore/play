package play

import (
	"bytes"
	"encoding/hex"
	"io/fs"
	"path"
)

// ValidationIssue describes a specific validation failure.
type ValidationIssue struct {
	Code    string
	Field   string
	Message string
}

func (issue ValidationIssue) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString(issue.Code)
	if issue.Field != "" {
		buffer.WriteString(": ")
		buffer.WriteString(issue.Field)
	}
	if issue.Message != "" {
		buffer.WriteString(": ")
		buffer.WriteString(issue.Message)
	}

	return buffer.String()
}

// ValidationErrors is a list of validation issues.
type ValidationErrors []ValidationIssue

func (issues ValidationErrors) Error() string {
	var buffer bytes.Buffer
	for index, issue := range issues {
		if index > 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(issue.Error())
	}

	return buffer.String()
}

// HasIssues reports whether validation produced any failures.
func (issues ValidationErrors) HasIssues() bool {
	return len(issues) > 0
}

// Validate reports manifest issues that block bundle execution.
func (manifest Manifest) Validate() ValidationErrors {
	var issues ValidationErrors

	if manifest.Name == "" {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/name-required",
			Field:   "name",
			Message: "name is required",
		})
	}
	if manifest.Title == "" {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/title-required",
			Field:   "title",
			Message: "title is required",
		})
	}
	if manifest.Platform == "" {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/platform-required",
			Field:   "platform",
			Message: "platform is required",
		})
	}
	if manifest.Licence == "" {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/licence-required",
			Field:   "licence",
			Message: "licence is required",
		})
	}

	issues = append(issues, validatePathField("artefact.path", manifest.Artefact.Path, "manifest/artefact-path-invalid", "artefact path must be a relative bundle path")...)
	issues = append(issues, validateRequiredField("artefact.sha256", manifest.Artefact.SHA256, "manifest/artefact-sha256-required", "artefact sha256 is required")...)
	if manifest.Artefact.SHA256 != "" && !validSHA256(manifest.Artefact.SHA256) {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/artefact-sha256-invalid",
			Field:   "artefact.sha256",
			Message: "artefact sha256 must be a valid sha256 value",
		})
	}
	issues = append(issues, validateRequiredField("runtime.engine", manifest.Runtime.Engine, "manifest/runtime-engine-required", "runtime engine is required")...)
	issues = append(issues, validatePathField("runtime.config", manifest.Runtime.Config, "manifest/runtime-config-invalid", "runtime config must be a relative bundle path")...)
	if !manifest.Runtime.Acceleration.valid() {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/runtime-acceleration-invalid",
			Field:   "runtime.acceleration",
			Message: "runtime acceleration must be off, auto, or required",
		})
	}
	if !manifest.Runtime.Filter.valid() {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/runtime-filter-invalid",
			Field:   "runtime.filter",
			Message: "runtime filter must be none, nearest, bilinear, scanline, or crt",
		})
	}
	if engineRequiresProfile(manifest.Runtime.Engine) && manifest.Runtime.Profile == "" {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/runtime-profile-required",
			Field:   "runtime.profile",
			Message: "runtime profile is required for the selected engine",
		})
	}

	if manifest.Runtime.Entrypoint != "" {
		issues = append(issues, validatePathField("runtime.entrypoint", manifest.Runtime.Entrypoint, "manifest/runtime-entrypoint-invalid", "runtime entrypoint must be a relative bundle path")...)
	}

	issues = append(issues, validatePathField("verification.chain", manifest.Verification.Chain, "manifest/verification-chain-invalid", "verification chain must be a relative bundle path")...)
	issues = append(issues, validatePathField("verification.sbom", manifest.Verification.SBOM, "manifest/verification-sbom-invalid", "SBOM path must be a relative bundle path")...)

	if manifest.Save.Path != "" {
		issues = append(issues, validatePathField("save.path", manifest.Save.Path, "manifest/save-path-invalid", "save path must be a relative bundle path")...)
	}
	if manifest.Save.Screenshots != "" {
		issues = append(issues, validatePathField("save.screenshots", manifest.Save.Screenshots, "manifest/save-screenshots-invalid", "screenshot path must be a relative bundle path")...)
	}

	for _, permissionPath := range manifest.Permissions.FileSystem.Read {
		issues = append(issues, validatePathField("permissions.filesystem.read", permissionPath, "manifest/filesystem-read-invalid", "filesystem read path must be a relative bundle path")...)
	}
	for _, permissionPath := range manifest.Permissions.FileSystem.Write {
		issues = append(issues, validatePathField("permissions.filesystem.write", permissionPath, "manifest/filesystem-write-invalid", "filesystem write path must be a relative bundle path")...)
	}

	if manifest.Distribution.Mode == "" && manifest.Distribution.BYOROM {
		issues = append(issues, ValidationIssue{
			Code:    "manifest/distribution-mode-required",
			Field:   "distribution.mode",
			Message: "distribution mode is required when BYOROM is declared",
		})
	}

	return issues
}

func validateRequiredField(field string, value string, code string, message string) ValidationErrors {
	if value == "" {
		return ValidationErrors{
			{
				Code:    code,
				Field:   field,
				Message: message,
			},
		}
	}

	return nil
}

func validatePathField(field string, value string, code string, message string) ValidationErrors {
	if !validBundlePath(value) {
		return ValidationErrors{
			{
				Code:    code,
				Field:   field,
				Message: message,
			},
		}
	}

	return nil
}

func validBundlePath(value string) bool {
	if value == "" {
		return false
	}

	cleanPath := path.Clean(value)
	if cleanPath == "." {
		return false
	}

	return fs.ValidPath(cleanPath)
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}

	_, err := hex.DecodeString(value)
	return err == nil
}

func engineRequiresProfile(engine string) bool {
	switch engine {
	case "retroarch", "scummvm":
		return true
	default:
		return false
	}
}
