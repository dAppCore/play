package play

import (
	"bytes"
	"strconv"

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

	sbomData := renderSBOM(plan.Manifest)
	checksumData := renderChecksumChain(plan.Manifest, manifestData, runtimeConfigData, sbomData)

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

func renderSBOM(manifest Manifest) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("{\n")
	buffer.WriteString("  \"bomFormat\": \"CycloneDX\",\n")
	buffer.WriteString("  \"specVersion\": \"1.5\",\n")
	buffer.WriteString("  \"version\": 1,\n")
	buffer.WriteString("  \"metadata\": {\n")
	buffer.WriteString("    \"component\": {\n")
	buffer.WriteString("      \"name\": ")
	buffer.WriteString(strconv.Quote(manifest.Name))
	buffer.WriteString(",\n")
	buffer.WriteString("      \"type\": \"application\"\n")
	buffer.WriteString("    }\n")
	buffer.WriteString("  }\n")
	buffer.WriteString("}\n")

	return buffer.Bytes()
}

func renderChecksumChain(manifest Manifest, manifestData []byte, runtimeConfigData []byte, sbomData []byte) []byte {
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

	return buffer.Bytes()
}
