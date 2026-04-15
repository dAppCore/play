package play

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"
)

// Manifest describes a runnable STIM bundle.
type Manifest struct {
	Name         string       `yaml:"name"`
	Title        string       `yaml:"title"`
	Author       string       `yaml:"author,omitempty"`
	Year         int          `yaml:"year,omitempty"`
	Platform     string       `yaml:"platform"`
	Genre        string       `yaml:"genre,omitempty"`
	Licence      string       `yaml:"licence"`
	Artefact     Artefact     `yaml:"artefact"`
	Runtime      Runtime      `yaml:"runtime"`
	Verification Verification `yaml:"verification"`
	Permissions  Permissions  `yaml:"permissions"`
	Save         Save         `yaml:"save,omitempty"`
	Distribution Distribution `yaml:"distribution,omitempty"`
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

// Verification describes integrity artefacts for a bundle.
type Verification struct {
	Chain         string `yaml:"chain"`
	SBOM          string `yaml:"sbom"`
	Deterministic bool   `yaml:"deterministic"`
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

	var trailing yaml.Node
	err := decoder.Decode(&trailing)
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

// ParseError reports manifest parsing problems.
type ParseError struct {
	Kind    string
	Message string
}

func (parseError ParseError) Error() string {
	return parseError.Kind + ": " + parseError.Message
}
