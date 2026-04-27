package play

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"
)

const (
	// CurrentManifestFormatVersion is the current STIM manifest schema version.
	CurrentManifestFormatVersion = "stim-v1"

	legacyManifestFormatVersion = "legacy"
)

// Manifest describes a runnable STIM bundle.
type Manifest struct {
	FormatVersion string       `yaml:"format_version,omitempty"`
	Name          string       `yaml:"name"`
	Version       string       `yaml:"version,omitempty"`
	Title         string       `yaml:"title"`
	Author        string       `yaml:"author,omitempty"`
	Year          int          `yaml:"year,omitempty"`
	Platform      string       `yaml:"platform"`
	Genre         string       `yaml:"genre,omitempty"`
	Licence       string       `yaml:"licence"`
	Artefact      Artefact     `yaml:"artefact"`
	Runtime       Runtime      `yaml:"runtime"`
	Preservation  Preservation `yaml:"preservation,omitempty"`
	Verification  Verification `yaml:"verification"`
	Permissions   Permissions  `yaml:"permissions"`
	Save          Save         `yaml:"save,omitempty"`
	Distribution  Distribution `yaml:"distribution,omitempty"`
}

// Artefact describes the preserved software payload.
type Artefact struct {
	Path      string `yaml:"path"`
	SHA256    string `yaml:"sha256"`
	Size      int64  `yaml:"size,omitempty"`
	MediaType string `yaml:"media_type,omitempty"`
	Source    string `yaml:"source,omitempty"`
}

// Runtime describes how a bundle should be launched.
type Runtime struct {
	Engine       string           `yaml:"engine"`
	Profile      string           `yaml:"profile,omitempty"`
	Config       string           `yaml:"config"`
	Entrypoint   string           `yaml:"entrypoint,omitempty"`
	Acceleration AccelerationMode `yaml:"acceleration,omitempty"`
	Filter       FrameFilter      `yaml:"filter,omitempty"`
}

// Preservation describes the RFC §2.1 hash-chain status.
type Preservation struct {
	Verified bool   `yaml:"verified"`
	Chain    string `yaml:"chain"`
}

// Verification describes integrity artefacts for a bundle.
type Verification struct {
	Chain         string  `yaml:"chain"`
	SBOM          string  `yaml:"sbom"`
	Deterministic bool    `yaml:"deterministic"`
	Engine        CodePin `yaml:"engine,omitempty"`
}

// CodePin records the runtime engine integrity value captured at bundle time.
type CodePin struct {
	Name   string `yaml:"name,omitempty" json:"name,omitempty"`
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	SHA256 string `yaml:"sha256,omitempty" json:"sha256,omitempty"`
}

// Permissions describes runtime capabilities requested by a bundle.
type Permissions struct {
	Network    bool                  `yaml:"network"`
	Microphone bool                  `yaml:"microphone"`
	FileSystem FileSystemPermissions `yaml:"filesystem"`
}

// FileSystemPermissions declares read and write access inside the bundle runtime.
type FileSystemPermissions struct {
	Read  []string `yaml:"read,omitempty"`
	Write []string `yaml:"write,omitempty"`
}

// Save describes save-state and screenshot directories.
type Save struct {
	Path        string `yaml:"path,omitempty"`
	Screenshots string `yaml:"screenshots,omitempty"`
}

// Distribution describes how a bundle is intended to be delivered.
type Distribution struct {
	Mode   string `yaml:"mode,omitempty"`
	BYOROM bool   `yaml:"byorom,omitempty"`
}

// LoadManifest parses a bundle manifest from YAML.
func LoadManifest(data []byte) (Manifest, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, err
	}
	var err error
	manifest, _, err = MigrateManifest(manifest)
	if err != nil {
		return Manifest{}, err
	}

	var trailing yaml.Node
	err = decoder.Decode(&trailing)
	if err == io.EOF {
		return manifest, nil
	}
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{}, ParseError{
		Kind:    "manifest/multiple-documents",
		Message: "manifest must contain exactly one YAML document",
	}
}

// ManifestMigration describes compatibility changes applied while loading a manifest.
type ManifestMigration struct {
	From    string
	To      string
	Applied []string
}

// Migrated reports whether any compatibility changes were applied.
func (migration ManifestMigration) Migrated() bool {
	return len(migration.Applied) > 0
}

// MigrateManifest upgrades a decoded manifest to the current STIM format version.
func MigrateManifest(manifest Manifest) (Manifest, ManifestMigration, error) {
	from := manifest.FormatVersion
	if from == "" {
		from = legacyManifestFormatVersion
	}

	migration := ManifestMigration{
		From: from,
		To:   CurrentManifestFormatVersion,
	}

	switch from {
	case legacyManifestFormatVersion:
		manifest.FormatVersion = CurrentManifestFormatVersion
		migration.Applied = append(migration.Applied, "manifest/legacy-format-version")
	case CurrentManifestFormatVersion:
		manifest.FormatVersion = CurrentManifestFormatVersion
	default:
		return Manifest{}, migration, ParseError{
			Kind:    "manifest/format-version-unsupported",
			Message: "unsupported STIM manifest format version",
		}
	}

	manifest = normaliseManifest(manifest, &migration)

	return manifest, migration, nil
}

func normaliseManifest(manifest Manifest, migration *ManifestMigration) Manifest {
	if manifest.Preservation.Chain == "" && manifest.Verification.Chain != "" {
		manifest.Preservation.Chain = manifest.Verification.Chain
		recordManifestMigration(migration, "manifest/preservation-chain")
	}
	if manifest.Verification.Chain == "" && manifest.Preservation.Chain != "" {
		manifest.Verification.Chain = manifest.Preservation.Chain
		recordManifestMigration(migration, "manifest/verification-chain")
	}
	if manifest.Verification.SBOM == "" && manifest.Preservation.Chain != "" {
		manifest.Verification.SBOM = "sbom.json"
		recordManifestMigration(migration, "manifest/default-sbom")
	}

	return manifest
}

func recordManifestMigration(migration *ManifestMigration, change string) {
	if migration == nil {
		return
	}

	migration.Applied = append(migration.Applied, change)
}

// ParseError reports manifest parsing problems.
type ParseError struct {
	Kind    string
	Message string
}

func (parseError ParseError) Error() string {
	return parseError.Kind + ": " + parseError.Message
}
