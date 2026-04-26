package play

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// RenderedBundle contains deterministic file content for a planned bundle.
type RenderedBundle struct {
	Path  string
	Files []RenderedFile
}

// RenderedFile is a generated bundle file ready to be written by another layer.
type RenderedFile struct {
	Path string
	Data []byte
}

type runtimeConfigFile struct {
	Engine  string                `yaml:"engine"`
	Profile string                `yaml:"profile,omitempty"`
	Display *runtimeDisplayConfig `yaml:"display,omitempty"`
}

type runtimeDisplayConfig struct {
	Acceleration AccelerationMode `yaml:"acceleration,omitempty"`
	Filter       FrameFilter      `yaml:"filter,omitempty"`
}

// Render generates deterministic file content for a planned bundle.
func (plan BundlePlan) Render() (RenderedBundle, error) {
	issues := plan.Manifest.Validate()
	if issues.HasIssues() {
		return RenderedBundle{}, issues
	}

	manifestData, err := renderManifest(plan.Manifest)
	if err != nil {
		return RenderedBundle{}, err
	}

	runtimeConfigData, err := renderRuntimeConfig(plan.Manifest)
	if err != nil {
		return RenderedBundle{}, err
	}

	sbomData, err := BuildSBOM(plan.Manifest)
	if err != nil {
		return RenderedBundle{}, err
	}
	checksumData := renderChecksumChain(plan.Manifest, manifestData, runtimeConfigData, sbomData, plan.EngineBinaryData)

	files := []RenderedFile{
		{
			Path: "manifest.yaml",
			Data: manifestData,
		},
		{
			Path: plan.Manifest.Runtime.Config,
			Data: runtimeConfigData,
		},
		{
			Path: plan.Manifest.Verification.SBOM,
			Data: sbomData,
		},
		{
			Path: plan.Manifest.Verification.Chain,
			Data: checksumData,
		},
	}
	if len(plan.ArtefactData) > 0 {
		files = append(files, RenderedFile{
			Path: plan.Manifest.Artefact.Path,
			Data: cloneBytes(plan.ArtefactData),
		})
	}
	if len(plan.EngineBinaryData) > 0 && plan.Manifest.Verification.Engine.Path != "" {
		files = append(files, RenderedFile{
			Path: plan.Manifest.Verification.Engine.Path,
			Data: cloneBytes(plan.EngineBinaryData),
		})
	}

	return RenderedBundle{
		Path:  plan.Path,
		Files: files,
	}, nil
}

func renderManifest(manifest Manifest) ([]byte, error) {
	return yaml.Marshal(manifest)
}

func renderRuntimeConfig(manifest Manifest) ([]byte, error) {
	runtimeConfig := runtimeConfigFile{
		Engine:  manifest.Runtime.Engine,
		Profile: manifest.Runtime.Profile,
	}
	if manifest.Runtime.Acceleration != "" || manifest.Runtime.Filter != "" {
		runtimeConfig.Display = &runtimeDisplayConfig{
			Acceleration: manifest.Runtime.Acceleration,
			Filter:       manifest.Runtime.Filter,
		}
	}

	return yaml.Marshal(runtimeConfig)
}

func renderChecksumChain(manifest Manifest, manifestData []byte, runtimeConfigData []byte, sbomData []byte, engineBinaryData []byte) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(sha256Hex(manifestData))
	buffer.WriteString("  manifest.yaml\n")
	buffer.WriteString(sha256Hex(runtimeConfigData))
	buffer.WriteString("  ")
	buffer.WriteString(manifest.Runtime.Config)
	buffer.WriteString("\n")
	buffer.WriteString(sha256Hex(sbomData))
	buffer.WriteString("  ")
	buffer.WriteString(manifest.Verification.SBOM)
	buffer.WriteString("\n")
	buffer.WriteString(manifest.Artefact.SHA256)
	buffer.WriteString("  ")
	buffer.WriteString(manifest.Artefact.Path)
	buffer.WriteString("\n")
	if len(engineBinaryData) > 0 && manifest.Verification.Engine.Path != "" {
		buffer.WriteString(sha256Hex(engineBinaryData))
		buffer.WriteString("  ")
		buffer.WriteString(manifest.Verification.Engine.Path)
		buffer.WriteString("\n")
	}

	return buffer.Bytes()
}
